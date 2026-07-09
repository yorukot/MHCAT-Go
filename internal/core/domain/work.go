package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
)

var (
	ErrInvalidWorkQuery       = errors.New("invalid work query")
	ErrWorkItemKeyConflict    = errors.New("work item key conflict")
	ErrWorkEnergyInsufficient = errors.New("work energy insufficient")
	ErrWorkAlreadyBusy        = errors.New("work user already working")
	ErrWorkStartUnavailable   = errors.New("work start unavailable")
	ErrWorkAdminUnavailable   = errors.New("work admin unavailable")
)

const WorkIdleState = "待業中"

type WorkConfig struct {
	GuildID     string
	DailyEnergy int64
	MaxEnergy   int64
	Captcha     bool
}

type WorkItem struct {
	GuildID     string
	Name        string
	DurationSec int64
	EnergyCost  int64
	CoinReward  int64
	RoleID      string
}

func (i WorkItem) Key() string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(i.Name)))
	return hex.EncodeToString(sum[:])[:12]
}

type WorkUserState struct {
	GuildID     string
	UserID      string
	State       string
	EndTimeUnix int64
	Energy      int64
	GetCoin     int64
	Initialized bool
}

type WorkStartCommand struct {
	GuildID     string
	UserID      string
	WorkName    string
	DurationSec int64
	EnergyCost  int64
	CoinReward  int64
	MaxEnergy   int64
	NowUnix     int64
	Override    bool
}

type WorkConfigCommand struct {
	GuildID     string
	DailyEnergy int64
	MaxEnergy   int64
	Captcha     bool
}

type WorkDeleteItemCommand struct {
	GuildID string
	Name    string
}

type WorkEnergyGrantCommand struct {
	GuildID   string
	UserID    string
	Amount    int64
	MaxEnergy int64
}

type WorkEnergyGrantAllCommand struct {
	GuildID   string
	Amount    int64
	MaxEnergy int64
}

type WorkEnergyGrantAllResult struct {
	Matched  int64
	Modified int64
}

func (s WorkUserState) EffectiveState(nowUnix int64) string {
	if s.EndTimeUnix-nowUnix > 0 && strings.TrimSpace(s.State) != "" {
		return "在" + s.State + "打工"
	}
	return WorkIdleState
}

func (s WorkUserState) RemainingTimeText(nowUnix int64) string {
	if s.EndTimeUnix-nowUnix > 0 {
		return "<t:" + strconv.FormatInt(s.EndTimeUnix, 10) + ":R>"
	}
	return "`沒有打工再進行`"
}

type WorkInterfaceView struct {
	Config        WorkConfig
	User          WorkUserState
	Items         []WorkItem
	VisibleItems  []WorkItem
	NowUnix       int64
	GuildName     string
	UserTag       string
	UserAvatarURL string
}
