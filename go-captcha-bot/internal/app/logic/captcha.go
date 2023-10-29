package logic

import (
	"captcha-bot/internal/pkg/conf"
	"context"
	"errors"
	"log"
	"time"

	tele "gopkg.in/telebot.v3"
)

type UserData struct {
	MessageIDs []int
	CurrentPos int
	CorrectPos int
	State      UserState
	StartTime  time.Time
}

type IFSM interface {
	SetState(userId int64, state UserState) error
	GetState(userId int64) (UserState, error)
	SetData(userId int64, data *UserData) error
	GetData(userId int64) (*UserData, error)
}

type CaptchaService struct {
	FSM    IFSM
	Config *conf.Config
	Bot    *tele.Bot
}

func NewCaptchaService(fsm IFSM, config *conf.Config) (*CaptchaService, error) {
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
		FSM:    fsm,
		Config: config,
		Bot:    bot,
	}, nil
}

func (service *CaptchaService) Run() {
	service.Bot.Start()
}

func (service *CaptchaService) InitCaptcha(ctx context.Context, member *tele.ChatMember, chat *tele.Chat) UserData {
	service.Bot.Restrict(chat, member)
	stateData := UserData{CurrentPos: 3, CorrectPos: 8, State: Check, StartTime: time.Now()}
	if err := service.FSM.SetData(member.User.ID, &stateData); err != nil {
		log.Println("Can't set state data", err)
	}
	go service.banCountdown(ctx, member, chat, time.Duration(service.Config.Bot.BanTimeout))

	return stateData
}

func (service *CaptchaService) ProcessButton(user *tele.ChatMember, chat *tele.Chat, button ButtonEvent) *UserData {
	userId := user.User.ID
	data, err := service.FSM.GetData(userId)

	if err != nil {
		if !errors.Is(err, ErrStateNotFound) {
			log.Printf("Process button error %s\n", err)
		}
		return nil
	}

	if button == Left {
		data.CurrentPos = (data.CurrentPos + CaptchaLength - 1) % CaptchaLength
		if data.CurrentPos < 0 {
			data.CurrentPos = -data.CurrentPos
		}
	}

	if button == Right {
		data.CurrentPos = (data.CurrentPos + CaptchaLength + 1) % CaptchaLength
	}

	if data.CurrentPos == data.CorrectPos {
		data.State = Approved
		if err := service.Bot.Promote(chat, user); err != nil {
			log.Printf("Promote error: %s", err)
		}
	}

	service.FSM.SetData(userId, data)
	return data
}

func (service *CaptchaService) banCountdown(ctx context.Context, user *tele.ChatMember, chat *tele.Chat, timeout time.Duration) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(timeout * time.Second):
		userId := user.User.ID
		state, err := service.FSM.GetState(userId)
		if err != nil {
			log.Printf("Get state error: %s\n", err)
			return
		}
		if state == Check {
			log.Println("Ban user")
			service.FSM.SetState(userId, Ban)
			service.Bot.Ban(chat, user)
		} else {
			service.FSM.SetState(userId, Approved)
		}
	}
}
