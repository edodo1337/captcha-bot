package fsm

import (
	"captcha-bot/internal/app/logic"
	"errors"
	"sync"
	"time"
)

type InMemoryFSM struct {
	stateMap sync.Map
}

func NewInMemoryFSM() *InMemoryFSM {
	return &InMemoryFSM{
		stateMap: sync.Map{},
	}
}

func (fsm *InMemoryFSM) SetState(userId int64, state logic.UserState) error {
	userData, err := fsm.GetData(userId)
	if errors.Is(err, logic.ErrStateNotFound) {
		fsm.stateMap.Store(userId, logic.UserData{State: state, StartTime: time.Now()})
		return nil
	}

	userData.State = state
	fsm.stateMap.Store(userId, userData)

	return nil
}

func (fsm *InMemoryFSM) GetState(userId int64) (logic.UserState, error) {
	userDataRaw, ok := fsm.stateMap.Load(userId)
	if !ok {
		return logic.Default, logic.ErrStateNotFound
	}

	userData, ok := userDataRaw.(*logic.UserData)
	if !ok {
		return logic.Default, errors.New("state decode error")
	}

	return userData.State, nil
}

func (fsm *InMemoryFSM) SetData(userId int64, data *logic.UserData) error {
	fsm.stateMap.Store(userId, data)
	return nil
}

func (fsm *InMemoryFSM) GetData(userId int64) (*logic.UserData, error) {
	userDataRaw, ok := fsm.stateMap.Load(userId)
	if !ok {
		return &logic.UserData{}, logic.ErrStateNotFound
	}

	userData, ok := userDataRaw.(*logic.UserData)
	if !ok {
		return nil, errors.New("state decode error")
	}

	return userData, nil
}
