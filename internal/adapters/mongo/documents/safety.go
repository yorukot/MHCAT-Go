package documents

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"

type GoodWebConfigDocument struct {
	Guild string `bson:"guild" json:"guild"`
	Open  bool   `bson:"open" json:"open"`
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
