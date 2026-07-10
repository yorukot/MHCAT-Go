package documents

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type GoodWebConfigDocument struct {
	Guild string `bson:"guild" json:"guild"`
	Open  bool   `bson:"open" json:"open"`
}

type GoodWebConfigReadDocument struct {
	Guild string        `bson:"guild" json:"guild"`
	Open  bson.RawValue `bson:"open" json:"open"`
}

type ScamURLDocument struct {
	Web string `bson:"web" json:"web"`
}

func GoodWebConfigDocumentFromDomain(config domain.AntiScamConfig) GoodWebConfigDocument {
	return GoodWebConfigDocument{
		Guild: config.GuildID,
		Open:  config.Open,
	}
}

func (d GoodWebConfigDocument) ToDomain() domain.AntiScamConfig {
	return domain.AntiScamConfig{
		GuildID: d.Guild,
		Open:    d.Open,
	}
}

func (d GoodWebConfigReadDocument) ToDomain() domain.AntiScamConfig {
	return domain.AntiScamConfig{
		GuildID: d.Guild,
		Open:    legacyMongooseBoolean(d.Open),
	}
}
