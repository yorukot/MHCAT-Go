package xp

import (
	"context"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

const RankPageSize = 10

type RankKind string

const (
	RankKindText  RankKind = "text"
	RankKindVoice RankKind = "voice"
)

type RankService struct {
	Repository ports.XPRankRepository
}

type RankQuery struct {
	GuildID  string
	ViewerID string
	Page     int
	Kind     RankKind
}

type RankEntry struct {
	Rank    int
	Profile domain.XPProfile
	TotalXP int64
}

type RankPage struct {
	GuildID          string
	ViewerID         string
	Page             int
	PageSize         int
	TotalEntries     int
	TotalPages       int
	Entries          []RankEntry
	ViewerRank       int
	ViewerHasProfile bool
	Kind             RankKind
}

func (s RankService) Query(ctx context.Context, query RankQuery) (RankPage, error) {
	if err := ctx.Err(); err != nil {
		return RankPage{}, err
	}
	if s.Repository == nil {
		return RankPage{}, domain.ErrInvalidXPRankQuery
	}
	query.GuildID = strings.TrimSpace(query.GuildID)
	query.ViewerID = strings.TrimSpace(query.ViewerID)
	if query.GuildID == "" || query.ViewerID == "" || !validRankKind(query.Kind) {
		return RankPage{}, domain.ErrInvalidXPRankQuery
	}
	if query.Page < 0 {
		query.Page = 0
	}
	profiles, err := s.listProfiles(ctx, query)
	if err != nil {
		return RankPage{}, err
	}
	// Legacy builds the sortable array from the Mongo result in reverse order.
	for left, right := 0, len(profiles)-1; left < right; left, right = left+1, right-1 {
		profiles[left], profiles[right] = profiles[right], profiles[left]
	}
	sort.SliceStable(profiles, func(i, j int) bool {
		return legacyRankSortTotal(profiles[i]) > legacyRankSortTotal(profiles[j])
	})

	totalPages := 0
	if len(profiles) > 0 {
		totalPages = (len(profiles) + RankPageSize - 1) / RankPageSize
	}
	entries := []RankEntry{}
	start := query.Page * RankPageSize
	if start < len(profiles) {
		end := start + RankPageSize
		if end > len(profiles) {
			end = len(profiles)
		}
		for index, profile := range profiles[start:end] {
			entries = append(entries, RankEntry{
				Rank:    start + index + 1,
				Profile: profile,
				TotalXP: legacyRankSortTotal(profile),
			})
		}
	}
	viewerRank, viewerHasProfile := legacyRankForViewer(profiles, query.ViewerID)
	return RankPage{
		GuildID:          query.GuildID,
		ViewerID:         query.ViewerID,
		Page:             query.Page,
		PageSize:         RankPageSize,
		TotalEntries:     len(profiles),
		TotalPages:       totalPages,
		Entries:          entries,
		ViewerRank:       viewerRank,
		ViewerHasProfile: viewerHasProfile,
		Kind:             query.Kind,
	}, nil
}

func (s RankService) listProfiles(ctx context.Context, query RankQuery) ([]domain.XPProfile, error) {
	if query.Kind == RankKindVoice {
		return s.Repository.ListVoiceXPProfiles(ctx, query.GuildID)
	}
	return s.Repository.ListTextXPProfiles(ctx, query.GuildID)
}

func validRankKind(kind RankKind) bool {
	return kind == RankKindText || kind == RankKindVoice
}

func legacyRankForViewer(profiles []domain.XPProfile, viewerID string) (int, bool) {
	var viewer domain.XPProfile
	found := false
	for _, profile := range profiles {
		if profile.UserID == viewerID {
			viewer = profile
			found = true
			break
		}
	}
	if !found {
		return 0, false
	}
	viewerTotal := legacyRankViewerTotal(viewer)
	rank := 0
	for _, profile := range profiles {
		if legacyRankComparableTotal(profile) >= viewerTotal {
			rank++
		}
	}
	return rank, true
}

func legacyRankSortTotal(profile domain.XPProfile) int64 {
	return legacyRankAccumulated(profile.Level, false, false) + 100 + profile.XP
}

func legacyRankComparableTotal(profile domain.XPProfile) int64 {
	return legacyRankAccumulated(profile.Level, true, true) + 100 + profile.XP
}

func legacyRankViewerTotal(profile domain.XPProfile) int64 {
	return legacyRankAccumulated(profile.Level, false, true) + 100 + profile.XP
}

func legacyRankAccumulated(level int64, includeZero bool, includePerLevelBase bool) int64 {
	total := int64(0)
	for y := level - 1; ; y-- {
		if includeZero {
			if y < 0 {
				break
			}
		} else if y <= 0 {
			break
		}
		total += int64(float64(y)*(float64(y)/3)) * 100
		if includePerLevelBase {
			total += 100
		}
	}
	return total
}

func LegacyRankAmount(value int64) string {
	switch {
	case value >= 1_000_000_000:
		return legacyRankScaled(value, 1_000_000_000, "G")
	case value >= 1_000_000:
		return legacyRankScaled(value, 1_000_000, "M")
	case value >= 1_000:
		return legacyRankScaled(value, 1_000, "K")
	default:
		return strconv.FormatInt(value, 10)
	}
}

func legacyRankScaled(value int64, divisor int64, suffix string) string {
	// JavaScript toFixed rounds the exact binary float with ties upward, unlike Go's tie-to-even formatter.
	scaled := new(big.Rat).SetFloat64(float64(value) / float64(divisor))
	scaled.Mul(scaled, big.NewRat(10, 1))
	rounded, remainder := new(big.Int), new(big.Int)
	rounded.QuoRem(scaled.Num(), scaled.Denom(), remainder)
	if new(big.Int).Lsh(new(big.Int).Set(remainder), 1).Cmp(scaled.Denom()) >= 0 {
		rounded.Add(rounded, big.NewInt(1))
	}
	whole, fraction := new(big.Int), new(big.Int)
	whole.QuoRem(rounded, big.NewInt(10), fraction)
	if fraction.Sign() == 0 {
		return whole.String() + suffix
	}
	return whole.String() + "." + fraction.String() + suffix
}
