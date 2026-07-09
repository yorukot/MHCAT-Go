package documents

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"

type LoggingConfigDocument struct {
	Guild             string `bson:"guild" json:"guild"`
	ChannelID         string `bson:"channel_id" json:"channel_id"`
	MessageUpdate     bool   `bson:"message_update" json:"message_update"`
	MessageDelete     bool   `bson:"message_delete" json:"message_delete"`
	ChannelUpdate     bool   `bson:"channel_update" json:"channel_update"`
	MemberVoiceUpdate bool   `bson:"member_voice_update" json:"member_voice_update"`
}

func LoggingConfigDocumentFromDomain(config domain.LoggingConfig) LoggingConfigDocument {
	return LoggingConfigDocument{
		Guild:             config.GuildID,
		ChannelID:         config.ChannelID,
		MessageUpdate:     config.MessageUpdate,
		MessageDelete:     config.MessageDelete,
		ChannelUpdate:     config.ChannelUpdate,
		MemberVoiceUpdate: config.MemberVoiceUpdate,
	}
}

func (d LoggingConfigDocument) ToDomain() domain.LoggingConfig {
	return domain.LoggingConfig{
		GuildID:           d.Guild,
		ChannelID:         d.ChannelID,
		MessageUpdate:     d.MessageUpdate,
		MessageDelete:     d.MessageDelete,
		ChannelUpdate:     d.ChannelUpdate,
		MemberVoiceUpdate: d.MemberVoiceUpdate,
	}
}
