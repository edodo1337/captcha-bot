package logic

import (
	"captcha-bot/internal/pkg/conf"
	"context"
	"errors"
	"log"
	"math/rand"
	"time"

	tele "gopkg.in/telebot.v3"
)

type UserRepository interface {
	GetByUserID(userID int64) (*UserData, error)
	Put(userData *UserData) error
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

	userData := UserData{CurrentPos: currentPos, CorrectPos: correctPos, State: Check, UserID: member.User.ID}

	if err := service.Storage.Put(&userData); err != nil {
		log.Println("Can't set state data", err)
	} else {
		log.Println("State set")
	}
	go service.banCountdown(ctx, member, chat, time.Duration(service.Config.Bot.BanTimeout))

	return &userData, nil
}

func (service *CaptchaService) ProcessButton(user *tele.ChatMember, chat *tele.Chat, button ButtonEvent) (*UserData, error) {
	userId := user.User.ID
	data, err := service.Storage.GetByUserID(userId)
	if err != nil {
		if !errors.Is(err, ErrStateNotFound) {
			log.Printf("Process button error %s\n", err)
		}
		return nil, err
	}

	data.CurrentPos = (data.CurrentPos + CaptchaLength + int(button)) % CaptchaLength
	if data.CurrentPos < 0 {
		data.CurrentPos = -data.CurrentPos
	}

	if data.CurrentPos == data.CorrectPos {
		data.State = Approved
		if err := service.Bot.Promote(chat, user); err != nil {
			log.Printf("Promote error: %s", err)
		}
		log.Printf("Correct answer, promote user %d", user.User.ID)
	}

	service.Storage.Put(data)

	return data, nil
}

func (service *CaptchaService) banCountdown(ctx context.Context, user *tele.ChatMember, chat *tele.Chat, timeout time.Duration) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(timeout * time.Second):
		userId := user.User.ID
		userData, err := service.Storage.GetByUserID(userId)
		if err != nil {
			log.Printf("Get state error: %s\n", err)
			return
		}
		if userData.State == Check {
			log.Printf("Ban user %d", user.User.ID)
			userData.State = Ban
			service.Storage.Put(userData)
			service.Bot.Ban(chat, user)
		} else {
			userData.State = Approved
			service.Storage.Put(userData)
		}
	}
}
