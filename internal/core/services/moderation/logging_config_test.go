package moderation

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestLoggingConfigServiceSavesValidConfig(t *testing.T) {
	repo := &fakemongo.LoggingConfigRepository{}
	service := NewLoggingConfigService(repo)
	config := domain.LoggingConfig{GuildID: "guild-1", ChannelID: "channel-1", MessageDelete: true}

	if err := service.Save(context.Background(), config); err != nil {
		t.Fatalf("save: %v", err)
	}
	saved, ok := repo.Last()
	if !ok || saved.GuildID != config.GuildID || !saved.MessageDelete {
		t.Fatalf("saved = %#v ok=%v", saved, ok)
	}
}

func TestLoggingConfigServiceRejectsInvalidConfig(t *testing.T) {
	service := NewLoggingConfigService(&fakemongo.LoggingConfigRepository{})

	if err := service.Save(context.Background(), domain.LoggingConfig{GuildID: "guild-1"}); !errors.Is(err, domain.ErrInvalidLoggingConfig) {
		t.Fatalf("err = %v", err)
	}
}
