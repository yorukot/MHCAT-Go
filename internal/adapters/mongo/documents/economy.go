package documents

import (
	"strconv"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type CoinDocument struct {
	Guild  string        `bson:"guild" json:"guild"`
	Member string        `bson:"member" json:"member"`
	Coin   bson.RawValue `bson:"coin" json:"coin"`
	Today  bson.RawValue `bson:"today" json:"today"`
}

type GiftChangeDocument struct {
	Guild      string        `bson:"guild" json:"guild"`
	CoinNumber bson.RawValue `bson:"coin_number" json:"coin_number"`
	SignCoin   bson.RawValue `bson:"sign_coin" json:"sign_coin"`
	Channel    string        `bson:"channel" json:"channel"`
	XPMultiple bson.RawValue `bson:"xp_multiple" json:"xp_multiple"`
	Time       bson.RawValue `bson:"time" json:"time"`
}

type SignListDocument struct {
	Guild  string                         `bson:"guild" json:"guild"`
	Member string                         `bson:"member" json:"member"`
	Date   map[string]map[string][]string `bson:"date" json:"date"`
}

func (d CoinDocument) ToDomain() domain.CoinBalance {
	return domain.CoinBalance{
		GuildID: d.Guild,
		UserID:  d.Member,
		Coins:   legacyInt64(d.Coin),
		Today:   legacyInt64(d.Today),
	}
}

func (d GiftChangeDocument) ToDomain() domain.EconomyConfig {
	return domain.EconomyConfig{
		GuildID:     d.Guild,
		GachaCost:   legacyInt64(d.CoinNumber),
		SignCoins:   legacyInt64(d.SignCoin),
		ChannelID:   d.Channel,
		XPMultiple:  legacyFloat64(d.XPMultiple),
		ResetMarker: legacyInt64(d.Time),
	}
}

func GiftChangeUpdateFromDomain(config domain.EconomyConfig) bson.D {
	return bson.D{
		{Key: "coin_number", Value: config.GachaCost},
		{Key: "sign_coin", Value: config.SignCoins},
		{Key: "channel", Value: config.ChannelID},
		{Key: "xp_multiple", Value: config.XPMultiple},
		{Key: "time", Value: config.ResetMarker},
	}
}

func (d SignListDocument) ToDomain() domain.SignCalendar {
	date := map[string]map[string][]string{}
	for year, months := range d.Date {
		date[year] = map[string][]string{}
		for month, days := range months {
			date[year][month] = append([]string(nil), days...)
		}
	}
	return domain.SignCalendar{
		GuildID: d.Guild,
		UserID:  d.Member,
		Date:    date,
	}
}

func SignListDocumentFromDomain(calendar domain.SignCalendar) SignListDocument {
	date := map[string]map[string][]string{}
	for year, months := range calendar.Date {
		date[year] = map[string][]string{}
		for month, days := range months {
			date[year][month] = append([]string(nil), days...)
		}
	}
	return SignListDocument{
		Guild:  calendar.GuildID,
		Member: calendar.UserID,
		Date:   date,
	}
}

func legacyInt64(value bson.RawValue) int64 {
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

func legacyFloat64(value bson.RawValue) float64 {
	if value.Type == 0 || value.Type == bson.TypeNull || value.Type == bson.TypeUndefined {
		return 0
	}
	if parsed, ok := value.DoubleOK(); ok {
		return parsed
	}
	if parsed, ok := value.AsInt64OK(); ok {
		return float64(parsed)
	}
	if text, ok := value.StringValueOK(); ok {
		number, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
		if err == nil {
			return number
		}
	}
	return 0
}
