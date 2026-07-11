package economy

import (
	"context"
	"fmt"
	"math"
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
	viewerBalance, viewerHasBalance := legacyCoinRankViewerBalance(balances, query.ViewerID)
	for left, right := 0, len(balances)-1; left < right; left, right = left+1, right-1 {
		balances[left], balances[right] = balances[right], balances[left]
	}
	sort.SliceStable(balances, func(i, j int) bool {
		left, leftOK := legacyCoinRankNumber(balances[i])
		right, rightOK := legacyCoinRankNumber(balances[j])
		return leftOK && rightOK && left > right
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

	viewerRank := legacyCoinRankForViewer(balances, viewerBalance, viewerHasBalance)
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

func legacyCoinRankViewerBalance(balances []domain.CoinBalance, viewerID string) (domain.CoinBalance, bool) {
	for _, balance := range balances {
		if balance.UserID == viewerID {
			return balance, true
		}
	}
	return domain.CoinBalance{}, false
}

func legacyCoinRankForViewer(balances []domain.CoinBalance, viewer domain.CoinBalance, found bool) int {
	if !found {
		return 0
	}
	viewerCoins, viewerNumeric := legacyCoinRankNumber(viewer)
	rank := 0
	for _, balance := range balances {
		coins, numeric := legacyCoinRankNumber(balance)
		if !numeric || !viewerNumeric || !(coins < viewerCoins) {
			rank++
		}
	}
	return rank
}

func legacyCoinRankNumber(balance domain.CoinBalance) (float64, bool) {
	text := balance.CoinsText
	if text == "" {
		text = strconv.FormatInt(balance.Coins, 10)
	}
	return legacyDisplayedNumber(text)
}

func LegacyCoinRankAmount(value int64) string {
	return LegacyCoinRankAmountText(strconv.FormatInt(value, 10))
}

func LegacyCoinRankAmountText(value string) string {
	number, numeric := legacyDisplayedNumber(value)
	if !numeric {
		return value
	}
	switch {
	case number >= 1_000_000_000:
		return legacyCoinRankScaled(number, 1_000_000_000, "G")
	case number >= 1_000_000:
		return legacyCoinRankScaled(number, 1_000_000, "M")
	case number >= 1_000:
		return legacyCoinRankScaled(number, 1_000, "K")
	default:
		return value
	}
}

func legacyCoinRankScaled(value float64, divisor float64, suffix string) string {
	if math.IsInf(value, 1) {
		return "Infinity" + suffix
	}
	rounded := math.Floor((value/divisor)*10+0.5) / 10
	text := fmt.Sprintf("%.1f", rounded)
	text = strings.TrimSuffix(text, ".0")
	return text + suffix
}
