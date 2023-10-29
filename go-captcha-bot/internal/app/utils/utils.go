package utils

import (
	"captcha-bot/internal/app/logic"
	"captcha-bot/internal/pkg/conf"
	f "captcha-bot/internal/pkg/fsm"
	"context"

	h "captcha-bot/internal/app/handlers"

	tele "gopkg.in/telebot.v3"
)

func RunPolling(ctx context.Context, config *conf.Config) error {
	fsm := f.NewInMemoryFSM()
	captchaService, err := logic.NewCaptchaService(fsm, config)
	if err != nil {
		return err
	}
	registerHandlers(ctx, captchaService)
	captchaService.Run()

	return nil
}

func registerHandlers(ctx context.Context, captchaService *logic.CaptchaService) {
	captchaService.Bot.Handle(tele.OnUserJoined, h.ShowCaptcha(ctx, captchaService))
}
