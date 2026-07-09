package documents

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"

type AutoChatConfigDocument struct {
	Guild   string `bson:"guild" json:"guild"`
	Channel string `bson:"channel" json:"channel"`
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
