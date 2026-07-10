package birthday

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestConfigServiceSavesValidConfig(t *testing.T) {
	repo := &fakemongo.BirthdayConfigRepository{}
	service := NewConfigService(repo)

	if err := service.Save(context.Background(), domain.BirthdayConfig{
		GuildID:                    " guild-1 ",
		Message:                    " {user} 生日快樂 ",
		UTCOffset:                  " +08:00 ",
		ChannelID:                  " channel-1 ",
		EveryoneCanSetBirthdayDate: true,
		RoleID:                     " role-1 ",
	}); err != nil {
		t.Fatalf("save: %v", err)
	}

	saved, ok := repo.Last()
	if !ok {
		t.Fatal("expected saved config")
	}
	if saved.GuildID != "guild-1" || saved.Message != " {user} 生日快樂 " || saved.UTCOffset != "+08:00" || saved.ChannelID != "channel-1" || saved.RoleID != "role-1" {
		t.Fatalf("saved = %#v", saved)
	}
}

func TestConfigServiceRejectsInvalidConfig(t *testing.T) {
	service := NewConfigService(&fakemongo.BirthdayConfigRepository{})

	err := service.Save(context.Background(), domain.BirthdayConfig{GuildID: "guild-1", Message: "hi", UTCOffset: "-01:00", ChannelID: "channel-1"})
	if !errors.Is(err, domain.ErrInvalidBirthdayConfig) {
		t.Fatalf("err = %v", err)
	}
}

func TestConfigServiceSavesWhitespaceMessage(t *testing.T) {
	repo := &fakemongo.BirthdayConfigRepository{}
	service := NewConfigService(repo)

	if err := service.Save(context.Background(), domain.BirthdayConfig{
		GuildID:   "guild-1",
		Message:   "   ",
		UTCOffset: "+08:00",
		ChannelID: "channel-1",
	}); err != nil {
		t.Fatalf("save: %v", err)
	}
	saved, ok := repo.Last()
	if !ok || saved.Message != "   " {
		t.Fatalf("saved = %#v", saved)
	}
}
