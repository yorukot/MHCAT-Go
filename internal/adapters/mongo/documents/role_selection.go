package documents

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"

type RoleReactionDocument struct {
	Guild   string `bson:"guild" json:"guild"`
	Message string `bson:"message" json:"message"`
	React   string `bson:"react" json:"react"`
	Role    string `bson:"role" json:"role"`
}

type RoleButtonDocument struct {
	Guild  string `bson:"guild" json:"guild"`
	Number string `bson:"number" json:"number"`
	Role   string `bson:"role" json:"role"`
}

func RoleReactionDocumentFromDomain(config domain.RoleReactionConfig) RoleReactionDocument {
	return RoleReactionDocument{
		Guild:   config.GuildID,
		Message: config.MessageID,
		React:   config.React,
		Role:    config.RoleID,
	}
}

func (d RoleReactionDocument) ToDomain() domain.RoleReactionConfig {
	return domain.RoleReactionConfig{
		GuildID:   d.Guild,
		MessageID: d.Message,
		React:     d.React,
		RoleID:    d.Role,
	}
}

func RoleButtonDocumentFromDomain(config domain.RoleButtonConfig) RoleButtonDocument {
	return RoleButtonDocument{
		Guild:  config.GuildID,
		Number: config.Number,
		Role:   config.RoleID,
	}
}

func (d RoleButtonDocument) ToDomain() domain.RoleButtonConfig {
	return domain.RoleButtonConfig{
		GuildID: d.Guild,
		Number:  d.Number,
		RoleID:  d.Role,
	}
}
