package features

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

type Module interface {
	Name() string
	Commands() []commands.Definition
	RegisterRoutes(router *interactions.Router) error
}
