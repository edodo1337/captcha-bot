package logic

import (
	"captcha-bot/internal/pkg/conf"
	"context"
	"errors"
	"log"
	"sync"
	"time"

	tele "gopkg.in/telebot.v3"
)

type PollRepository interface {
	GetByUserID(userID int64) (*PollData, error)
	Put(pollData *PollData) error
}

type PollService struct {
	Storage PollRepository
	Config  *conf.Config
	Bot     *tele.Bot
}

func NewPollService(bot *tele.Bot, st PollRepository, config *conf.Config) (*PollService, error) {
	return &PollService{
		Storage: st,
		Config:  config,
		Bot:     bot,
	}, nil
}

func (service *PollService) InitVoteKick(
	ctx context.Context,
	author, memberToKick *tele.ChatMember,
	chat *tele.Chat,
	pollMsg *tele.Message,
) error {
	pollData := PollData{
		AuthorUserID:  author.User.ID,
		UserToKickID:  memberToKick.User.ID,
		UsersVotedMap: sync.Map{},
	}

	if err := service.Storage.Put(&pollData); err != nil {
		log.Println("Can't put poll data:", err)
		return err
	} else {
		log.Println("Poll data put")
	}

	go service.pollCountdown(
		ctx,
		memberToKick,
		pollMsg,
		chat,
		time.Duration(service.Config.Bot.VoteKickTimeout)*time.Second,
		service.Config.Bot.MinKickVotesFor,
	)

	return nil
}

func (service *PollService) PollIsAlreadyExist(
	ctx context.Context,
	memberToKick *tele.ChatMember,
	chat *tele.Chat,
) (bool, error) {
	existPoll, err := service.Storage.GetByUserID(memberToKick.User.ID)
	if err != nil && !errors.Is(err, ErrPollNotFound) {
		return false, err
	}

	if existPoll != nil && !existPoll.Expired() {
		return true, nil
	}

	return false, nil
}

func (service *PollService) pollCountdown(
	ctx context.Context,
	memberToKick *tele.ChatMember,
	pollMsg *tele.Message,
	chat *tele.Chat,
	timeout time.Duration,
	minVotes uint,
) {
	select {
	case <-ctx.Done():
		log.Println("Stop votekick countdown due to ctx signal")
		return
	case <-time.After(timeout):
		log.Printf("Check votekick status for user=%d", memberToKick.User.ID)
		memberToKickUserID := memberToKick.User.ID
		pollData, err := service.Storage.GetByUserID(memberToKickUserID)
		if err != nil {
			log.Printf("Get poll data error: %s\n", err)
			return
		}

		if pollData.Expired() {
			userToKick := memberToKick.User

			if pollData.VotesFor < minVotes {
				msg := VoteKickMsgFailed(userToKick, MinVotesThreesold, minVotes)
				service.Bot.Edit(pollMsg, msg, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
				return
			}

			if pollData.VotesFor <= pollData.VotesAgainst {
				msg := VoteKickMsgFailed(userToKick, NotEnoughVotes, minVotes)
				service.Bot.Edit(pollMsg, msg, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
				return
			}

			msg := VoteKickMsgSucess(userToKick)
			service.Bot.Edit(pollMsg, msg, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
			service.Bot.Ban(chat, memberToKick)
		}
	}
}

func (service *PollService) ProcessButton(user, memberToKick *tele.ChatMember, chat *tele.Chat, button ButtonEvent) (*PollData, error) {
	userID := user.User.ID
	memberToKickID := memberToKick.User.ID
	data, err := service.Storage.GetByUserID(memberToKickID)
	if err != nil {
		if !errors.Is(err, ErrStateNotFound) {
			log.Printf("Process poll button error %s\n", err)
		}
		return nil, err
	}

	userVotedOption, ok := data.UsersVotedMap.Load(userID)
	if !ok {
		switch button {
		case Left: // +
			data.VotesFor++
		case Right: // -
			data.VotesAgainst++
		}
		data.UsersVotedMap.Store(userID, button)
	} else {
		if userVotedOption == button {
			switch button {
			case Left:
				data.VotesFor--
			case Right:
				data.VotesAgainst--
			}
			data.UsersVotedMap.Delete(userID)
		} else {
			switch button {
			case Left:
				data.VotesFor++
				data.VotesAgainst--
			case Right:
				data.VotesAgainst++
				data.VotesFor--
			}
			data.UsersVotedMap.Store(userID, button)
		}
	}

	service.Storage.Put(data)

	return data, nil
}
