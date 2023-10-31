package logic

import "errors"

type UserState int8

const (
	Default UserState = iota
	Check
	Approved
	Ban
)

type ButtonEvent int8

const (
	Left  ButtonEvent = -1
	Right ButtonEvent = 1
)

var ErrStateNotFound = errors.New("no state found for user")

const CaptchaLength = 11
