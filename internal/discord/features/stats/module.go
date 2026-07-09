package stats

import (
	"crypto/rand"
	"math/big"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	corestats "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/stats"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

const legacyRandomColorFallback = 0x5865F2

type Module struct {
	service       corestats.ConfigService
	createService corestats.CreateService
	usage         ports.UsageTracker
	color         func() int
	defs          []commands.Definition
	feature       string
	queryEnabled  bool
	createEnabled bool
	deleteEnabled bool
	botUserID     string
}

func NewModule(usage ports.UsageTracker) Module {
	return NewModuleWithColor(usage, legacyRandomColor)
}

func NewModuleWithColor(usage ports.UsageTracker, color func() int) Module {
	if color == nil {
		color = func() int { return legacyRandomColorFallback }
	}
	return Module{
		usage:        usage,
		color:        color,
		defs:         QueryDefinitions(),
		feature:      "stats-query",
		queryEnabled: true,
	}
}

func NewDeleteModule(repo ports.StatsConfigRepository, usage ports.UsageTracker) Module {
	return Module{
		service:       corestats.ConfigService{Repository: repo},
		usage:         usage,
		color:         func() int { return legacyRandomColorFallback },
		defs:          DeleteDefinitions(),
		feature:       "stats-delete",
		deleteEnabled: true,
	}
}

func NewCreateModule(repo ports.StatsConfigRepository, channels ports.DiscordChannelPort, guildStats ports.DiscordGuildStatsReader, usage ports.UsageTracker, botUserID string) Module {
	return Module{
		createService: corestats.CreateService{
			Repository: repo,
			Channels:   channels,
			GuildStats: guildStats,
		},
		usage:         usage,
		color:         func() int { return legacyRandomColorFallback },
		defs:          CreateDefinitions(),
		feature:       "stats-create",
		createEnabled: true,
		botUserID:     botUserID,
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
		if err := router.RegisterSlash(StatsQueryCommandName, m.QueryHandler()); err != nil {
			return err
		}
	}
	if m.createEnabled {
		if err := router.RegisterSlash(StatsCreateCommandName, m.CreateHandler()); err != nil {
			return err
		}
	}
	if m.deleteEnabled {
		if err := router.RegisterSlash(StatsDeleteCommandName, m.DeleteHandler()); err != nil {
			return err
		}
	}
	return nil
}

func legacyRandomColor() int {
	max := big.NewInt(0xFFFFFF + 1)
	value, err := rand.Int(rand.Reader, max)
	if err != nil {
		return legacyRandomColorFallback
	}
	return int(value.Int64())
}
