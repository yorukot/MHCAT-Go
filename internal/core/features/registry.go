package features

import (
	"fmt"
	"sort"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

type Registry struct {
	modules []Module
}

func NewRegistry(modules ...Module) (*Registry, error) {
	registry := &Registry{}
	for _, module := range modules {
		if err := registry.Add(module); err != nil {
			return nil, err
		}
	}
	return registry, nil
}

func (r *Registry) Add(module Module) error {
	if module == nil {
		return fmt.Errorf("feature module is required")
	}
	if module.Name() == "" {
		return fmt.Errorf("feature module name is required")
	}
	for _, existing := range r.modules {
		if existing.Name() == module.Name() {
			return fmt.Errorf("duplicate feature module %q", module.Name())
		}
	}
	r.modules = append(r.modules, module)
	sort.SliceStable(r.modules, func(i, j int) bool {
		return r.modules[i].Name() < r.modules[j].Name()
	})
	return nil
}

func (r *Registry) Modules() []Module {
	return append([]Module(nil), r.modules...)
}

func (r *Registry) CommandRegistry(scope commands.Scope) (commands.Registry, error) {
	definitions := make([]commands.Definition, 0)
	for _, module := range r.modules {
		definitions = append(definitions, module.Commands()...)
	}
	registry := commands.NewRegistry(scope, definitions)
	if err := commands.ValidateRegistry(registry); err != nil {
		return registry, err
	}
	return registry, nil
}

func (r *Registry) RegisterRoutes(router *interactions.Router) error {
	if router == nil {
		return fmt.Errorf("interaction router is required")
	}
	for _, module := range r.modules {
		if err := module.RegisterRoutes(router); err != nil {
			return fmt.Errorf("register feature %s routes: %w", module.Name(), err)
		}
	}
	return nil
}
