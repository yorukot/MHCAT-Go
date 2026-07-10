package documents

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type TicketConfigDocument struct {
	Guild         string `bson:"guild" json:"guild"`
	TicketChannel string `bson:"ticket_channel" json:"ticket_channel"`
	AdminID       string `bson:"admin_id" json:"admin_id"`
	EveryoneID    string `bson:"everyone_id" json:"everyone_id"`
}

type TicketConfigReadDocument struct {
	Guild         bson.RawValue `bson:"guild" json:"guild"`
	TicketChannel bson.RawValue `bson:"ticket_channel" json:"ticket_channel"`
	AdminID       bson.RawValue `bson:"admin_id" json:"admin_id"`
	EveryoneID    bson.RawValue `bson:"everyone_id" json:"everyone_id"`
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

func (d TicketConfigReadDocument) ToDomain() domain.TicketConfig {
	guild, _ := legacyMongooseString(d.Guild)
	ticketChannel, _ := legacyMongooseString(d.TicketChannel)
	adminID, _ := legacyMongooseString(d.AdminID)
	everyoneID, _ := legacyMongooseString(d.EveryoneID)
	return domain.TicketConfig{
		GuildID:        guild,
		CategoryID:     ticketChannel,
		AdminRoleID:    adminID,
		EveryoneRoleID: everyoneID,
	}
}
