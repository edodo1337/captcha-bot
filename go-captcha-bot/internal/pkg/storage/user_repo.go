package storage

import (
	"captcha-bot/internal/app/logic"
	"errors"
	"log"
	"sync"
	"time"
)

type UserInMemoryRepo struct {
	stateMap sync.Map
	ttl      time.Duration
	stop     chan struct{}
}

func NewUserInMemoryRepo(stateTTL, cleanupInterval time.Duration) *UserInMemoryRepo {
	st := &UserInMemoryRepo{
		stateMap: sync.Map{},
		ttl:      stateTTL,
		stop:     make(chan struct{}),
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

func (st *UserInMemoryRepo) Stop() {
	close(st.stop)
}

func (st *UserInMemoryRepo) removeExpired() {
	st.stateMap.Range(func(key, value interface{}) bool {
		itm, ok := value.(*logic.UserData)
		if !ok {
			return true
		}
		if time.Now().UnixNano() > itm.Expiration {
			log.Printf("Remove expired user state data for: %d", key)
			st.stateMap.Delete(key)
		}
		return true
	})
}

func (st *UserInMemoryRepo) GetByUserID(userID int64) (*logic.UserData, error) {
	userDataRaw, ok := st.stateMap.Load(userID)
	if !ok {
		return nil, logic.ErrStateNotFound
	}

	userData, ok := userDataRaw.(*logic.UserData)
	if !ok {
		return nil, errors.New("state decode error")
	}

	if userData.Expired() {
		st.stateMap.Delete(userID)
		return nil, logic.ErrStateNotFound
	}

	return userData, nil
}

func (st *UserInMemoryRepo) Put(userData *logic.UserData) error {
	if userData.UserID == 0 {
		return errors.New("couldn't put user with empty UserID")
	}

	expiration := int64(0)
	if st.ttl > 0 {
		expiration = time.Now().Add(st.ttl * time.Second).UnixNano()
	}
	userData.Expiration = expiration

	st.stateMap.Store(userData.UserID, userData)
	return nil
}
