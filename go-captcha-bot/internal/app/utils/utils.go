package utils

import (
	"captcha-bot/internal/app/logic"
	"captcha-bot/internal/pkg/conf"
	"captcha-bot/internal/pkg/storage"
	"context"

	h "captcha-bot/internal/app/handlers"

	tele "gopkg.in/telebot.v3"
)

func RunPolling(ctx context.Context, config *conf.Config) error {
	st := storage.NewUserInMemoryStorage(config.Bot.UserStateTTL, config.Bot.CleanupInterval)
	captchaService, err := logic.NewCaptchaService(st, config)
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
