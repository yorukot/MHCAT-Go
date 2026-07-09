package utility

import (
	"context"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

func (m Module) HelpHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		query := helpQuery(interaction)
		msg := legacyHelpOverview(interaction)
		if query != "" {
			if category, ok := legacyHelpCategoryMessage(interaction, query); ok {
				msg = category
				msg.Ephemeral = false
			} else if detail, ok := legacyHelpCommandDetail(interaction, query); ok {
				msg = detail
			} else {
				msg = legacyHelpInvalidCommand()
			}
		}
		if err := responder.EditOriginal(ctx, msg); err != nil {
			return err
		}
		return m.track(ctx, interaction, "help")
	}
}

func (m Module) HelpComponentHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{Ephemeral: true}); err != nil {
			return err
		}
		selected := ""
		if len(interaction.Values) > 0 {
			selected = interaction.Values[0]
		}
		msg := legacyHelpOverview(interaction)
		msg.Ephemeral = true
		if selected != "" {
			if category, ok := legacyHelpCategoryMessage(interaction, selected); ok {
				msg = category
			} else if detail, ok := legacyHelpCommandDetail(interaction, selected); ok {
				msg = detail
				msg.Ephemeral = true
			} else {
				msg = legacyHelpInvalidCommand()
				msg.Ephemeral = true
			}
		}
		if err := responder.EditOriginal(ctx, msg); err != nil {
			return err
		}
		return m.track(ctx, interaction, "help")
	}
}

func helpQuery(interaction interactions.Interaction) string {
	for _, key := range []string{"指令名稱", "command", "name"} {
		if value := strings.TrimSpace(interaction.Options[key]); value != "" {
			return strings.Fields(value)[0]
		}
	}
	return ""
}
