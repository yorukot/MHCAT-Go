package domain

import (
	"errors"
	"math"
	"strings"
)

var ErrInvalidAutoChatConfig = errors.New("invalid autochat config")
var ErrInvalidAutoChatPaidRequest = errors.New("invalid paid autochat request")

type AutoChatConfig struct {
	GuildID   string
	ChannelID string
}

type AutoChatFallbackReply struct {
	Content        string
	UseTypingDelay bool
}

type AutoChatPaidSubmissionState string

const (
	AutoChatPaidIgnored     AutoChatPaidSubmissionState = "ignored"
	AutoChatPaidUnsafeInput AutoChatPaidSubmissionState = "unsafe_input"
	AutoChatPaidBusy        AutoChatPaidSubmissionState = "busy"
	AutoChatPaidQueued      AutoChatPaidSubmissionState = "queued"
)

type AutoChatPaidRequest struct {
	GuildID          string
	Content          string
	Cost             float64
	RequestedAtMilli int64
}

type AutoChatPaidDispatch struct {
	RequestTimeMilli  int64
	Cost              float64
	ConversationReset bool
}

type AutoChatPaidSubmission struct {
	State    AutoChatPaidSubmissionState
	Dispatch AutoChatPaidDispatch
}

type AutoChatPaidResponse struct {
	GuildID          string
	Content          string
	RequestTimeMilli int64
	Reply            bool
}

func (c AutoChatConfig) Validate() error {
	if strings.TrimSpace(c.GuildID) == "" || strings.TrimSpace(c.ChannelID) == "" {
		return ErrInvalidAutoChatConfig
	}
	return nil
}

func (r AutoChatPaidRequest) Validate() error {
	if strings.TrimSpace(r.GuildID) == "" || r.RequestedAtMilli <= 0 || r.Cost < 0 || math.IsNaN(r.Cost) || math.IsInf(r.Cost, 0) {
		return ErrInvalidAutoChatPaidRequest
	}
	return nil
}
