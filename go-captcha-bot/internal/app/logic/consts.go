package logic

import "errors"

type UserState int8

const (
	Default UserState = iota
	Check
	Approved
	Ban
)

type ButtonEvent uint8

const (
	Left ButtonEvent = iota
	Right
)

var ErrStateNotFound = errors.New("no state found for user")

const CaptchaLength = 11
