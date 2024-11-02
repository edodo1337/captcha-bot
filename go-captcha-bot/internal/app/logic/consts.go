package logic

import "errors"

type UserState int8

type ButtonEvent int8

type KickFailedReason int8

const (
	Default UserState = iota
	Check
	CaptchaPassed
	Approved
	Ban
	Left              ButtonEvent      = -1
	Right             ButtonEvent      = 1
	NotEnoughVotes    KickFailedReason = 1
	MinVotesThreesold KickFailedReason = 2
)

var ErrStateNotFound = errors.New("no state found for user")
var ErrPollNotFound = errors.New("no poll found by user id")
var ErrUserIsAdmin = errors.New("user is admin")
var ErrPollAlreadyExist = errors.New("poll already exist for user")

const CaptchaLength = 11
const DEFAULT_LINES_NUM = 10
