package documents

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"

type StatsConfigDocument struct {
	Guild  string `bson:"guild" json:"guild"`
	Parent string `bson:"parent" json:"parent"`
}

func (d StatsConfigDocument) ToDomain() domain.StatsConfig {
	return domain.StatsConfig{
		GuildID:  d.Guild,
		ParentID: d.Parent,
	}
}
