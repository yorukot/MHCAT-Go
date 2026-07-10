package economy

import (
	"context"
	"strings"
	"unicode/utf16"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

const MaxLegacyShopItems = 25
const MaxLegacyShopNameLength = 15

func LegacyShopNameLength(name string) int {
	return len(utf16.Encode([]rune(name)))
}

type ShopService struct {
	Repository ports.EconomyShopRepository
}

func (s ShopService) List(ctx context.Context, guildID string) ([]domain.ShopItem, error) {
	if s.Repository == nil {
		return nil, ports.ErrShopItemMissing
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return nil, domain.ErrInvalidShopItem
	}
	items, err := s.Repository.ListShopItems(ctx, guildID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, ports.ErrShopItemMissing
	}
	return items, nil
}

func (s ShopService) Detail(ctx context.Context, guildID string, commodityID int64) (domain.ShopItem, error) {
	if s.Repository == nil {
		return domain.ShopItem{}, ports.ErrShopItemMissing
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" || commodityID <= 0 {
		return domain.ShopItem{}, domain.ErrInvalidShopItem
	}
	return s.Repository.GetShopItem(ctx, guildID, commodityID)
}

func (s ShopService) Create(ctx context.Context, item domain.ShopItem) (domain.ShopItem, error) {
	if s.Repository == nil {
		return domain.ShopItem{}, ports.ErrShopItemMissing
	}
	item = item.Normalize()
	if err := item.Validate(); err != nil {
		return domain.ShopItem{}, err
	}
	if LegacyShopNameLength(item.Name) > MaxLegacyShopNameLength {
		return domain.ShopItem{}, domain.ErrInvalidShopItem
	}
	items, err := s.Repository.ListShopItems(ctx, item.GuildID)
	if err != nil {
		return domain.ShopItem{}, err
	}
	if len(items) >= MaxLegacyShopItems {
		return domain.ShopItem{}, ports.ErrShopItemLimit
	}
	return s.Repository.CreateShopItem(ctx, item)
}

func (s ShopService) Delete(ctx context.Context, guildID string, commodityID int64) (domain.ShopItem, error) {
	if s.Repository == nil {
		return domain.ShopItem{}, ports.ErrShopItemMissing
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" || commodityID <= 0 {
		return domain.ShopItem{}, domain.ErrInvalidShopItem
	}
	return s.Repository.DeleteShopItem(ctx, guildID, commodityID)
}

func (s ShopService) Purchase(ctx context.Context, command domain.ShopPurchaseCommand) (domain.ShopPurchaseResult, error) {
	if s.Repository == nil {
		return domain.ShopPurchaseResult{}, ports.ErrShopItemMissing
	}
	command = command.Normalize()
	if err := command.Validate(); err != nil {
		return domain.ShopPurchaseResult{}, err
	}
	return s.Repository.PurchaseShopItem(ctx, command)
}
