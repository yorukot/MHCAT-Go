package documents

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

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

type RoleReactionReadDocument struct {
	Guild   bson.RawValue `bson:"guild" json:"guild"`
	Message bson.RawValue `bson:"message" json:"message"`
	React   bson.RawValue `bson:"react" json:"react"`
	Role    bson.RawValue `bson:"role" json:"role"`
}

type RoleButtonReadDocument struct {
	Guild  bson.RawValue `bson:"guild" json:"guild"`
	Number bson.RawValue `bson:"number" json:"number"`
	Role   bson.RawValue `bson:"role" json:"role"`
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

func (d RoleReactionReadDocument) ToDomain() domain.RoleReactionConfig {
	guild, _ := legacyMongooseString(d.Guild)
	message, _ := legacyMongooseString(d.Message)
	react, _ := legacyMongooseString(d.React)
	role, _ := legacyMongooseString(d.Role)
	return domain.RoleReactionConfig{
		GuildID:   guild,
		MessageID: message,
		React:     react,
		RoleID:    role,
	}
}

func (d RoleButtonReadDocument) ToDomain() domain.RoleButtonConfig {
	guild, _ := legacyMongooseString(d.Guild)
	number, _ := legacyMongooseString(d.Number)
	role, _ := legacyMongooseString(d.Role)
	return domain.RoleButtonConfig{
		GuildID: guild,
		Number:  number,
		RoleID:  role,
	}
}
