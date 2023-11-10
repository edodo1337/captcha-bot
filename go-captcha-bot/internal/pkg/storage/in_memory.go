package storage

import (
	"captcha-bot/internal/app/logic"
	"errors"
	"log"
	"sync"
	"time"
)

type UserInMemoryStorage struct {
	stateMap sync.Map
	ttl      time.Duration
	stop     chan struct{}
}

func NewUserInMemoryStorage(stateTTL, cleanupInterval time.Duration) *UserInMemoryStorage {
	st := &UserInMemoryStorage{
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

func (st *UserInMemoryStorage) Stop() {
	close(st.stop)
}

func (st *UserInMemoryStorage) removeExpired() {
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

func (st *UserInMemoryStorage) SetState(userId int64, state logic.UserState) error {
	userData, err := st.GetData(userId)

	expiration := int64(0)
	if st.ttl > 0 {
		expiration = time.Now().Add(st.ttl * time.Second).UnixNano()
	}

	if errors.Is(err, logic.ErrStateNotFound) {
		st.stateMap.Store(
			userId,
			&logic.UserData{
				State:      state,
				Expiration: expiration,
			},
		)
		return nil
	}

	userData.State = state
	userData.Expiration = expiration
	st.stateMap.Store(userId, &userData)

	return nil
}

func (st *UserInMemoryStorage) GetState(userId int64) (logic.UserState, error) {
	userDataRaw, ok := st.stateMap.Load(userId)
	if !ok {
		return logic.Default, logic.ErrStateNotFound
	}

	userData, ok := userDataRaw.(*logic.UserData)
	if !ok {
		return logic.Default, errors.New("state decode error")
	}

	if userData.Expired() {
		st.stateMap.Delete(userId)
		return logic.Default, logic.ErrStateNotFound
	}

	return userData.State, nil
}

func (st *UserInMemoryStorage) SetData(userId int64, data *logic.UserData) error {
	expiration := int64(0)
	if st.ttl > 0 {
		expiration = time.Now().Add(st.ttl * time.Second).UnixNano()
	}
	data.Expiration = expiration

	st.stateMap.Store(userId, data)
	return nil
}

func (st *UserInMemoryStorage) GetData(userId int64) (*logic.UserData, error) {
	userDataRaw, ok := st.stateMap.Load(userId)
	if !ok {
		return &logic.UserData{}, logic.ErrStateNotFound
	}

	userData, ok := userDataRaw.(*logic.UserData)
	if !ok {
		return nil, errors.New("state decode error")
	}

	if userData.Expired() {
		st.stateMap.Delete(userId)
		return nil, logic.ErrStateNotFound
	}

	return userData, nil
}
