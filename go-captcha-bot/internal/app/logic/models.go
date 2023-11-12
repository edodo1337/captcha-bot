package logic

import (
	"sync"
	"time"
)

type UserData struct {
	CurrentPos int
	CorrectPos int
	State      UserState
	Expiration int64
	UserID     int64
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

// func NewPollData(authorUserID, userToKickID, pollMsgID int64) (*PollData, error) {
// 	if authorUserID <= 0 {
// 		return nil, errors.New(fmt.Sprintf("incorrect authorUserID: %s recieved", authorUserID))
// 	}
// 	if userToKickID <= 0 {
// 		return nil, errors.New(fmt.Sprintf("incorrect userToKickID: %s recieved", userToKickID))
// 	}
// 	if pollMsgID <= 0 {
// 		return nil, errors.New(fmt.Sprintf("incorrect pollMsgID: %s recieved", pollMsgID))
// 	}

// 	return &PollData{
// 		authorUserID: authorUserID,
// 		userToKickID: userToKickID,
// 		userVotedMap: sync.Map{},
// 		pollMsgID:    pollMsgID,
// 	}, nil
// }
