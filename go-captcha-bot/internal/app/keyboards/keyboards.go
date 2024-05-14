package keyboards

import (
	"captcha-bot/internal/app/logic"
	"fmt"
	"log"

	tele "gopkg.in/telebot.v3"
)

func SliderCaptchaKeyboard(captchaService *logic.CaptchaService) tele.ReplyMarkup {
	var keyboard = tele.ReplyMarkup{ResizeKeyboard: true}
	btnLeft := keyboard.Data("⬅️", "leftBtn")
	btnRight := keyboard.Data("➡️", "rightBtn")
	keyboard.Inline(
		keyboard.Row(btnLeft, btnRight),
	)

	handler := func(button logic.ButtonEvent) func(c tele.Context) error {
		return func(c tele.Context) error {
			chat := c.Chat()
			user, err := c.Bot().ChatMemberOf(chat, c.Update().Callback.Sender)
			if err != nil {
				log.Printf("Get chat member error: %s", err)
				return err
			}

			userData, err := captchaService.ProcessButton(user, chat, button)
			if err != nil {
				log.Printf("Process captcha button, userID: %d, error: %s", user.User.ID, err)
				return nil
			}

			if userData == nil {
				log.Printf("Couldn't process captcha button, userData is null for userID %d", user.User.ID)
				return nil
			}

			if userData.State == logic.Approved {
				c.Bot().Delete(c.Message())
				return nil
			}

			msg := logic.CaptchaMessage(
				logic.CaptchaLength,
				userData.CurrentPos,
				userData.CorrectPos,
				captchaService.Config.Bot.CaptchaMsg,
				captchaService.Config.Bot.BanTimeout,
			)

			c.Bot().Edit(c.Message(), msg, &tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: &keyboard})

			return nil
		}
	}

	captchaService.Bot.Handle(&btnLeft, handler(logic.Left))
	captchaService.Bot.Handle(&btnRight, handler(logic.Right))

	return keyboard
}

func VoteKickKeyboard(pollService *logic.PollService, memberToKick *tele.ChatMember) tele.ReplyMarkup {
	var keyboard = tele.ReplyMarkup{ResizeKeyboard: true}
	btnLeft := keyboard.Data("✅За: 0", "leftBtn")
	btnRight := keyboard.Data("❌Против: 0", "rightBtn")
	keyboard.Inline(
		keyboard.Row(btnLeft, btnRight),
	)

	handler := func(button logic.ButtonEvent) func(c tele.Context) error {
		return func(c tele.Context) error {
			chat := c.Chat()
			member, err := c.Bot().ChatMemberOf(chat, c.Update().Callback.Sender)
			if err != nil {
				log.Printf("Get chat member error: %s", err)
				return err
			}

			pollData, err := pollService.ProcessButton(member, memberToKick, chat, button)
			if err != nil {
				log.Printf("Process vote button error: %s", err)
				return nil
			}

			if pollData == nil {
				log.Printf("Couldn't process vote button, userData is null")
				return nil
			}

			var keyboard = tele.ReplyMarkup{ResizeKeyboard: true}
			btnLeft := keyboard.Data(fmt.Sprintf("✅За: %d", pollData.VotesFor), "leftBtn")
			btnRight := keyboard.Data(fmt.Sprintf("❌Против: %d", pollData.VotesAgainst), "rightBtn")
			keyboard.Inline(
				keyboard.Row(btnLeft, btnRight),
			)

			_, err = c.Bot().Edit(c.Message(), c.Message().Text, &tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: &keyboard})
			if err != nil {
				log.Printf("Process vote button error: %s", err)
				return err
			}

			return nil
		}
	}

	pollService.Bot.Handle(&btnLeft, handler(logic.Left))
	pollService.Bot.Handle(&btnRight, handler(logic.Right))

	return keyboard
}
