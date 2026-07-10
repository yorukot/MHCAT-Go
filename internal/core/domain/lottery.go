package domain

import (
	"errors"
	"strings"
	"time"
)

var ErrInvalidLottery = errors.New("invalid lottery")

type Lottery struct {
	GuildID         string
	ID              string
	EndsAtUnix      int64
	Gift            string
	WinnerCount     int
	Participants    []LotteryParticipant
	Ended           bool
	ChannelID       string
	RequiredRoleID  string
	ForbiddenRoleID string
	MaxParticipants int
	OwnerID         string
}

type LotteryParticipant struct {
	UserID         string
	JoinedAtMillis int64
	JoinedAtRaw    string
}

type LotteryJoinRequest struct {
	GuildID        string
	ID             string
	UserID         string
	JoinedAtMillis int64
	NowUnix        int64
}

func (l Lottery) Normalized() Lottery {
	l.GuildID = strings.TrimSpace(l.GuildID)
	l.ID = strings.TrimSpace(l.ID)
	l.ChannelID = strings.TrimSpace(l.ChannelID)
	l.RequiredRoleID = strings.TrimSpace(l.RequiredRoleID)
	l.ForbiddenRoleID = strings.TrimSpace(l.ForbiddenRoleID)
	l.OwnerID = strings.TrimSpace(l.OwnerID)
	participants := make([]LotteryParticipant, 0, len(l.Participants))
	for _, participant := range l.Participants {
		participant.UserID = strings.TrimSpace(participant.UserID)
		participant.JoinedAtRaw = strings.TrimSpace(participant.JoinedAtRaw)
		if participant.UserID != "" {
			participants = append(participants, participant)
		}
	}
	l.Participants = participants
	return l
}

func (l Lottery) HasParticipant(userID string) bool {
	userID = strings.TrimSpace(userID)
	for _, participant := range l.Participants {
		if participant.UserID == userID {
			return true
		}
	}
	return false
}

func (l Lottery) IsExpired(now time.Time) bool {
	if now.IsZero() {
		now = time.Now()
	}
	return l.Ended || l.EndsAtUnix <= 0 || l.EndsAtUnix < now.Unix()
}

func (l Lottery) AtCapacity() bool {
	return l.MaxParticipants != 0 && len(l.Participants) >= l.MaxParticipants
}

func (r LotteryJoinRequest) Normalized() LotteryJoinRequest {
	r.GuildID = strings.TrimSpace(r.GuildID)
	r.ID = strings.TrimSpace(r.ID)
	r.UserID = strings.TrimSpace(r.UserID)
	return r
}

func (r LotteryJoinRequest) Validate() error {
	r = r.Normalized()
	if r.GuildID == "" || r.ID == "" || r.UserID == "" || r.JoinedAtMillis <= 0 || r.NowUnix <= 0 {
		return ErrInvalidLottery
	}
	return nil
}

func ValidateLotteryKey(guildID string, id string) error {
	if strings.TrimSpace(guildID) == "" || strings.TrimSpace(id) == "" {
		return ErrInvalidLottery
	}
	return nil
}
