package documents

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"

type GuildAnnouncementDocument struct {
	Guild          string `bson:"guild" json:"guild"`
	AnnouncementID string `bson:"announcement_id,omitempty" json:"announcement_id,omitempty"`
	VoiceDetection string `bson:"voice_detection,omitempty" json:"voice_detection,omitempty"`
}

type BoundAnnouncementDocument struct {
	Guild          string `bson:"guild" json:"guild"`
	AnnouncementID string `bson:"announcement_id" json:"announcement_id"`
	Tag            string `bson:"tag" json:"tag"`
	Color          string `bson:"color" json:"color"`
	Title          string `bson:"title" json:"title"`
}

func GuildAnnouncementDocumentFromDomain(config domain.AnnouncementChannelConfig) GuildAnnouncementDocument {
	return GuildAnnouncementDocument{
		Guild:          config.GuildID,
		AnnouncementID: config.ChannelID,
	}
}

func BoundAnnouncementDocumentFromDomain(config domain.BoundAnnouncementConfig) BoundAnnouncementDocument {
	return BoundAnnouncementDocument{
		Guild:          config.GuildID,
		AnnouncementID: config.ChannelID,
		Tag:            config.Tag,
		Color:          config.Color,
		Title:          config.Title,
	}
}
