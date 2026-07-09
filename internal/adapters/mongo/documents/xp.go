package documents

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"

type TextXPChannelDocument struct {
	Guild      string `bson:"guild" json:"guild"`
	Channel    string `bson:"channel" json:"channel"`
	Background string `bson:"background,omitempty" json:"background,omitempty"`
	Color      string `bson:"color,omitempty" json:"color,omitempty"`
	Message    string `bson:"message,omitempty" json:"message,omitempty"`
}

type VoiceXPChannelDocument struct {
	Guild      string `bson:"guild" json:"guild"`
	Channel    string `bson:"channel" json:"channel"`
	Background string `bson:"background,omitempty" json:"background,omitempty"`
	Color      string `bson:"color,omitempty" json:"color,omitempty"`
	Message    string `bson:"message,omitempty" json:"message,omitempty"`
}

func TextXPChannelDocumentFromDomain(config domain.TextXPConfig) TextXPChannelDocument {
	return TextXPChannelDocument{
		Guild:   config.GuildID,
		Channel: config.ChannelID,
		Color:   config.Color,
		Message: config.Message,
	}
}

func VoiceXPChannelDocumentFromDomain(config domain.VoiceXPConfig) VoiceXPChannelDocument {
	return VoiceXPChannelDocument{
		Guild:   config.GuildID,
		Channel: config.ChannelID,
		Color:   config.Color,
		Message: config.Message,
	}
}

func (d TextXPChannelDocument) ToDomain() domain.TextXPConfig {
	return domain.TextXPConfig{
		GuildID:   d.Guild,
		ChannelID: d.Channel,
		Color:     d.Color,
		Message:   d.Message,
	}
}

func (d VoiceXPChannelDocument) ToDomain() domain.VoiceXPConfig {
	return domain.VoiceXPConfig{
		GuildID:   d.Guild,
		ChannelID: d.Channel,
		Color:     d.Color,
		Message:   d.Message,
	}
}
