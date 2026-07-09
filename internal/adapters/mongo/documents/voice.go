package documents

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"

type VoiceRoomConfigDocument struct {
	Guild         string `bson:"guild" json:"guild"`
	TicketChannel string `bson:"ticket_channel" json:"ticket_channel"`
	Limit         int    `bson:"limit" json:"limit"`
	Name          string `bson:"name" json:"name"`
	Parent        string `bson:"parent,omitempty" json:"parent,omitempty"`
	Lock          bool   `bson:"lock" json:"lock"`
}

func VoiceRoomConfigDocumentFromDomain(config domain.VoiceRoomConfig) VoiceRoomConfigDocument {
	return VoiceRoomConfigDocument{
		Guild:         config.GuildID,
		TicketChannel: config.TriggerChannelID,
		Limit:         config.Limit,
		Name:          config.Name,
		Parent:        config.ParentID,
		Lock:          config.Lock,
	}
}

func (d VoiceRoomConfigDocument) ToDomain() domain.VoiceRoomConfig {
	return domain.VoiceRoomConfig{
		GuildID:          d.Guild,
		TriggerChannelID: d.TicketChannel,
		ParentID:         d.Parent,
		Name:             d.Name,
		Limit:            d.Limit,
		Lock:             d.Lock,
	}
}
