package stats

import (
	"crypto/rand"
	"math/big"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

const legacyRandomColorFallback = 0x5865F2

type Module struct {
	usage ports.UsageTracker
	color func() int
}

func NewModule(usage ports.UsageTracker) Module {
	return NewModuleWithColor(usage, legacyRandomColor)
}

func NewModuleWithColor(usage ports.UsageTracker, color func() int) Module {
	if color == nil {
		color = func() int { return legacyRandomColorFallback }
	}
	return Module{usage: usage, color: color}
}

func (m Module) Name() string {
	return "stats-query"
}

func (m Module) Commands() []commands.Definition {
	return Definitions()
}

func (m Module) RegisterRoutes(router *interactions.Router) error {
	return router.RegisterSlash(StatsQueryCommandName, m.QueryHandler())
}

func legacyRandomColor() int {
	max := big.NewInt(0xFFFFFF + 1)
	value, err := rand.Int(rand.Reader, max)
	if err != nil {
		return legacyRandomColorFallback
	}
	return int(value.Int64())
}
