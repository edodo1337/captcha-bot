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
		user, err := c.Bot().ChatMemberOf(chat, c.Message().Sender)
		log.Printf(
			"User joined name=%s %s, username=%s, user_id=%d\n",
			user.User.FirstName,
			user.User.LastName,
			user.User.Username,
			user.User.ID,
		)

		if err != nil {
			log.Println("Show captcha error", err)
			return err
		}

		stateData := captchaService.InitCaptcha(ctx, user, chat)
		keyboard := keyboards.SliderCaptchaKeyboard(captchaService)

		msg := logic.CaptchaMessage(logic.CaptchaLength, stateData.CurrentPos, stateData.CorrectPos)

		c.Reply(msg, &keyboard)

		return nil
	}
}
