package documents

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

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

type GuildAnnouncementReadDocument struct {
	Guild          bson.RawValue `bson:"guild" json:"guild"`
	AnnouncementID bson.RawValue `bson:"announcement_id" json:"announcement_id"`
}

type BoundAnnouncementReadDocument struct {
	Guild          bson.RawValue `bson:"guild" json:"guild"`
	AnnouncementID bson.RawValue `bson:"announcement_id" json:"announcement_id"`
	Tag            bson.RawValue `bson:"tag" json:"tag"`
	Color          bson.RawValue `bson:"color" json:"color"`
	Title          bson.RawValue `bson:"title" json:"title"`
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

func (d GuildAnnouncementReadDocument) ToDomain() domain.AnnouncementChannelConfig {
	guild, _ := legacyMongooseString(d.Guild)
	announcementID, _ := legacyMongooseString(d.AnnouncementID)
	return domain.AnnouncementChannelConfig{GuildID: guild, ChannelID: announcementID}
}

func (d BoundAnnouncementReadDocument) ToDomain() domain.BoundAnnouncementConfig {
	guild, _ := legacyMongooseString(d.Guild)
	announcementID, _ := legacyMongooseString(d.AnnouncementID)
	tag, _ := legacyMongooseString(d.Tag)
	color, _ := legacyMongooseString(d.Color)
	title, _ := legacyMongooseString(d.Title)
	return domain.BoundAnnouncementConfig{
		GuildID:   guild,
		ChannelID: announcementID,
		Tag:       tag,
		Color:     color,
		Title:     title,
	}
}
