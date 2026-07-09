package discordgo

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

var _ responses.Responder = (*InteractionResponder)(nil)
