package keyboards

import (
	"captcha-bot/internal/app/logic"

	tele "gopkg.in/telebot.v3"
)

func SliderCaptchaKeyboard(captchaService *logic.CaptchaService) tele.ReplyMarkup {
	var keyboard = tele.ReplyMarkup{ResizeKeyboard: true}
	btnLeft := keyboard.Data("⬅️", "leftBtn")
	btnRight := keyboard.Data("➡️", "rightBtn")
	keyboard.Inline(
		keyboard.Row(btnLeft, btnRight),
	)

	captchaService.Bot.Handle(&btnLeft, func(c tele.Context) error {
		chat := c.Chat()
		user, err := c.Bot().ChatMemberOf(chat, c.Update().Callback.Sender)
		if err != nil {
			return err
		}

		userData := captchaService.ProcessButton(user, chat, logic.Left)
		if userData != nil {
			msg := logic.CaptchaMessage(logic.CaptchaLength, userData.CurrentPos, userData.CorrectPos)
			c.Bot().Edit(c.Message(), msg, c.Message().ReplyMarkup)
		}

		return nil
	})

	captchaService.Bot.Handle(&btnRight, func(c tele.Context) error {
		chat := c.Chat()
		user, err := c.Bot().ChatMemberOf(chat, c.Update().Callback.Sender)
		if err != nil {
			return err
		}
		userData := captchaService.ProcessButton(user, chat, logic.Right)
		if userData == nil {
			return nil
		}

		if userData.State == logic.Approved {
			c.Bot().Delete(c.Message())
			return nil
		}

		msg := logic.CaptchaMessage(logic.CaptchaLength, userData.CurrentPos, userData.CorrectPos)
		c.Bot().Edit(c.Message(), msg, c.Message().ReplyMarkup)

		return nil
	})

	return keyboard
}
