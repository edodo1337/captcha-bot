package utils

import (
	"captcha-bot/internal/app/logic"
	"captcha-bot/internal/pkg/conf"
	"captcha-bot/internal/pkg/storage"
	"context"
	"log"
	"time"

	h "captcha-bot/internal/app/handlers"

	tele "gopkg.in/telebot.v3"
)

func RunPolling(ctx context.Context, config *conf.Config) error {
	userRepo := storage.NewUserInMemoryRepo(config.Bot.UserStateTTL, config.Bot.CleanupInterval)
	pollRepo := storage.NewPollInMemoryRepo(config.Bot.VoteKickTimeout, config.Bot.CleanupInterval)

	pref := tele.Settings{
		Token:  config.BotToken(),
		Poller: &tele.LongPoller{Timeout: 1 * time.Second},
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return err
	}

	// clean up old updates
	startTime := time.Now().Unix()
	bot.Use(func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			if c.Message().Unixtime < startTime {
				return nil
			}
			return next(c)
		}
	})

	captchaService, err := logic.NewCaptchaService(bot, userRepo, config)
	if err != nil {
		return err
	}
	pollService, err := logic.NewPollService(bot, pollRepo, config)
	if err != nil {
		return err
	}

	registerHandlers(ctx, captchaService, pollService)
	captchaService.Run()

	return nil
}

func registerHandlers(ctx context.Context, captchaService *logic.CaptchaService, pollService *logic.PollService) {
	captchaService.Bot.Handle(tele.OnUserJoined, h.ShowCaptcha(ctx, captchaService))
	captchaService.Bot.Handle("/votekick", h.VoteKick(ctx, pollService))
	captchaService.Bot.Handle("/hello", func(c tele.Context) error {
		log.Println("Hello cmd")
		return c.Send("Привет!")
	})
}
