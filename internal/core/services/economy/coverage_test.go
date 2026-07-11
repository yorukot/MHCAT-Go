package economy

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

func TestLegacyXPAccumulatedIntegerCompatibility(t *testing.T) {
	if got := legacyXPAccumulated(3, 3, true); got != 400 {
		t.Fatalf("legacy accumulated xp = %d", got)
	}
}

func TestShopServiceDeleteValidatesAndDelegates(t *testing.T) {
	service := ShopService{Repository: &shopServiceRepo{}}
	if _, err := service.Delete(context.Background(), "", 1); !errors.Is(err, domain.ErrInvalidShopItem) {
		t.Fatalf("invalid delete: %v", err)
	}
	if _, err := service.Delete(context.Background(), " guild-1 ", 1); !errors.Is(err, ports.ErrShopItemMissing) {
		t.Fatalf("delegated delete: %v", err)
	}
}
