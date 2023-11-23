package storage

import (
	"captcha-bot/internal/app/logic"
	"errors"
	"log"
	"sync"
	"time"
)

type PollInMemoryRepo struct {
	m    sync.Map
	ttl  time.Duration
	stop chan struct{}
}

func NewPollInMemoryRepo(voteKickTimeout, cleanupInterval time.Duration) *PollInMemoryRepo {
	st := &PollInMemoryRepo{
		m:    sync.Map{},
		ttl:  voteKickTimeout,
		stop: make(chan struct{}),
	}

	go func() {
		ticker := time.NewTicker(cleanupInterval * time.Second)
		for {
			select {
			case <-st.stop:
				return
			case <-ticker.C:
				st.removeExpired()
			}
		}
	}()

	return st
}
func (st *PollInMemoryRepo) removeExpired() {
	st.m.Range(func(key, value interface{}) bool {
		itm, ok := value.(*logic.UserData)
		if !ok {
			return true
		}
		if itm.Expired() {
			log.Printf("Remove expired poll data for: %d", key)
			st.m.Delete(key)
		}
		return true
	})
}

func (st *PollInMemoryRepo) GetByUserID(userID int64) (*logic.PollData, error) {
	pollDataRaw, ok := st.m.Load(userID)
	if !ok {
		return nil, logic.ErrPollNotFound
	}

	pollData, ok := pollDataRaw.(*logic.PollData)
	if !ok {
		return nil, errors.New("poll data decode error")
	}

	// if pollData.Expired() {
	// 	st.m.Delete(userID)
	// 	return nil, logic.ErrPollNotFound
	// }

	return pollData, nil
}

func (st *PollInMemoryRepo) Put(pollData *logic.PollData) error {
	if pollData.AuthorUserID == 0 {
		return errors.New("couldn't put poll with empty AuthorUserID")
	}
	if pollData.UserToKickID == 0 {
		return errors.New("couldn't put poll with empty UserToKickID")
	}

	existPoll, err := st.GetByUserID(pollData.UserToKickID)
	if err != nil {
		if !errors.Is(err, logic.ErrPollNotFound) {
			return err
		}

		expiration := int64(0)
		if st.ttl > 0 {
			expiration = time.Now().Add(st.ttl * time.Second).UnixNano()
		}
		pollData.Expiration = expiration
	} else {
		pollData.Expiration = existPoll.Expiration
	}

	st.m.Store(pollData.UserToKickID, pollData)

	return nil
}

func (st *PollInMemoryRepo) Remove(userID int64) {
	st.m.Delete(userID)
}
