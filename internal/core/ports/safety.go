package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

var ErrAntiScamConfigMissing = errors.New("anti-scam config is missing")

type AntiScamConfigRepository interface {
	FindAntiScamConfig(ctx context.Context, guildID string) (domain.AntiScamConfig, error)
	SaveAntiScamConfig(ctx context.Context, config domain.AntiScamConfig) error
}

type ScamURLCatalog interface {
	ContainsScamURL(ctx context.Context, rawURL string) (bool, error)
	FindScamURLInContent(ctx context.Context, content string) (string, bool, error)
}

type ScamReportSender interface {
	SendScamURLReport(ctx context.Context, report domain.ScamURLReport) error
}
