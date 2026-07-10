package safety

import (
	"context"
	"errors"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type MessageService struct {
	configs ports.AntiScamConfigRepository
	catalog ports.ScamURLCatalog
}

type MessageScanResult struct {
	MatchedURL string
	Delete     bool
}

func NewMessageService(configs ports.AntiScamConfigRepository, catalog ports.ScamURLCatalog) MessageService {
	return MessageService{configs: configs, catalog: catalog}
}

func (s MessageService) Scan(ctx context.Context, guildID string, content string) (MessageScanResult, error) {
	if err := ctx.Err(); err != nil {
		return MessageScanResult{}, err
	}
	if s.configs == nil || s.catalog == nil {
		return MessageScanResult{}, domain.ErrInvalidAntiScamConfig
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return MessageScanResult{}, domain.ErrInvalidAntiScamConfig
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return MessageScanResult{}, ctx.Err()
	}
	config, err := s.configs.FindAntiScamConfig(ctx, guildID)
	if errors.Is(err, ports.ErrAntiScamConfigMissing) {
		return MessageScanResult{}, ctx.Err()
	}
	if err != nil {
		return MessageScanResult{}, err
	}
	if !config.Open {
		return MessageScanResult{}, ctx.Err()
	}
	matched, ok, err := s.catalog.FindScamURLInContent(ctx, content)
	if err != nil {
		return MessageScanResult{}, err
	}
	if !ok {
		return MessageScanResult{}, ctx.Err()
	}
	return MessageScanResult{MatchedURL: matched, Delete: true}, ctx.Err()
}
