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

type UserData struct {
	MessageIDs []int
	CurrentPos int
	CorrectPos int
	State      UserState
	Expiration int64
}

func (item UserData) Expired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

type IUserStorage interface {
	SetState(userId int64, state UserState) error
	GetState(userId int64) (UserState, error)
	SetData(userId int64, data *UserData) error
	GetData(userId int64) (*UserData, error)
}

type CaptchaService struct {
	Storage IUserStorage
	Config  *conf.Config
	Bot     *tele.Bot
}

func NewCaptchaService(st IUserStorage, config *conf.Config) (*CaptchaService, error) {
	pref := tele.Settings{
		Token:  config.BotToken(),
		Poller: &tele.LongPoller{Timeout: 1 * time.Second},
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// bot.Use(middleware.Logger())
	// b.Use(middleware.AutoRespond())

	return &CaptchaService{
		Storage: st,
		Config:  config,
		Bot:     bot,
	}, nil
}

func (service *CaptchaService) Run() {
	service.Bot.Start()
}

func (service *CaptchaService) InitCaptcha(ctx context.Context, member *tele.ChatMember, chat *tele.Chat) UserData {
	log.Printf("Restrict user %d\n", member.User.ID)
	service.Bot.Restrict(chat, member)

	currentPos := rand.Intn(CaptchaLength)
	correctPos := rand.Intn(CaptchaLength)

	for correctPos == currentPos {
		randOffset := 2 + rand.Intn(5)
		correctPos = (correctPos + randOffset + CaptchaLength + 1) % CaptchaLength
	}

	stateData := UserData{CurrentPos: currentPos, CorrectPos: correctPos, State: Check}
	if err := service.Storage.SetData(member.User.ID, &stateData); err != nil {
		log.Println("Can't set state data", err)
	} else {
		log.Println("State set")
	}
	go service.banCountdown(ctx, member, chat, time.Duration(service.Config.Bot.BanTimeout))

	return stateData
}

func (service *CaptchaService) ProcessButton(user *tele.ChatMember, chat *tele.Chat, button ButtonEvent) *UserData {
	userId := user.User.ID
	data, err := service.Storage.GetData(userId)

	if err != nil {
		if !errors.Is(err, ErrStateNotFound) {
			log.Printf("Process button error %s\n", err)
		}
		return nil
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

	service.Storage.SetData(userId, data)
	return data
}

func (service *CaptchaService) banCountdown(ctx context.Context, user *tele.ChatMember, chat *tele.Chat, timeout time.Duration) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(timeout * time.Second):
		userId := user.User.ID
		state, err := service.Storage.GetState(userId)
		if err != nil {
			log.Printf("Get state error: %s\n", err)
			return
		}
		if state == Check {
			log.Printf("Ban user %d", user.User.ID)
			service.Storage.SetState(userId, Ban)
			service.Bot.Ban(chat, user)
		} else {
			service.Storage.SetState(userId, Approved)
		}
	}
}
