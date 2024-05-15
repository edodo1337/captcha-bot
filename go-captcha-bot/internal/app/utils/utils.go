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
		Token: config.BotToken(),
		Poller: &tele.LongPoller{Timeout: 1 * time.Second, AllowedUpdates: []string{
			"update_id",
			"message",
			"edited_message",
			"channel_post",
			"edited_channel_post",
			"business_connection",
			"business_message",
			"edited_business_message",
			"deleted_business_messages",
			"message_reaction",
			"message_reaction_count",
			"inline_query",
			"chosen_inline_result",
			"callback_query",
			"shipping_query",
			"pre_checkout_query",
			"poll",
			"poll_answer",
			"my_chat_member",
			"chat_member",
			"chat_join_request",
			"chat_boost",
			"removed_chat_boost",
		}},
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return err
	}

	// bot.Use(middleware.Logger())
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
	adminService, err := logic.NewAdminService(bot, config)
	if err != nil {
		return err
	}

	registerHandlers(ctx, captchaService, pollService, adminService)
	captchaService.Run()

	return nil
}

func registerHandlers(
	ctx context.Context,
	captchaService *logic.CaptchaService,
	pollService *logic.PollService,
	adminService *logic.AdminService,
) {
	captchaService.Bot.Handle(tele.OnUserJoined, h.ShowCaptcha(ctx, captchaService))
	captchaService.Bot.Handle(tele.OnAddedToGroup, h.ShowCaptcha(ctx, captchaService))

	captchaService.Bot.Handle("/votekick", h.VoteKick(ctx, pollService))
	captchaService.Bot.Handle("/hello", func(c tele.Context) error {
		log.Println("Hello cmd")
		return c.Send("Привет!")
	})
	adminService.Bot.Handle("/logs", h.TailLogs(ctx, adminService))
}
