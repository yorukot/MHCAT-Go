package fakefeatures

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

type Module struct {
	ModuleName string
	Defs       []commands.Definition
	Register   func(*interactions.Router) error
}

func (m Module) Name() string {
	return m.ModuleName
}

func (m Module) Commands() []commands.Definition {
	return append([]commands.Definition(nil), m.Defs...)
}

func (m Module) RegisterRoutes(router *interactions.Router) error {
	if m.Register != nil {
		return m.Register(router)
	}
	return nil
}
