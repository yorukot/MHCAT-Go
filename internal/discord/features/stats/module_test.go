package stats

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestModuleRegistersStatsRoute(t *testing.T) {
	module := NewModule(nil)
	if module.Name() != "stats-query" {
		t.Fatalf("name = %q", module.Name())
	}
	if len(module.Commands()) != 1 || module.Commands()[0].Name != StatsQueryCommandName {
		t.Fatalf("commands = %#v", module.Commands())
	}
	router := interactions.NewRouter()
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register routes: %v", err)
	}
}

func TestModuleRegistersStatsDeleteRoute(t *testing.T) {
	module := NewDeleteModule(fakemongo.NewStatsConfigRepository(), nil)
	if module.Name() != "stats-delete" {
		t.Fatalf("name = %q", module.Name())
	}
	if len(module.Commands()) != 1 || module.Commands()[0].Name != StatsDeleteCommandName {
		t.Fatalf("commands = %#v", module.Commands())
	}
	router := interactions.NewRouter()
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register routes: %v", err)
	}
}

func TestModuleRegistersStatsCreateRoute(t *testing.T) {
	discord := fakediscord.NewSideEffects()
	module := NewCreateModule(fakemongo.NewStatsConfigRepository(), discord, discord, nil, "bot-1")
	if module.Name() != "stats-create" {
		t.Fatalf("name = %q", module.Name())
	}
	if len(module.Commands()) != 1 || module.Commands()[0].Name != StatsCreateCommandName {
		t.Fatalf("commands = %#v", module.Commands())
	}
	router := interactions.NewRouter()
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register routes: %v", err)
	}
}
