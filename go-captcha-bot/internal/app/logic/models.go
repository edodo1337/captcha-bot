package logic

import (
	"encoding/json"
	"sync"
	"time"
)

type UserData struct {
	CurrentPos      int
	CorrectPos      int
	State           UserState
	Expiration      int64
	UserID          int64
	ChatID          int64
	CaptchaMessages []MessageStub
	Messages        []MessageStub
	JoinedAt        time.Time
}

func (item *UserData) Expired() bool {
	if item.Expiration == 0 {
		return false
	}

	return time.Now().UnixNano() > item.Expiration
}

type PollData struct {
	UsersVotedMap sync.Map
	VotesFor      uint
	VotesAgainst  uint
	Expiration    int64
	AuthorUserID  int64
	UserToKickID  int64
	PollMsgID     int64
}

func (item *PollData) Expired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

type MessageStub struct {
	ChatID    int64
	MessageID string
}

func (m MessageStub) MessageSig() (messageID string, chatID int64) {
	return m.MessageID, int64(m.ChatID)
}

func UnmarshalUserData(data []byte) (*UserData, error) {
	var userData UserData
	err := json.Unmarshal(data, &userData)
	if err != nil {
		return nil, err
	}
	return &userData, nil
}

func MarshalUserData(userData *UserData) ([]byte, error) {
	data, err := json.Marshal(userData)
	if err != nil {
		return nil, err
	}
	return data, nil
}
