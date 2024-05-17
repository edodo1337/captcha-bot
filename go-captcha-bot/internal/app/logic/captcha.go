package logic

import (
	"captcha-bot/internal/pkg/conf"
	"context"
	"errors"
	"log"
	"math/rand"
	"strconv"
	"time"

	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

type UserRepository interface {
	GetUserData(userID int64, chatID int64) (*UserData, error)
	Put(userData *UserData) error
	Remove(userID int64, chatID int64)
}

type CaptchaService struct {
	Storage UserRepository
	Config  *conf.Config
	Bot     *tele.Bot
}

func NewCaptchaService(bot *tele.Bot, st UserRepository, config *conf.Config) (*CaptchaService, error) {
	return &CaptchaService{
		Storage: st,
		Config:  config,
		Bot:     bot,
	}, nil
}

func (service *CaptchaService) Run() {
	service.Bot.Start()
	service.Bot.Use(middleware.Logger())
	service.Bot.Use(middleware.AutoRespond())
}

func (service *CaptchaService) InitCaptcha(ctx context.Context, member *tele.ChatMember, chat *tele.Chat) (*UserData, error) {
	log.Printf("Restrict user %d\n", member.User.ID)
	service.Bot.Restrict(chat, member)

	currentPos := rand.Intn(CaptchaLength)
	correctPos := rand.Intn(CaptchaLength)

	for correctPos == currentPos {
		randOffset := 2 + rand.Intn(5)
		correctPos = (correctPos + randOffset + CaptchaLength + 1) % CaptchaLength
	}

	userData := UserData{
		State:      Check,
		CurrentPos: currentPos, CorrectPos: correctPos,
		UserID: member.User.ID, ChatID: chat.ID,
	}

	if err := service.Storage.Put(&userData); err != nil {
		log.Println("Can't set state data", err)
	} else {
		log.Println("State set")
	}
	go service.banCountdown(ctx, member, chat, time.Duration(service.Config.Bot.BanTimeout))

	return &userData, nil
}

func (service *CaptchaService) ProcessButton(member *tele.ChatMember, chat *tele.Chat, button ButtonEvent) (*UserData, error) {
	userID := member.User.ID
	data, err := service.Storage.GetUserData(userID, chat.ID)
	if err != nil {
		if !errors.Is(err, ErrStateNotFound) {
			log.Printf("Process button error userID: %d, chatID: %d %s\n", data.UserID, data.ChatID, err)
		}
		return nil, err
	}
	if data.State != Check {
		return nil, errors.New("user has no active captcha status")
	}

	data.CurrentPos = (data.CurrentPos + CaptchaLength + int(button)) % CaptchaLength
	if data.CurrentPos < 0 {
		data.CurrentPos = -data.CurrentPos
	}

	if data.CurrentPos == data.CorrectPos {
		data.State = Approved

		permissions := tele.Rights{
			CanPostMessages:   true,
			CanEditMessages:   true,
			CanDeleteMessages: true,
			CanInviteUsers:    true,
			CanSendMessages:   true,
			CanSendMedia:      true,
			CanSendPolls:      true,
			CanSendOther:      true,
			CanAddPreviews:    true,
		}

		unrestrictedMember := &tele.ChatMember{
			User:   member.User,
			Rights: permissions,
		}

		if err := service.Bot.Restrict(chat, unrestrictedMember); err != nil {
			log.Printf("Promote error: %s", err)
			return nil, err
		}
		log.Printf("Correct answer, promote user %d", member.User.ID)

		if err := service.FlushCaptcha(member.User, chat, data.CaptchaMessages); err != nil {
			log.Println("Flush captcha error", err)
		}
	}

	service.Storage.Put(data)

	return data, nil
}

func (service *CaptchaService) banCountdown(ctx context.Context, member *tele.ChatMember, chat *tele.Chat, timeout time.Duration) {
	log.Printf("Run ban countdown for userID: %d\n", member.User.ID)
	select {
	case <-ctx.Done():
		log.Println("Shutdown countdown due to ctx signal")
		return
	case <-time.After(timeout * time.Second):
		log.Printf("Captcha timeout for user=%d", member.User.ID)
		userID := member.User.ID
		userData, err := service.Storage.GetUserData(userID, chat.ID)
		if err != nil {
			log.Printf("Get state error: %s\n", err)
			return
		}
		if userData.State == Check {
			log.Printf("Ban user %d", member.User.ID)
			service.Bot.Ban(chat, member, true)
		}
		if err := service.FlushCaptcha(member.User, chat, userData.CaptchaMessages); err != nil {
			log.Println("Flush captcha error", err)
		}
		service.Storage.Remove(userID, chat.ID)
	}
}

func (service *CaptchaService) SaveMessages(user *tele.User, chat *tele.Chat, messages []*tele.Message) (*UserData, error) {
	data, err := service.Storage.GetUserData(user.ID, chat.ID)
	if err != nil {
		if !errors.Is(err, ErrStateNotFound) {
			log.Printf("SaveMessages error userID: %d, chatID: %d %s\n", data.UserID, data.ChatID, err)
		}
		return nil, err
	}

	messageStubs := make([]MessageStub, 2)
	for _, m := range messages {
		messageStubs = append(messageStubs, MessageStub{chatID: m.Chat.ID, messageID: strconv.Itoa(m.ID)})
	}

	data.CaptchaMessages = append(data.CaptchaMessages, messageStubs...)
	if err := service.Storage.Put(data); err != nil {
		log.Println("Can't save user data", err)
		return nil, err
	}
	return data, nil
}

func (service *CaptchaService) FlushCaptcha(user *tele.User, chat *tele.Chat, messages []MessageStub) error {
	log.Println("Clean captcha")
	data, err := service.Storage.GetUserData(user.ID, chat.ID)
	if err != nil {
		if !errors.Is(err, ErrStateNotFound) {
			log.Printf("FlushCaptcha error userID: %d, chatID: %d %s\n", data.UserID, data.ChatID, err)
		}
		return err
	}

	for _, msg := range data.CaptchaMessages {
		if err := service.Bot.Delete(msg); err != nil {
			log.Println("Couldn't delete message")
		}
	}

	return nil
}
