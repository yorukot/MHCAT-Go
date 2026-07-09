package domain

import (
	"errors"
	"strings"
)

var ErrInvalidDeleteDataRequest = errors.New("invalid delete data request")

type DeleteDataTarget string

const (
	DeleteDataTargetJoinMessage  DeleteDataTarget = "加入訊息"
	DeleteDataTargetLeaveMessage DeleteDataTarget = "離開訊息"
	DeleteDataTargetLogging      DeleteDataTarget = "審核日誌"
	DeleteDataTargetStats        DeleteDataTarget = "統計系統"
	DeleteDataTargetAutoChat     DeleteDataTarget = "自動聊天"
	DeleteDataTargetVerification DeleteDataTarget = "驗證設置"
	DeleteDataTargetTextXP       DeleteDataTarget = "聊天經驗設置"
	DeleteDataTargetVoiceXP      DeleteDataTarget = "語音經驗設置"
	DeleteDataTargetTicket       DeleteDataTarget = "私人頻道設置"
)

var legacyDeleteDataTargets = []DeleteDataTarget{
	DeleteDataTargetJoinMessage,
	DeleteDataTargetLeaveMessage,
	DeleteDataTargetLogging,
	DeleteDataTargetStats,
	DeleteDataTargetAutoChat,
	DeleteDataTargetVerification,
	DeleteDataTargetTextXP,
	DeleteDataTargetVoiceXP,
	DeleteDataTargetTicket,
}

type DeleteDataRequest struct {
	GuildID string
	Target  DeleteDataTarget
}

func LegacyDeleteDataTargets() []DeleteDataTarget {
	return append([]DeleteDataTarget(nil), legacyDeleteDataTargets...)
}

func ParseDeleteDataTarget(value string) (DeleteDataTarget, bool) {
	target := DeleteDataTarget(strings.TrimSpace(value))
	for _, candidate := range legacyDeleteDataTargets {
		if target == candidate {
			return target, true
		}
	}
	return "", false
}

func (r DeleteDataRequest) Normalize() DeleteDataRequest {
	target, _ := ParseDeleteDataTarget(string(r.Target))
	return DeleteDataRequest{
		GuildID: strings.TrimSpace(r.GuildID),
		Target:  target,
	}
}

func (r DeleteDataRequest) Validate() error {
	normalized := r.Normalize()
	if normalized.GuildID == "" || normalized.Target == "" {
		return ErrInvalidDeleteDataRequest
	}
	return nil
}
