package documents

import (
	"strconv"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type CoinDocument struct {
	ID     bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
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

type ShopItemDocument struct {
	Guild                string        `bson:"guild" json:"guild"`
	CommodityID          bson.RawValue `bson:"commodity_id" json:"commodity_id"`
	Name                 string        `bson:"name" json:"name"`
	NeedCoin             bson.RawValue `bson:"need_coin" json:"need_coin"`
	CommodityDescription string        `bson:"commodity_description" json:"commodity_description"`
	Code                 *string       `bson:"code" json:"code"`
	AutoDelete           bool          `bson:"auto_delete" json:"auto_delete"`
	Role                 *string       `bson:"role" json:"role"`
	CommodityCount       bson.RawValue `bson:"commodity_count" json:"commodity_count"`
}

func (d CoinDocument) ToDomain() domain.CoinBalance {
	return domain.CoinBalance{
		GuildID:   d.Guild,
		UserID:    d.Member,
		Coins:     legacyInt64(d.Coin),
		CoinsText: legacyPriceString(d.Coin),
		Today:     legacyInt64(d.Today),
	}
}

func (d GiftChangeDocument) ToDomain() domain.EconomyConfig {
	return domain.EconomyConfig{
		GuildID:       d.Guild,
		GachaCost:     legacyInt64(d.CoinNumber),
		GachaCostText: legacyPriceString(d.CoinNumber),
		SignCoins:     legacyInt64(d.SignCoin),
		ChannelID:     d.Channel,
		XPMultiple:    legacyFloat64(d.XPMultiple),
		ResetMarker:   legacyInt64(d.Time),
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

func (d ShopItemDocument) ToDomain() domain.ShopItem {
	code := ""
	if d.Code != nil {
		code = *d.Code
	}
	roleID := ""
	if d.Role != nil {
		roleID = *d.Role
	}
	return domain.ShopItem{
		GuildID:     d.Guild,
		CommodityID: legacyInt64(d.CommodityID),
		Name:        d.Name,
		NeedCoins:   legacyInt64(d.NeedCoin),
		Description: d.CommodityDescription,
		Code:        code,
		AutoDelete:  d.AutoDelete,
		RoleID:      roleID,
		Count:       legacyInt64(d.CommodityCount),
	}
}

func ShopItemWriteDocumentFromDomain(item domain.ShopItem) bson.D {
	item = item.Normalize()
	var code any
	if item.Code != "" {
		code = item.Code
	}
	var role any
	if item.RoleID != "" {
		role = item.RoleID
	}
	return bson.D{
		{Key: "guild", Value: item.GuildID},
		{Key: "commodity_id", Value: item.CommodityID},
		{Key: "name", Value: item.Name},
		{Key: "need_coin", Value: item.NeedCoins},
		{Key: "commodity_description", Value: item.Description},
		{Key: "code", Value: code},
		{Key: "auto_delete", Value: item.AutoDelete},
		{Key: "role", Value: role},
		{Key: "commodity_count", Value: item.Count},
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
