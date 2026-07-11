package onboarding

import (
	"context"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestAccountAgeConfigDelete(t *testing.T) {
	repository := fakemongo.NewAccountAgeConfigRepository()
	repository.Configs["guild-1"] = domain.AccountAgeConfig{GuildID: "guild-1", RequiredSeconds: 3600}
	service := AccountAgeConfigService{Repository: repository}
	if err := service.DeleteConfig(context.Background(), "guild-1"); err != nil {
		t.Fatalf("delete config: %v", err)
	}
	if _, exists := repository.Configs["guild-1"]; exists {
		t.Fatal("account age config was not deleted")
	}
}

func TestRandomLeaveMessageColorBounds(t *testing.T) {
	if color := randomLeaveMessageColor(); color < 0 || color > 0xFFFFFF {
		t.Fatalf("leave message color = %#x", color)
	}
}
