package economy

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

func TestShopServiceRejectsOverfullLegacyShop(t *testing.T) {
	repo := &shopServiceRepo{}
	for i := 0; i < MaxLegacyShopItems; i++ {
		repo.items = append(repo.items, domain.ShopItem{GuildID: "guild-1", CommodityID: int64(i + 1), Name: fmt.Sprintf("item-%d", i)})
	}
	service := ShopService{Repository: repo}

	_, err := service.Create(context.Background(), domain.ShopItem{
		GuildID:     "guild-1",
		CommodityID: 26,
		Name:        "new",
		NeedCoins:   10,
		Description: "desc",
		Count:       1,
	})
	if !errors.Is(err, ports.ErrShopItemLimit) {
		t.Fatalf("expected ErrShopItemLimit, got %v", err)
	}
}

func TestShopServiceRejectsInvalidItem(t *testing.T) {
	service := ShopService{Repository: &shopServiceRepo{}}

	_, err := service.Create(context.Background(), domain.ShopItem{
		GuildID:     "guild-1",
		CommodityID: 1,
		Name:        "這個商品名稱真的已經超過十五個字",
		NeedCoins:   10,
		Description: "desc",
		Count:       1,
	})
	if !errors.Is(err, domain.ErrInvalidShopItem) {
		t.Fatalf("expected ErrInvalidShopItem, got %v", err)
	}
}

func TestShopServiceUsesLegacyUTF16NameLength(t *testing.T) {
	service := ShopService{Repository: &shopServiceRepo{}}
	item := domain.ShopItem{
		GuildID:     "guild-1",
		CommodityID: 1,
		Name:        strings.Repeat("\U0001F600", 8),
		NeedCoins:   10,
		Description: "desc",
		Count:       1,
	}

	if _, err := service.Create(context.Background(), item); !errors.Is(err, domain.ErrInvalidShopItem) {
		t.Fatalf("expected eight emoji to exceed 15 UTF-16 units, got %v", err)
	}
	item.Name = strings.Repeat("\U0001F600", 7) + "a"
	if _, err := service.Create(context.Background(), item); err != nil {
		t.Fatalf("expected 15 UTF-16 units to be accepted, got %v", err)
	}
}

type shopServiceRepo struct {
	items []domain.ShopItem
}

func (r *shopServiceRepo) ListShopItems(ctx context.Context, guildID string) ([]domain.ShopItem, error) {
	return append([]domain.ShopItem(nil), r.items...), ctx.Err()
}

func (r *shopServiceRepo) GetShopItem(ctx context.Context, guildID string, commodityID int64) (domain.ShopItem, error) {
	return domain.ShopItem{}, ports.ErrShopItemMissing
}

func (r *shopServiceRepo) CreateShopItem(ctx context.Context, item domain.ShopItem) (domain.ShopItem, error) {
	r.items = append(r.items, item)
	return item, ctx.Err()
}

func (r *shopServiceRepo) DeleteShopItem(ctx context.Context, guildID string, commodityID int64) (domain.ShopItem, error) {
	return domain.ShopItem{}, ports.ErrShopItemMissing
}

func (r *shopServiceRepo) PurchaseShopItem(ctx context.Context, command domain.ShopPurchaseCommand) (domain.ShopPurchaseResult, error) {
	return domain.ShopPurchaseResult{}, ports.ErrShopItemMissing
}
