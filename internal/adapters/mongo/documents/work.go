package documents

import (
	"strconv"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type WorkConfigDocument struct {
	Guild      string        `bson:"guild" json:"guild"`
	GetEnergy  bson.RawValue `bson:"get_energy" json:"get_energy"`
	MaxEnergy  bson.RawValue `bson:"max_energy" json:"max_energy"`
	CaptchaRaw bson.RawValue `bson:"captcha" json:"captcha"`
}

type WorkItemDocument struct {
	Guild  string        `bson:"guild" json:"guild"`
	Name   string        `bson:"name" json:"name"`
	Time   bson.RawValue `bson:"time" json:"time"`
	Energy bson.RawValue `bson:"energy" json:"energy"`
	Coin   bson.RawValue `bson:"coin" json:"coin"`
	Role   string        `bson:"role" json:"role"`
}

type WorkUserDocument struct {
	Guild   string        `bson:"guild" json:"guild"`
	User    string        `bson:"user" json:"user"`
	State   string        `bson:"state" json:"state"`
	EndTime bson.RawValue `bson:"end_time" json:"end_time"`
	Energi  bson.RawValue `bson:"energi" json:"energi"`
	GetCoin bson.RawValue `bson:"get_coin" json:"get_coin"`
}

func (d WorkConfigDocument) ToDomain() domain.WorkConfig {
	return domain.WorkConfig{
		GuildID:         d.Guild,
		DailyEnergy:     workLegacyInt64(d.GetEnergy),
		DailyEnergyText: legacyPriceString(d.GetEnergy),
		MaxEnergy:       workLegacyInt64(d.MaxEnergy),
		MaxEnergyText:   legacyPriceString(d.MaxEnergy),
		Captcha:         workLegacyBool(d.CaptchaRaw),
	}
}

func (d WorkItemDocument) ToDomain() domain.WorkItem {
	return domain.WorkItem{
		GuildID:     d.Guild,
		Name:        d.Name,
		DurationSec: workLegacyInt64(d.Time),
		EnergyCost:  workLegacyInt64(d.Energy),
		CoinReward:  workLegacyInt64(d.Coin),
		RoleID:      d.Role,
	}
}

func (d WorkUserDocument) ToDomain() domain.WorkUserState {
	state := strings.TrimSpace(d.State)
	if state == "" {
		state = domain.WorkIdleState
	}
	return domain.WorkUserState{
		GuildID:     d.Guild,
		UserID:      d.User,
		State:       state,
		EndTimeUnix: workLegacyInt64(d.EndTime),
		EndTimeText: legacyPriceString(d.EndTime),
		Energy:      workLegacyInt64(d.Energi),
		EnergyText:  legacyPriceString(d.Energi),
		GetCoin:     workLegacyInt64(d.GetCoin),
		Initialized: true,
	}
}

func workLegacyInt64(value bson.RawValue) int64 {
	if value.Type == 0 || value.Type == bson.TypeNull || value.Type == bson.TypeUndefined {
		return 0
	}
	if parsed, ok := value.AsInt64OK(); ok {
		return parsed
	}
	if parsed, ok := value.DoubleOK(); ok {
		return int64(parsed)
	}
	if text, ok := value.StringValueOK(); ok {
		number, err := strconv.ParseInt(strings.TrimSpace(text), 10, 64)
		if err == nil {
			return number
		}
	}
	return 0
}

func workLegacyBool(value bson.RawValue) bool {
	if value.Type == 0 || value.Type == bson.TypeNull || value.Type == bson.TypeUndefined {
		return false
	}
	if parsed, ok := value.BooleanOK(); ok {
		return parsed
	}
	if text, ok := value.StringValueOK(); ok {
		parsed, err := strconv.ParseBool(strings.TrimSpace(text))
		return err == nil && parsed
	}
	return false
}
