package handlers

import (
	"captcha-bot/internal/app/keyboards"
	"captcha-bot/internal/app/logic"
	"context"
	"fmt"
	"log"

	tele "gopkg.in/telebot.v3"
)

func ShowCaptcha(ctx context.Context, captchaService *logic.CaptchaService) tele.HandlerFunc {
	return func(c tele.Context) error {
		chat := c.Chat()
		member, err := c.Bot().ChatMemberOf(chat, c.Message().Sender)
		log.Printf(
			"User joined name=%s %s, username=%s, user_id=%d\n",
			member.User.FirstName,
			member.User.LastName,
			member.User.Username,
			member.User.ID,
		)

		if err != nil {
			log.Println("ChatMemberOf error", err)
			return err
		}

		if member.Role == tele.Administrator || member.Role == tele.Creator {
			log.Println("User is admin. Skip.")
			return nil
		}

		stateData, err := captchaService.InitCaptcha(ctx, member, chat)
		if err != nil {
			log.Printf("Init captcha error: %s", err)
		}
		keyboard := keyboards.SliderCaptchaKeyboard(captchaService)

		msg := logic.CaptchaMessage(
			logic.CaptchaLength,
			stateData.CurrentPos,
			stateData.CorrectPos,
			captchaService.Config.Bot.CaptchaMsg,
			captchaService.Config.Bot.BanTimeout,
		)

		c.Reply(msg, &tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: &keyboard})

		return nil
	}
}

func VoteKick(ctx context.Context, pollService *logic.PollService) tele.HandlerFunc {
	return func(c tele.Context) error {
		if !c.Message().IsReply() {
			c.Reply(
				`Используйте команду /votekick в ответ `+
					`(реплай) на сообщение пользователя, которого хотите кикнуть`,
				&tele.SendOptions{ParseMode: tele.ModeMarkdown},
			)
			return nil
		}

		userToKick := c.Message().ReplyTo.Sender

		chat := c.Chat()
		pollAuthor, err := c.Bot().ChatMemberOf(chat, c.Message().Sender)
		if err != nil {
			return err
		}

		memberToKick, err := c.Bot().ChatMemberOf(chat, userToKick)
		if err != nil {
			return err
		}

		if memberToKick.Role == tele.Administrator || memberToKick.Role == tele.Creator {
			log.Println("User is admin. Skip votekick.")
			c.Reply("❌Нельзя кикать админов")
			return logic.ErrUserIsAdmin
		}

		exist, err := pollService.PollIsAlreadyExist(ctx, memberToKick, chat)
		if err != nil {
			log.Printf("Find poll error\n %s", err)
			return err
		}
		if exist {
			log.Printf("Poll already exist %s\n", err)
			c.Reply("❌ Голосование для выбранного пользователя уже идет")
			return nil
		}

		msg := logic.VoteKickMsg(userToKick)
		keyboard := keyboards.VoteKickKeyboard(pollService, memberToKick)

		pollMsg, err := pollService.Bot.Reply(
			c.Message(),
			msg,
			&tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: &keyboard},
		)
		if err != nil {
			return err
		}

		err = pollService.InitVoteKick(ctx, pollAuthor, memberToKick, chat, pollMsg)
		if err != nil {
			log.Printf("Votekick init error %s", err)
			return err
		}

		return nil
	}
}

func OnNewMessage(ctx context.Context, spamFilterService *logic.SpamFilterService) tele.HandlerFunc {
	return func(c tele.Context) error {
		result := spamFilterService.CheckMessage(ctx, c.Message().Text)
		fmt.Println(c.Message().Text, result)

		var response string
		if result {
			response = "Это спам"
		} else {
			response = "Это не спам"
		}
		c.Reply(response)

		return nil
	}
}
