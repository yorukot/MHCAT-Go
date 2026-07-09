package economy

import (
	"crypto/rand"
	"math/big"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreeconomy "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/economy"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

type Module struct {
	query        coreeconomy.CoinQueryService
	signIn       coreeconomy.SignInService
	signInList   coreeconomy.SignInListService
	settings     coreeconomy.SettingsService
	coinAdmin    coreeconomy.CoinAdminService
	coinRank     coreeconomy.CoinRankService
	rps          coreeconomy.RockPaperScissorsService
	profile      coreeconomy.ProfileService
	discord      ports.DiscordInfoProvider
	usage        ports.UsageTracker
	clock        ports.Clock
	color        func() int
	rpsChoice    func() domain.RockPaperScissorsChoice
	defs         []commands.Definition
	feature      string
	queryEnabled bool
}

func NewModule(repo ports.EconomyQueryRepository, discordInfo ports.DiscordInfoProvider, usage ports.UsageTracker) Module {
	return Module{
		query:        coreeconomy.CoinQueryService{Repository: repo},
		discord:      discordInfo,
		usage:        usage,
		color:        legacyRandomColor,
		rpsChoice:    legacyRandomRockPaperScissorsChoice,
		defs:         Definitions(),
		feature:      "economy-query",
		queryEnabled: true,
	}
}

func NewModuleWithSignIn(queryRepo ports.EconomyQueryRepository, signInRepo ports.EconomySignInRepository, discordInfo ports.DiscordInfoProvider, clock ports.Clock, usage ports.UsageTracker) Module {
	module := NewModule(queryRepo, discordInfo, usage)
	module.signIn = coreeconomy.SignInService{Repository: signInRepo}
	module.signInList = coreeconomy.SignInListService{Repository: signInRepo}
	module.clock = clock
	module.defs = append(module.defs, SignInDefinitions()...)
	module.feature = "economy"
	return module
}

func NewSignInOnlyModule(repo ports.EconomySignInRepository, discordInfo ports.DiscordInfoProvider, clock ports.Clock, usage ports.UsageTracker) Module {
	return Module{
		query:  coreeconomy.CoinQueryService{Repository: repo},
		signIn: coreeconomy.SignInService{Repository: repo},
		signInList: coreeconomy.SignInListService{
			Repository: repo,
		},
		discord:   discordInfo,
		usage:     usage,
		clock:     clock,
		color:     legacyRandomColor,
		rpsChoice: legacyRandomRockPaperScissorsChoice,
		defs:      SignInDefinitions(),
		feature:   "economy-signin",
	}
}

func NewSignInModule(repo ports.EconomySignInRepository, discordInfo ports.DiscordInfoProvider, clock ports.Clock, usage ports.UsageTracker) Module {
	return NewModuleWithSignIn(repo, repo, discordInfo, clock, usage)
}

func NewSettingsModule(repo ports.EconomySettingsRepository, discordInfo ports.DiscordInfoProvider, usage ports.UsageTracker) Module {
	return Module{
		settings:  coreeconomy.SettingsService{Repository: repo},
		discord:   discordInfo,
		usage:     usage,
		color:     legacyRandomColor,
		rpsChoice: legacyRandomRockPaperScissorsChoice,
		defs:      SettingsDefinitions(),
		feature:   "economy-settings",
	}
}

func NewCoinAdminModule(repo ports.EconomyCoinAdminRepository, discordInfo ports.DiscordInfoProvider, usage ports.UsageTracker) Module {
	return Module{
		coinAdmin: coreeconomy.CoinAdminService{Repository: repo},
		discord:   discordInfo,
		usage:     usage,
		color:     legacyRandomColor,
		rpsChoice: legacyRandomRockPaperScissorsChoice,
		defs:      CoinAdminDefinitions(),
		feature:   "economy-coin-admin",
	}
}

func NewCoinRankModule(repo ports.EconomyCoinRankRepository, discordInfo ports.DiscordInfoProvider, usage ports.UsageTracker) Module {
	return Module{
		coinRank:  coreeconomy.CoinRankService{Repository: repo},
		discord:   discordInfo,
		usage:     usage,
		color:     legacyRandomColor,
		rpsChoice: legacyRandomRockPaperScissorsChoice,
		defs:      CoinRankDefinitions(),
		feature:   "economy-coin-rank",
	}
}

func NewRockPaperScissorsModule(repo ports.EconomyRockPaperScissorsRepository, discordInfo ports.DiscordInfoProvider, usage ports.UsageTracker) Module {
	return Module{
		rps:       coreeconomy.RockPaperScissorsService{Repository: repo},
		discord:   discordInfo,
		usage:     usage,
		color:     legacyRandomColor,
		rpsChoice: legacyRandomRockPaperScissorsChoice,
		defs:      RockPaperScissorsDefinitions(),
		feature:   "economy-rps",
	}
}

func NewProfileModule(repo ports.EconomyProfileRepository, discordInfo ports.DiscordInfoProvider, clock ports.Clock, usage ports.UsageTracker) Module {
	return Module{
		profile:   coreeconomy.ProfileService{Repository: repo},
		discord:   discordInfo,
		usage:     usage,
		clock:     clock,
		color:     legacyRandomColor,
		rpsChoice: legacyRandomRockPaperScissorsChoice,
		defs:      ProfileDefinitions(),
		feature:   "economy-profile",
	}
}

func (m Module) Name() string {
	return m.feature
}

func (m Module) Commands() []commands.Definition {
	return append([]commands.Definition(nil), m.defs...)
}

func (m Module) RegisterRoutes(router *interactions.Router) error {
	if m.queryEnabled {
		if err := router.RegisterSlash("代幣查詢", m.CoinQueryHandler()); err != nil {
			return err
		}
	}
	if m.signIn.Repository != nil {
		if err := router.RegisterSlash("簽到", m.SignInHandler()); err != nil {
			return err
		}
		if err := router.RegisterSlash(SignInListCommandName, m.SignInListHandler()); err != nil {
			return err
		}
		for _, key := range []interactions.RouteKey{
			{Kind: interactions.TypeComponent, Version: "v1", Feature: "economy", Action: "sign_page"},
			{Kind: interactions.TypeComponent, Version: "legacy", Feature: "economy", Action: "sign_page", Legacy: true},
		} {
			if err := router.RegisterRoute(key, m.SignPageHandler()); err != nil {
				return err
			}
		}
	}
	if m.settings.Repository != nil {
		if err := router.RegisterSlash(EconomySettingsCommandName, m.SettingsHandler()); err != nil {
			return err
		}
	}
	if m.coinAdmin.Repository != nil {
		if err := router.RegisterSlash(CoinAdminCommandName, m.CoinAdminHandler()); err != nil {
			return err
		}
	}
	if m.coinRank.Repository != nil {
		if err := router.RegisterSlash(CoinRankCommandName, m.CoinRankHandler()); err != nil {
			return err
		}
		if err := router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeComponent, Version: "legacy", Feature: "rank", Action: "coin_page", Legacy: true}, m.CoinRankPageHandler()); err != nil {
			return err
		}
	}
	if m.rps.Repository != nil {
		if err := router.RegisterSlash(RockPaperScissorsCommandName, m.RockPaperScissorsHandler()); err != nil {
			return err
		}
	}
	if m.profile.Repository != nil {
		if err := router.RegisterSlash(ProfileCommandName, m.ProfileHandler()); err != nil {
			return err
		}
		if err := router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeComponent, Version: "legacy", Feature: "profile", Action: "refresh", Legacy: true}, m.ProfileRefreshHandler()); err != nil {
			return err
		}
	}
	return nil
}

func legacyRandomColor() int {
	max := big.NewInt(0xFFFFFF + 1)
	value, err := rand.Int(rand.Reader, max)
	if err != nil {
		return coinQuerySuccessColor
	}
	return int(value.Int64())
}

func legacyRandomRockPaperScissorsChoice() domain.RockPaperScissorsChoice {
	choices := []domain.RockPaperScissorsChoice{
		domain.RockPaperScissorsChoiceScissors,
		domain.RockPaperScissorsChoiceRock,
		domain.RockPaperScissorsChoicePaper,
	}
	max := big.NewInt(int64(len(choices)))
	value, err := rand.Int(rand.Reader, max)
	if err != nil {
		return domain.RockPaperScissorsChoiceScissors
	}
	return choices[value.Int64()]
}
