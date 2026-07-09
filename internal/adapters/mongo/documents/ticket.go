package documents

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"

type TicketConfigDocument struct {
	Guild         string `bson:"guild" json:"guild"`
	TicketChannel string `bson:"ticket_channel" json:"ticket_channel"`
	AdminID       string `bson:"admin_id" json:"admin_id"`
	EveryoneID    string `bson:"everyone_id" json:"everyone_id"`
}

func TicketConfigDocumentFromDomain(config domain.TicketConfig) TicketConfigDocument {
	return TicketConfigDocument{
		Guild:         config.GuildID,
		TicketChannel: config.CategoryID,
		AdminID:       config.AdminRoleID,
		EveryoneID:    config.EveryoneRoleID,
	}
}

func (d TicketConfigDocument) ToDomain() domain.TicketConfig {
	return domain.TicketConfig{
		GuildID:        d.Guild,
		CategoryID:     d.TicketChannel,
		AdminRoleID:    d.AdminID,
		EveryoneRoleID: d.EveryoneID,
	}
}
