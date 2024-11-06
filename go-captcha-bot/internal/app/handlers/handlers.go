package handlers

import (
	"captcha-bot/internal/app/keyboards"
	"captcha-bot/internal/app/logic"
	"fmt"

	"context"
	"log"

	tele "gopkg.in/telebot.v3"
)

func ShowCaptchaJoined(ctx context.Context, captchaService *logic.CaptchaService) tele.HandlerFunc {
	return func(c tele.Context) error {
		chat := c.Chat()
		if c.Update().ChatMember == nil {
			return nil
		}

		userJoined := c.Update().ChatMember.NewChatMember.User
		oldRole := c.Update().ChatMember.OldChatMember.Role
		newRole := c.Update().ChatMember.NewChatMember.Role
		isMemberOld := c.Update().ChatMember.OldChatMember.Member
		isMemberNew := c.Update().ChatMember.NewChatMember.Member

		conditionLeft := (oldRole == tele.Left) || (oldRole == tele.Kicked) || (oldRole == tele.Restricted) && (!isMemberOld)
		conditionRight := (newRole == tele.Member) || (newRole == tele.Restricted) && isMemberNew

		if !(conditionLeft && conditionRight) {
			return nil
		}

		log.Printf(
			"New user joined username=%s, firstLastName=%s %s, userID=%d\n",
			userJoined.Username, userJoined.FirstName, userJoined.LastName, userJoined.ID,
		)

		member, err := c.Bot().ChatMemberOf(chat, userJoined)
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
		keyboard := keyboards.SliderCaptchaKeyboard(ctx, captchaService)

		msg := logic.CaptchaMessage(
			logic.CaptchaLength,
			stateData.CurrentPos,
			stateData.CorrectPos,
			captchaService.Config.Bot.CaptchaMsg,
			captchaService.Config.Bot.BanTimeout,
		)

		log.Println("Reply captcha message")
		msgHello := fmt.Sprintf("Добро пожаловать, [%s %s](tg://user?id=%d)!", userJoined.FirstName, userJoined.LastName, userJoined.ID)
		msg1, err := c.Bot().Send(c.Chat(), msgHello, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
		if err != nil {
			log.Println("Send hello msg err", err)
		}

		msg2, err := c.Bot().Send(c.Chat(), msg, &tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: &keyboard})
		if err != nil {
			log.Println("Send captcha err", err)
		}

		_, err = captchaService.SaveMessages(ctx, userJoined, c.Chat(), []*tele.Message{msg1, msg2})
		if err != nil {
			log.Println("Save messages err", err)
		}

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

func OnNewMessage(ctx context.Context, service *logic.SpamFilterService) tele.HandlerFunc {
	return func(c tele.Context) error {
		approved, err := service.CheckAlreadyApproved(ctx, c.Message())
		if err != nil {
			log.Println("Check if user approved error", err)
			return err
		}

		if approved {
			log.Printf("User %d already approved", c.Message().Sender.ID)
			return nil
		}

		log.Printf("User %d not approved, check if message is spam", c.Message().Sender.ID)

		isSpam, err := service.CheckMessage(ctx, c.Message().Text)
		if err != nil {
			log.Println("Check message error", err)
			return nil
		}

		if isSpam {
			log.Printf("User %d is spammer", c.Message().Sender.ID)
			msg := fmt.Sprintf("Пользователь [%s %s](tg://user?id=%d) забанен за спам", c.Message().Sender.FirstName, c.Message().Sender.LastName, c.Message().Sender.ID)
			service.Bot.Reply(c.Message(), msg, &tele.SendOptions{ReplyTo: c.Message()})
			if err := service.BanAndFlushMessages(ctx, c.Message().Sender, c.Chat()); err != nil {
				log.Println("Ban and flush messages error", err)
			}
		}

		log.Println("Message is not spam")
		return nil
	}
}

func TailLogs(ctx context.Context, adminService *logic.AdminService) tele.HandlerFunc {
	return func(c tele.Context) error {
		msg, err := adminService.TailLogs(ctx, *c.Message())
		if err != nil {
			c.Reply(err)
			return err
		} else {
			c.Reply(msg)
		}

		return nil
	}
}
