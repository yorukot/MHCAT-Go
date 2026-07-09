package features_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/features"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

func TestEmptyRegistryValid(t *testing.T) {
	registry, err := features.NewRegistry()
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	commandsRegistry, err := registry.CommandRegistry(commands.Scope{Kind: commands.ScopeGlobal})
	if err != nil {
		t.Fatalf("command registry: %v", err)
	}
	if len(commandsRegistry.Commands) != 0 {
		t.Fatalf("commands = %#v", commandsRegistry.Commands)
	}
}

func TestRegistrySortsModulesDeterministically(t *testing.T) {
	registry, err := features.NewRegistry(
		fakeModule{name: "z", commands: []commands.Definition{{Name: "zeta", Description: "zeta"}}},
		fakeModule{name: "a", commands: []commands.Definition{{Name: "alpha", Description: "alpha"}}},
	)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	var names []string
	for _, module := range registry.Modules() {
		names = append(names, module.Name())
	}
	if !reflect.DeepEqual(names, []string{"a", "z"}) {
		t.Fatalf("module order = %#v", names)
	}
}

func TestRegistryDuplicateCommandsFailValidation(t *testing.T) {
	registry, err := features.NewRegistry(
		fakeModule{name: "one", commands: []commands.Definition{{Name: "ping", Description: "one"}}},
		fakeModule{name: "two", commands: []commands.Definition{{Name: "ping", Description: "two"}}},
	)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	_, err = registry.CommandRegistry(commands.Scope{Kind: commands.ScopeGlobal})
	if !errors.Is(err, commands.ErrInvalidRegistry) {
		t.Fatalf("expected ErrInvalidRegistry, got %v", err)
	}
}

func TestRegistryRegistersRoutes(t *testing.T) {
	called := false
	registry, err := features.NewRegistry(fakeModule{name: "utility", register: func(router *interactions.Router) error {
		called = true
		return router.RegisterSlash("ping", func(context.Context, interactions.Interaction, responses.Responder) error { return nil })
	}})
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	router := interactions.NewRouter()
	if err := registry.RegisterRoutes(router); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	if !called {
		t.Fatal("module RegisterRoutes was not called")
	}
}

type fakeModule struct {
	name     string
	commands []commands.Definition
	register func(*interactions.Router) error
}

func (m fakeModule) Name() string { return m.name }

func (m fakeModule) Commands() []commands.Definition {
	return append([]commands.Definition(nil), m.commands...)
}

func (m fakeModule) RegisterRoutes(router *interactions.Router) error {
	if m.register != nil {
		return m.register(router)
	}
	return nil
}
