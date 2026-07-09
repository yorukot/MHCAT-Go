package economy

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

const CoinRankPageSize = 10

type CoinRankService struct {
	Repository ports.EconomyCoinRankRepository
}

type CoinRankQuery struct {
	GuildID  string
	ViewerID string
	Page     int
}

type CoinRankEntry struct {
	Rank    int
	Balance domain.CoinBalance
}

type CoinRankPage struct {
	GuildID          string
	ViewerID         string
	Page             int
	PageSize         int
	TotalEntries     int
	TotalPages       int
	Entries          []CoinRankEntry
	ViewerRank       int
	ViewerHasBalance bool
}

func (s CoinRankService) Query(ctx context.Context, query CoinRankQuery) (CoinRankPage, error) {
	if err := ctx.Err(); err != nil {
		return CoinRankPage{}, err
	}
	if s.Repository == nil {
		return CoinRankPage{}, domain.ErrInvalidCoinRankQuery
	}
	query.GuildID = strings.TrimSpace(query.GuildID)
	query.ViewerID = strings.TrimSpace(query.ViewerID)
	if query.GuildID == "" || query.ViewerID == "" {
		return CoinRankPage{}, domain.ErrInvalidCoinRankQuery
	}
	if query.Page < 0 {
		query.Page = 0
	}
	balances, err := s.Repository.ListCoinBalances(ctx, query.GuildID)
	if err != nil {
		return CoinRankPage{}, err
	}
	sort.SliceStable(balances, func(i, j int) bool {
		return balances[i].Coins > balances[j].Coins
	})

	totalPages := 0
	if len(balances) > 0 {
		totalPages = (len(balances) + CoinRankPageSize - 1) / CoinRankPageSize
	}
	entries := []CoinRankEntry{}
	start := query.Page * CoinRankPageSize
	if start < len(balances) {
		end := start + CoinRankPageSize
		if end > len(balances) {
			end = len(balances)
		}
		for index, balance := range balances[start:end] {
			entries = append(entries, CoinRankEntry{
				Rank:    start + index + 1,
				Balance: balance,
			})
		}
	}

	viewerRank, viewerHasBalance := legacyCoinRankForViewer(balances, query.ViewerID)
	return CoinRankPage{
		GuildID:          query.GuildID,
		ViewerID:         query.ViewerID,
		Page:             query.Page,
		PageSize:         CoinRankPageSize,
		TotalEntries:     len(balances),
		TotalPages:       totalPages,
		Entries:          entries,
		ViewerRank:       viewerRank,
		ViewerHasBalance: viewerHasBalance,
	}, nil
}

func legacyCoinRankForViewer(balances []domain.CoinBalance, viewerID string) (int, bool) {
	var viewerCoins int64
	found := false
	for _, balance := range balances {
		if balance.UserID == viewerID {
			viewerCoins = balance.Coins
			found = true
			break
		}
	}
	if !found {
		return 0, false
	}
	rank := 0
	for _, balance := range balances {
		if balance.Coins >= viewerCoins {
			rank++
		}
	}
	return rank, true
}

func LegacyCoinRankAmount(value int64) string {
	switch {
	case value >= 1_000_000_000:
		return legacyCoinRankScaled(value, 1_000_000_000, "G")
	case value >= 1_000_000:
		return legacyCoinRankScaled(value, 1_000_000, "M")
	case value >= 1_000:
		return legacyCoinRankScaled(value, 1_000, "K")
	default:
		return strconv.FormatInt(value, 10)
	}
}

func legacyCoinRankScaled(value int64, divisor int64, suffix string) string {
	text := fmt.Sprintf("%.1f", float64(value)/float64(divisor))
	text = strings.TrimSuffix(text, ".0")
	return text + suffix
}
