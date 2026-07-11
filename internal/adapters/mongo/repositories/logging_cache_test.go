package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

func TestLoggingConfigCacheServesPositiveAndNegativeEntries(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	repository := &LoggingConfigRepository{now: func() time.Time { return now }}
	config := domain.LoggingConfig{
		GuildID: "guild-1", ChannelID: "log-1", MessageUpdate: true, MemberVoiceUpdate: true,
	}
	repository.storeCachedLoggingConfig(config.GuildID, config, true)
	repository.storeCachedLoggingConfig("guild-2", domain.LoggingConfig{}, false)

	got, err := repository.GetLoggingConfig(context.Background(), " guild-1 ")
	if err != nil || got != config {
		t.Fatalf("cached config = %#v, err=%v", got, err)
	}
	if _, err := repository.GetLoggingConfig(context.Background(), "guild-2"); !errors.Is(err, ports.ErrLoggingConfigMissing) {
		t.Fatalf("negative cache error = %v", err)
	}
}

func TestLoggingConfigCacheRejectsEmptyGuild(t *testing.T) {
	repository := &LoggingConfigRepository{now: time.Now}
	if _, err := repository.GetLoggingConfig(context.Background(), " "); !errors.Is(err, domain.ErrInvalidLoggingConfig) {
		t.Fatalf("empty guild error = %v", err)
	}
}
