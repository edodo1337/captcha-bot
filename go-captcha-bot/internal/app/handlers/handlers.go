package handlers

import (
	"captcha-bot/internal/app/keyboards"
	"captcha-bot/internal/app/logic"
	"context"
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

		stateData := captchaService.InitCaptcha(ctx, member, chat)
		keyboard := keyboards.SliderCaptchaKeyboard(captchaService)

		msg := logic.CaptchaMessage(
			logic.CaptchaLength,
			stateData.CurrentPos,
			stateData.CorrectPos,
			captchaService.Config.Bot.CaptchaMsg,
			captchaService.Config.Bot.BanTimeout,
		)

		c.Reply(msg, &keyboard)

		return nil
	}
}
