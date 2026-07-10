package xp

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestTextCoinRewardServiceAppliesLegacyLevelReward(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutConfig(domain.EconomyConfig{GuildID: "guild-1", XPMultiple: 2.5})
	service := TextCoinRewardService{Repository: repo}

	balance, err := service.ApplyLevelUp(context.Background(), " guild-1 ", " user-1 ", 3)
	if err != nil {
		t.Fatalf("apply reward: %v", err)
	}
	if balance.GuildID != "guild-1" || balance.UserID != "user-1" || balance.Coins != 7 || balance.Today != 0 {
		t.Fatalf("balance = %#v", balance)
	}
}

func TestVoiceCoinRewardServiceAppliesLegacyLevelReward(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutConfig(domain.EconomyConfig{GuildID: "guild-1", XPMultiple: 2.5})
	service := VoiceCoinRewardService{Repository: repo}

	balance, err := service.ApplyLevelUp(context.Background(), " guild-1 ", " user-1 ", 3)
	if err != nil {
		t.Fatalf("apply reward: %v", err)
	}
	if balance.GuildID != "guild-1" || balance.UserID != "user-1" || balance.Coins != 7 || balance.Today != 0 {
		t.Fatalf("balance = %#v", balance)
	}
}

func TestTextCoinRewardServiceRejectsInvalidInput(t *testing.T) {
	service := TextCoinRewardService{Repository: fakemongo.NewEconomyRepository()}
	if _, err := service.ApplyLevelUp(context.Background(), "", "user-1", 1); !errors.Is(err, domain.ErrInvalidXPAdjustment) {
		t.Fatalf("expected invalid adjustment, got %v", err)
	}
	if _, err := (TextCoinRewardService{}).ApplyLevelUp(context.Background(), "guild-1", "user-1", 1); !errors.Is(err, domain.ErrInvalidXPAdjustment) {
		t.Fatalf("expected invalid service, got %v", err)
	}
}

func TestVoiceCoinRewardServiceRejectsInvalidInput(t *testing.T) {
	service := VoiceCoinRewardService{Repository: fakemongo.NewEconomyRepository()}
	if _, err := service.ApplyLevelUp(context.Background(), "", "user-1", 1); !errors.Is(err, domain.ErrInvalidXPAdjustment) {
		t.Fatalf("expected invalid adjustment, got %v", err)
	}
	if _, err := (VoiceCoinRewardService{}).ApplyLevelUp(context.Background(), "guild-1", "user-1", 1); !errors.Is(err, domain.ErrInvalidXPAdjustment) {
		t.Fatalf("expected invalid service, got %v", err)
	}
}

func TestLegacyTextXPCoinRewardTruncatesFractionalRewards(t *testing.T) {
	if got := domain.LegacyTextXPCoinReward(3, 2.5); got != 7 {
		t.Fatalf("reward = %d", got)
	}
}
