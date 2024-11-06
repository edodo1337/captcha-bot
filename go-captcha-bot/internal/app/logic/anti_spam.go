package logic

import (
	"captcha-bot/internal/pkg/clients"
	"captcha-bot/internal/pkg/conf"
	"context"
	"errors"
	"log"
	"math/rand"
	"strconv"

	tele "gopkg.in/telebot.v3"
)

type SpamFilterClient interface {
	Shutdown() error
	IsSpam(ctx context.Context, text string) (bool, error)
}

type SpamFilterService struct {
	UserStorage UserRepository
	Bot         *tele.Bot
	Config      *conf.Config
}

func NewSpamFilterService(userRepo UserRepository, bot *tele.Bot, config *conf.Config) (*SpamFilterService, error) {
	log.Printf("Creating SpamFilterService with GPT client %s\n", config.Bot.GPTClient)
	return &SpamFilterService{
		UserStorage: userRepo,
		Bot:         bot,
		Config:      config,
	}, nil
}

func (sfs *SpamFilterService) GetSpamFilterClient(ctx context.Context) SpamFilterClient {
	l := len(sfs.Config.Bot.GeminiApiTokens)
	tokenPos := rand.Intn(l)

	switch sfs.Config.Bot.GPTClient {
	case "gemini":
		return clients.NewGeminiClient(ctx, sfs.Config.Bot.GeminiApiTokens[tokenPos], sfs.Config.Bot.GeminiModel, sfs.Config.Bot.PromptWrap)
	case "yandexgpt":
		return clients.NewYandexGPTClient(ctx, sfs.Config.Bot.PromptWrap, sfs.Config.Bot.YandexCatalogIDs[tokenPos], sfs.Config.Bot.YandexApiTokens[tokenPos])
	default:
		panic("Unknown GPT client")
	}
}

func (sfs *SpamFilterService) CheckAlreadyApproved(ctx context.Context, msg *tele.Message) (bool, error) {
	log.Printf("Checking if user %d already approved", msg.Sender.ID)
	userID := msg.Sender.ID
	chat := msg.Chat
	data, err := sfs.UserStorage.GetUserData(ctx, userID, chat.ID)
	if err != nil {
		if !errors.Is(err, ErrStateNotFound) {
			log.Printf("Can't get user data: %s", err)
		}
		return false, err
	}
	if data.State == Approved {
		return true, nil
	}

	data.Messages = append(data.Messages, MessageStub{ChatID: chat.ID, MessageID: strconv.Itoa(msg.ID)})
	if err := sfs.UserStorage.Put(ctx, data); err != nil {
		log.Println("Can't save user data", err)
		return false, err
	}

	passThreshold := len(data.Messages) > sfs.Config.Bot.MsgCountThreshold
	log.Printf("User has %d messages, threshold is %d", len(data.Messages), sfs.Config.Bot.MsgCountThreshold)

	return passThreshold, nil
}

func (sfs *SpamFilterService) CheckMessage(ctx context.Context, text string) (bool, error) {
	spamFilterClient := sfs.GetSpamFilterClient(ctx)
	isSpam, err := spamFilterClient.IsSpam(ctx, text)
	if err != nil {
		return false, err
	}
	if err := spamFilterClient.Shutdown(); err != nil {
		log.Println("Can't shutdown spam filter client", err)
		return false, err
	}

	return isSpam, nil
}

func (sfs *SpamFilterService) BanAndFlushMessages(ctx context.Context, user *tele.User, chat *tele.Chat) error {
	log.Println("AntiSpam Ban and clean messages")

	member, err := sfs.Bot.ChatMemberOf(chat, user)
	if err != nil {
		log.Println("Can't get chat member", err)
		return err
	}

	if err := sfs.Bot.Ban(chat, member, true); err != nil {
		log.Println("Can't ban user", err)
		return err
	}

	data, err := sfs.UserStorage.GetUserData(ctx, user.ID, chat.ID)
	if err != nil {
		if !errors.Is(err, ErrStateNotFound) {
			log.Printf("FlushMessages error userID: %d, chatID: %d %s\n", data.UserID, data.ChatID, err)
		}
		return err
	}

	for _, msg := range data.Messages {
		if err := sfs.Bot.Delete(msg); err != nil {
			log.Println("Delete message err", err)
		}
	}

	return nil
}
