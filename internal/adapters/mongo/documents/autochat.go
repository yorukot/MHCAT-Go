package documents

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type AutoChatConfigDocument struct {
	Guild   string `bson:"guild" json:"guild"`
	Channel string `bson:"channel" json:"channel"`
}

type AutoChatConfigReadDocument struct {
	Guild   bson.RawValue `bson:"guild" json:"guild"`
	Channel bson.RawValue `bson:"channel" json:"channel"`
}

func AutoChatConfigDocumentFromDomain(config domain.AutoChatConfig) AutoChatConfigDocument {
	return AutoChatConfigDocument{
		Guild:   config.GuildID,
		Channel: config.ChannelID,
	}
}

func (d AutoChatConfigDocument) ToDomain() domain.AutoChatConfig {
	return domain.AutoChatConfig{
		GuildID:   d.Guild,
		ChannelID: d.Channel,
	}
}

func (d AutoChatConfigReadDocument) ToDomain() domain.AutoChatConfig {
	return domain.AutoChatConfig{
		GuildID:   autoChatMongooseString(d.Guild),
		ChannelID: autoChatMongooseString(d.Channel),
	}
}

func autoChatMongooseString(value bson.RawValue) string {
	if text, ok := legacyMongooseString(value); ok {
		return text
	}
	if _, data, ok := value.BinaryOK(); ok {
		return string(data)
	}
	return ""
}
