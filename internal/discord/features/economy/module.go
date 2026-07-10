package economy

import (
	"context"
	"crypto/rand"
	"log/slog"
	"math/big"
	"sync"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreeconomy "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/economy"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

type Module struct {
	query         coreeconomy.CoinQueryService
	signIn        coreeconomy.SignInService
	signInList    coreeconomy.SignInListService
	settings      coreeconomy.SettingsService
	coinAdmin     coreeconomy.CoinAdminService
	coinReset     coreeconomy.CoinResetService
	coinRank      coreeconomy.CoinRankService
	rps           coreeconomy.RockPaperScissorsService
	game          coreeconomy.CoinGameService
	shop          coreeconomy.ShopService
	profile       coreeconomy.ProfileService
	discord       ports.DiscordInfoProvider
	messages      ports.DiscordMessagePort
	direct        ports.DiscordDirectMessagePort
	roles         ports.DiscordRolePort
	roleInspector ports.DiscordRoleInspector
	usage         ports.UsageTracker
	clock         ports.Clock
	confirmations *coinResetConfirmationStore
	gameSessions  *coinGameSessionStore
	gameTimeouts  coinGameTimeoutScheduler
	shopSessions  *shopSessionStore
	color         func() int
	rpsChoice     func() domain.RockPaperScissorsChoice
	gameRandInt   func(int) int
	logger        *slog.Logger
	defs          []commands.Definition
	feature       string
	queryEnabled  bool
}

func NewModule(repo ports.EconomyQueryRepository, discordInfo ports.DiscordInfoProvider, usage ports.UsageTracker) Module {
	return Module{
		query:        coreeconomy.CoinQueryService{Repository: repo},
		discord:      discordInfo,
		usage:        usage,
		color:        legacyRandomColor,
		rpsChoice:    legacyRandomRockPaperScissorsChoice,
		gameRandInt:  legacyRandomInt,
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
		discord:     discordInfo,
		usage:       usage,
		clock:       clock,
		color:       legacyRandomColor,
		rpsChoice:   legacyRandomRockPaperScissorsChoice,
		gameRandInt: legacyRandomInt,
		defs:        SignInDefinitions(),
		feature:     "economy-signin",
	}
}

func NewSignInModule(repo ports.EconomySignInRepository, discordInfo ports.DiscordInfoProvider, clock ports.Clock, usage ports.UsageTracker) Module {
	return NewModuleWithSignIn(repo, repo, discordInfo, clock, usage)
}

func NewSettingsModule(repo ports.EconomySettingsRepository, discordInfo ports.DiscordInfoProvider, usage ports.UsageTracker) Module {
	return Module{
		settings:    coreeconomy.SettingsService{Repository: repo},
		discord:     discordInfo,
		usage:       usage,
		color:       legacyRandomColor,
		rpsChoice:   legacyRandomRockPaperScissorsChoice,
		gameRandInt: legacyRandomInt,
		defs:        SettingsDefinitions(),
		feature:     "economy-settings",
	}
}

func NewCoinAdminModule(repo ports.EconomyCoinAdminRepository, discordInfo ports.DiscordInfoProvider, usage ports.UsageTracker) Module {
	return Module{
		coinAdmin:   coreeconomy.CoinAdminService{Repository: repo},
		discord:     discordInfo,
		usage:       usage,
		color:       legacyRandomColor,
		rpsChoice:   legacyRandomRockPaperScissorsChoice,
		gameRandInt: legacyRandomInt,
		defs:        CoinAdminDefinitions(),
		feature:     "economy-coin-admin",
	}
}

func NewCoinResetModule(repo ports.EconomyCoinResetRepository, discordInfo ports.DiscordInfoProvider, messages ports.DiscordMessagePort, usage ports.UsageTracker, clock ports.Clock) Module {
	confirmations := defaultCoinResetConfirmationStore
	if clock == nil {
		clock = ports.SystemClock{}
	} else {
		confirmations = newCoinResetConfirmationStore(clock, time.Minute)
	}
	return Module{
		coinReset:     coreeconomy.CoinResetService{Repository: repo},
		discord:       discordInfo,
		messages:      messages,
		usage:         usage,
		clock:         clock,
		confirmations: confirmations,
		color:         legacyRandomColor,
		rpsChoice:     legacyRandomRockPaperScissorsChoice,
		gameRandInt:   legacyRandomInt,
		defs:          CoinResetDefinitions(),
		feature:       "economy-coin-reset",
	}
}

func NewCoinRankModule(repo ports.EconomyCoinRankRepository, discordInfo ports.DiscordInfoProvider, usage ports.UsageTracker) Module {
	return Module{
		coinRank:    coreeconomy.CoinRankService{Repository: repo},
		discord:     discordInfo,
		usage:       usage,
		color:       legacyRandomColor,
		rpsChoice:   legacyRandomRockPaperScissorsChoice,
		gameRandInt: legacyRandomInt,
		defs:        CoinRankDefinitions(),
		feature:     "economy-coin-rank",
	}
}

func NewRockPaperScissorsModule(repo ports.EconomyRockPaperScissorsRepository, discordInfo ports.DiscordInfoProvider, usage ports.UsageTracker) Module {
	return Module{
		rps:         coreeconomy.RockPaperScissorsService{Repository: repo},
		discord:     discordInfo,
		usage:       usage,
		color:       legacyRandomColor,
		rpsChoice:   legacyRandomRockPaperScissorsChoice,
		gameRandInt: legacyRandomInt,
		defs:        RockPaperScissorsDefinitions(),
		feature:     "economy-rps",
	}
}

func NewCoinGameModule(repo ports.EconomyCoinGameRepository, discordInfo ports.DiscordInfoProvider, usage ports.UsageTracker, clock ports.Clock) Module {
	return NewCoinGameModuleWithMessages(repo, discordInfo, nil, usage, clock)
}

func NewCoinGameModuleWithMessages(repo ports.EconomyCoinGameRepository, discordInfo ports.DiscordInfoProvider, messages ports.DiscordMessagePort, usage ports.UsageTracker, clock ports.Clock) Module {
	if clock == nil {
		clock = ports.SystemClock{}
	}
	return Module{
		game:         coreeconomy.CoinGameService{Repository: repo},
		discord:      discordInfo,
		messages:     messages,
		usage:        usage,
		clock:        clock,
		gameSessions: newCoinGameSessionStore(clock),
		gameTimeouts: newCoinGameTimeoutManager(clock),
		color:        legacyRandomColor,
		rpsChoice:    legacyRandomRockPaperScissorsChoice,
		gameRandInt:  legacyRandomInt,
		logger:       slog.Default(),
		defs:         CoinGameDefinitions(),
		feature:      "economy-game",
	}
}

func (m Module) WithLogger(logger *slog.Logger) Module {
	if logger != nil {
		m.logger = logger
	}
	return m
}

func (m Module) StopCoinGameLifecycle(ctx context.Context) error {
	if m.gameTimeouts == nil {
		return nil
	}
	return m.gameTimeouts.Stop(ctx)
}

func NewShopModule(repo ports.EconomyShopRepository, discordInfo ports.DiscordInfoProvider, roleInspector ports.DiscordRoleInspector, roles ports.DiscordRolePort, direct ports.DiscordDirectMessagePort, usage ports.UsageTracker, clock ports.Clock) Module {
	if clock == nil {
		clock = ports.SystemClock{}
	}
	return Module{
		shop:          coreeconomy.ShopService{Repository: repo},
		discord:       discordInfo,
		roleInspector: roleInspector,
		roles:         roles,
		direct:        direct,
		usage:         usage,
		clock:         clock,
		shopSessions:  newShopSessionStore(),
		color:         legacyRandomColor,
		rpsChoice:     legacyRandomRockPaperScissorsChoice,
		gameRandInt:   legacyRandomInt,
		defs:          ShopDefinitions(),
		feature:       "economy-shop",
	}
}

func NewProfileModule(repo ports.EconomyProfileRepository, discordInfo ports.DiscordInfoProvider, clock ports.Clock, usage ports.UsageTracker) Module {
	return Module{
		profile:     coreeconomy.ProfileService{Repository: repo},
		discord:     discordInfo,
		usage:       usage,
		clock:       clock,
		color:       legacyRandomColor,
		rpsChoice:   legacyRandomRockPaperScissorsChoice,
		gameRandInt: legacyRandomInt,
		defs:        ProfileDefinitions(),
		feature:     "economy-profile",
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
	if m.coinReset.Repository != nil {
		if err := router.RegisterSlash(CoinResetCommandName, m.CoinResetHandler()); err != nil {
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
	if m.game.Repository != nil {
		if err := router.RegisterSlash(CoinGameCommandName, m.CoinGameHandler()); err != nil {
			return err
		}
		for _, key := range []interactions.RouteKey{
			{Kind: interactions.TypeComponent, Version: "legacy", Feature: "game", Action: "yes", Legacy: true},
			{Kind: interactions.TypeComponent, Version: "legacy", Feature: "game", Action: "no", Legacy: true},
			{Kind: interactions.TypeComponent, Version: "legacy", Feature: "game", Action: "main_stand", Legacy: true},
			{Kind: interactions.TypeComponent, Version: "legacy", Feature: "game", Action: "main_hit", Legacy: true},
			{Kind: interactions.TypeComponent, Version: "legacy", Feature: "game", Action: "user_stand", Legacy: true},
			{Kind: interactions.TypeComponent, Version: "legacy", Feature: "game", Action: "user_hit", Legacy: true},
			{Kind: interactions.TypeComponent, Version: "legacy", Feature: "game", Action: "show_number", Legacy: true},
			{Kind: interactions.TypeComponent, Version: "legacy", Feature: "game", Action: "teach_21_point", Legacy: true},
			{Kind: interactions.TypeComponent, Version: "legacy", Feature: "game", Action: "than_size_help", Legacy: true},
			{Kind: interactions.TypeComponent, Version: "legacy", Feature: "game", Action: "knowledge_answer", Legacy: true},
		} {
			if err := router.RegisterRoute(key, m.CoinGameComponentHandler()); err != nil {
				return err
			}
		}
	}
	if m.shop.Repository != nil {
		if err := router.RegisterSlash(ShopCommandName, m.ShopHandler()); err != nil {
			return err
		}
		if err := router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeComponent, Version: "legacy", Feature: "shop", Action: "item", Legacy: true}, m.ShopItemHandler()); err != nil {
			return err
		}
		if err := router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeComponent, Version: "legacy", Feature: "shop", Action: "detail", Legacy: true}, m.ShopItemHandler()); err != nil {
			return err
		}
		if err := router.RegisterRoute(interactions.RouteKey{Kind: interactions.TypeComponent, Version: "legacy", Feature: "shop", Action: "quantity", Legacy: true}, m.ShopQuantityHandler()); err != nil {
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

type shopSession struct {
	GuildID     string
	UserID      string
	MessageID   string
	CommodityID int64
	Quantity    string
	ExpiresAt   time.Time
}

type shopSessionStore struct {
	mu             sync.Mutex
	sessions       map[shopSessionKey]shopSession
	browseSessions map[shopBrowseSessionKey]shopBrowseSession
}

type shopSessionKey struct {
	GuildID     string
	UserID      string
	MessageID   string
	CommodityID int64
}

type shopBrowseSession struct {
	GuildID       string
	UserID        string
	InteractionID string
	ExpiresAt     time.Time
}

type shopBrowseSessionKey struct {
	GuildID       string
	UserID        string
	InteractionID string
}

func newShopSessionStore() *shopSessionStore {
	return &shopSessionStore{
		sessions:       map[shopSessionKey]shopSession{},
		browseSessions: map[shopBrowseSessionKey]shopBrowseSession{},
	}
}

func (s *shopSessionStore) Put(session shopSession) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.key()] = session
}

func (s *shopSessionStore) Get(guildID string, userID string, messageID string, commodityID int64) (shopSession, bool) {
	if s == nil {
		return shopSession{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[shopSessionKey{GuildID: guildID, UserID: userID, MessageID: messageID, CommodityID: commodityID}]
	return session, ok
}

func (s *shopSessionStore) GetByMessage(guildID string, userID string, messageID string) (shopSession, bool) {
	if s == nil {
		return shopSession{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, session := range s.sessions {
		if key.GuildID == guildID && key.UserID == userID && key.MessageID == messageID {
			return session, true
		}
	}
	return shopSession{}, false
}

func (s *shopSessionStore) Update(session shopSession) {
	s.Put(session)
}

func (s *shopSessionStore) Delete(session shopSession) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, session.key())
}

func (s *shopSessionStore) PutBrowse(session shopBrowseSession, now time.Time) {
	if s == nil || session.GuildID == "" || session.UserID == "" || session.InteractionID == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneLocked(now)
	s.browseSessions[session.key()] = session
}

func (s *shopSessionStore) ClaimBrowse(guildID string, userID string, interactionID string, now time.Time) bool {
	if s == nil || guildID == "" || userID == "" || interactionID == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneLocked(now)
	key := shopBrowseSessionKey{GuildID: guildID, UserID: userID, InteractionID: interactionID}
	_, ok := s.browseSessions[key]
	delete(s.browseSessions, key)
	return ok
}

func (s *shopSessionStore) pruneLocked(now time.Time) {
	for key, session := range s.sessions {
		if !session.ExpiresAt.IsZero() && !now.Before(session.ExpiresAt) {
			delete(s.sessions, key)
		}
	}
	for key, session := range s.browseSessions {
		if !session.ExpiresAt.IsZero() && !now.Before(session.ExpiresAt) {
			delete(s.browseSessions, key)
		}
	}
}

func (s shopSession) key() shopSessionKey {
	return shopSessionKey{GuildID: s.GuildID, UserID: s.UserID, MessageID: s.MessageID, CommodityID: s.CommodityID}
}

func (s shopBrowseSession) key() shopBrowseSessionKey {
	return shopBrowseSessionKey{GuildID: s.GuildID, UserID: s.UserID, InteractionID: s.InteractionID}
}

func (m Module) RegisterEventRoutes(dispatcher *events.Dispatcher) {
	if dispatcher != nil && m.coinReset.Repository != nil {
		dispatcher.Register(events.TypeMessageCreate, m.CoinResetConfirmationHandler())
	}
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

func legacyRandomInt(maxExclusive int) int {
	if maxExclusive <= 0 {
		return 0
	}
	value, err := rand.Int(rand.Reader, big.NewInt(int64(maxExclusive)))
	if err != nil {
		return 0
	}
	return int(value.Int64())
}
