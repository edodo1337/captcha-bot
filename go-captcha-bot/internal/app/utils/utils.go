package utils

import (
	"captcha-bot/internal/app/logic"
	"captcha-bot/internal/pkg/conf"
	"captcha-bot/internal/pkg/storage"
	"context"
	"log"
	"strconv"
	"time"

	h "captcha-bot/internal/app/handlers"

	tele "gopkg.in/telebot.v3"
)

func RunPolling(ctx context.Context, config *conf.Config) error {
	userRepo := storage.NewUserInMemoryRepo(config.Bot.UserStateTTL, config.Bot.CleanupInterval)
	pollRepo := storage.NewPollInMemoryRepo(config.Bot.VoteKickTimeout, config.Bot.CleanupInterval)

	poller := &tele.LongPoller{Timeout: 1 * time.Second, AllowedUpdates: []string{
		"message",
		"edited_message",
		"channel_post",
		"edited_channel_post",
		"inline_query",
		"chosen_inline_result",
		"callback_query",
		"pre_checkout_query",
		"my_chat_member",
		"chat_member",
		"chat_join_request",
	}}
	pref := tele.Settings{
		Token:  config.BotToken(),
		Poller: poller,
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
			if c.Message() != nil {
				if c.Message().Unixtime < startTime {
					return nil
				}
			}
			if c.ChatMember() != nil {
				if c.ChatMember().Unixtime < startTime {
					return nil
				}
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
	adminService, err := logic.NewAdminService(bot, poller, config)
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
	captchaService.Bot.Handle(tele.OnAddedToGroup, h.ShowCaptchaJoined(ctx, captchaService))
	captchaService.Bot.Handle(tele.OnChatMember, h.ShowCaptchaJoined(ctx, captchaService))

	captchaService.Bot.Handle("/votekick", h.VoteKick(ctx, pollService))
	captchaService.Bot.Handle("/hello", func(c tele.Context) error {
		log.Println("Hello cmd")
		msg1 := "Hello, [user](tg://user?id=" + strconv.Itoa(int(c.Sender().ID)) + ")!"
		return c.Send(msg1, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
	})
	adminService.Bot.Handle("/logs", h.TailLogs(ctx, adminService))
}
