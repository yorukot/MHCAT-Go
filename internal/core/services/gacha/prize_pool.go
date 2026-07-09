package gacha

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

const (
	DefaultGachaCost  int64 = 500
	DefaultSignCoins  int64 = 25
	DefaultXPMultiple       = 0
)

type PrizePoolService struct {
	Repository ports.GachaPrizePoolRepository
}

type PrizeDeleteService struct {
	Repository ports.GachaPrizeDeleteRepository
}

type PrizeCreateService struct {
	Repository ports.GachaPrizeCreateRepository
}

type PrizeEditService struct {
	Repository ports.GachaPrizeEditRepository
}

type RandomFloat func() float64

type DrawService struct {
	Repository ports.GachaDrawRepository
	Random     RandomFloat
}

func (s PrizePoolService) Query(ctx context.Context, guildID string) (domain.GachaPrizePool, error) {
	if err := ctx.Err(); err != nil {
		return domain.GachaPrizePool{}, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" || s.Repository == nil {
		return domain.GachaPrizePool{}, domain.ErrInvalidGachaQuery
	}
	prizes, err := s.Repository.ListGachaPrizes(ctx, guildID)
	if err != nil {
		return domain.GachaPrizePool{}, err
	}
	if len(prizes) == 0 {
		return domain.GachaPrizePool{}, ports.ErrGachaPrizePoolEmpty
	}
	config, err := s.Repository.GetGachaConfig(ctx, guildID)
	configFound := true
	if err != nil {
		if !errors.Is(err, ports.ErrEconomyConfigMissing) {
			return domain.GachaPrizePool{}, err
		}
		configFound = false
		config = domain.EconomyConfig{
			GuildID:    guildID,
			GachaCost:  DefaultGachaCost,
			SignCoins:  DefaultSignCoins,
			XPMultiple: DefaultXPMultiple,
		}
	}
	return domain.GachaPrizePool{
		GuildID:     guildID,
		Prizes:      append([]domain.GachaPrize(nil), prizes...),
		Config:      config,
		ConfigFound: configFound,
	}, ctx.Err()
}

func (s PrizeDeleteService) Delete(ctx context.Context, guildID string, prizeName string) (domain.GachaPrize, error) {
	if err := ctx.Err(); err != nil {
		return domain.GachaPrize{}, err
	}
	guildID = strings.TrimSpace(guildID)
	prizeName = strings.TrimSpace(prizeName)
	if guildID == "" || prizeName == "" || s.Repository == nil {
		return domain.GachaPrize{}, domain.ErrInvalidGachaQuery
	}
	return s.Repository.DeleteGachaPrize(ctx, guildID, prizeName)
}

func (s PrizeCreateService) Create(ctx context.Context, prize domain.GachaPrizeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	prize.GuildID = strings.TrimSpace(prize.GuildID)
	prize.Name = strings.TrimSpace(prize.Name)
	if prize.GuildID == "" || prize.Name == "" || prize.Count <= 0 || s.Repository == nil {
		return domain.ErrInvalidGachaPrize
	}
	count, err := s.Repository.CountGachaPrizes(ctx, prize.GuildID)
	if err != nil {
		return err
	}
	if count > 24 {
		return ports.ErrGachaPrizePoolFull
	}
	return s.Repository.CreateGachaPrize(ctx, prize)
}

func (s PrizeEditService) Edit(ctx context.Context, edit domain.GachaPrizeEdit) (domain.GachaPrizeConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.GachaPrizeConfig{}, err
	}
	edit.GuildID = strings.TrimSpace(edit.GuildID)
	edit.Name = strings.TrimSpace(edit.Name)
	if edit.GuildID == "" || edit.Name == "" || edit.Count <= 0 || s.Repository == nil {
		return domain.GachaPrizeConfig{}, domain.ErrInvalidGachaPrize
	}
	return s.Repository.EditGachaPrize(ctx, edit)
}

func (s DrawService) Draw(ctx context.Context, command domain.GachaDrawCommand) (domain.GachaDrawResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.GachaDrawResult{}, err
	}
	command.GuildID = strings.TrimSpace(command.GuildID)
	command.UserID = strings.TrimSpace(command.UserID)
	if command.GuildID == "" || command.UserID == "" || s.Repository == nil {
		return domain.GachaDrawResult{}, domain.ErrInvalidGachaDraw
	}
	paidDraws, actualDraws, err := LegacyDrawCounts(command.Choice)
	if err != nil {
		return domain.GachaDrawResult{}, err
	}
	randomValues := make([]float64, actualDraws)
	random := s.Random
	if random == nil {
		random = secureRandomFloat
	}
	for index := range randomValues {
		randomValues[index] = random()
	}
	return s.Repository.DrawGacha(ctx, domain.GachaDrawRequest{
		GuildID:      command.GuildID,
		UserID:       command.UserID,
		PaidDraws:    paidDraws,
		ActualDraws:  actualDraws,
		RandomValues: randomValues,
	})
}

func LegacyDrawCounts(choice string) (int, int, error) {
	switch strings.TrimSpace(choice) {
	case "":
		return 1, 1, nil
	case "5":
		return 5, 5, nil
	case "11":
		return 10, 11, nil
	case "17":
		return 15, 17, nil
	case "23":
		return 20, 23, nil
	default:
		return 0, 0, domain.ErrInvalidGachaDraw
	}
}

func secureRandomFloat() float64 {
	max := big.NewInt(1 << 53)
	value, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0.5
	}
	return float64(value.Int64()) / float64(1<<53)
}
