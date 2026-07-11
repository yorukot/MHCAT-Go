package documents

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type StatsConfigDocument struct {
	Guild             bson.RawValue `bson:"guild" json:"guild"`
	Parent            bson.RawValue `bson:"parent" json:"parent"`
	MemberNumber      bson.RawValue `bson:"memberNumber" json:"memberNumber"`
	MemberNumberName  bson.RawValue `bson:"memberNumber_name" json:"memberNumber_name"`
	UserNumber        bson.RawValue `bson:"userNumber" json:"userNumber"`
	UserNumberName    bson.RawValue `bson:"userNumber_name" json:"userNumber_name"`
	BotNumber         bson.RawValue `bson:"BotNumber" json:"BotNumber"`
	BotNumberName     bson.RawValue `bson:"BotNumber_name" json:"BotNumber_name"`
	ChannelNumber     bson.RawValue `bson:"channelnumber" json:"channelnumber"`
	ChannelNumberName bson.RawValue `bson:"channelnumber_name" json:"channelnumber_name"`
	TextNumber        bson.RawValue `bson:"textnumber" json:"textnumber"`
	TextNumberName    bson.RawValue `bson:"textnumber_name" json:"textnumber_name"`
	VoiceNumber       bson.RawValue `bson:"voicenumber" json:"voicenumber"`
	VoiceNumberName   bson.RawValue `bson:"voicenumber_name" json:"voicenumber_name"`
}

type StatsRoleConfigDocument struct {
	Guild       bson.RawValue `bson:"guild" json:"guild"`
	Channel     bson.RawValue `bson:"channel" json:"channel"`
	ChannelName bson.RawValue `bson:"channel_name" json:"channel_name"`
	Role        bson.RawValue `bson:"role" json:"role"`
}

type StatsRoleConfigWriteDocument struct {
	Guild       string `bson:"guild" json:"guild"`
	Channel     string `bson:"channel" json:"channel"`
	ChannelName string `bson:"channel_name" json:"channel_name"`
	Role        string `bson:"role" json:"role"`
}

func StatsRoleConfigDocumentFromDomain(config domain.StatsRoleConfig) StatsRoleConfigWriteDocument {
	config = config.Normalize()
	return StatsRoleConfigWriteDocument{
		Guild:       config.GuildID,
		Channel:     config.ChannelID,
		ChannelName: config.ChannelName,
		Role:        config.RoleID,
	}
}

func (d StatsRoleConfigDocument) ToDomain() domain.StatsRoleConfig {
	return domain.StatsRoleConfig{
		GuildID:     statsMongooseString(d.Guild),
		ChannelID:   statsMongooseString(d.Channel),
		ChannelName: statsMongooseString(d.ChannelName),
		RoleID:      statsMongooseString(d.Role),
	}
}

func (d StatsConfigDocument) ToDomain() domain.StatsConfig {
	return domain.StatsConfig{
		GuildID:           statsMongooseString(d.Guild),
		ParentID:          statsMongooseString(d.Parent),
		MemberNumberID:    statsMongooseString(d.MemberNumber),
		MemberNumberName:  statsMongooseString(d.MemberNumberName),
		UserNumberID:      statsMongooseString(d.UserNumber),
		UserNumberName:    statsMongooseString(d.UserNumberName),
		BotNumberID:       statsMongooseString(d.BotNumber),
		BotNumberName:     statsMongooseString(d.BotNumberName),
		ChannelNumberID:   statsMongooseString(d.ChannelNumber),
		ChannelNumberName: statsMongooseString(d.ChannelNumberName),
		TextNumberID:      statsMongooseString(d.TextNumber),
		TextNumberName:    statsMongooseString(d.TextNumberName),
		VoiceNumberID:     statsMongooseString(d.VoiceNumber),
		VoiceNumberName:   statsMongooseString(d.VoiceNumberName),
	}
}

func statsMongooseString(value bson.RawValue) string {
	if text, ok := legacyMongooseString(value); ok {
		return text
	}
	return ""
}
