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
		query, supplied := helpQuery(interaction)
		msg := legacyHelpOverview(interaction)
		if supplied {
			if category, ok := legacyHelpSlashCategoryMessage(interaction, query); ok {
				msg = category
			} else if detail, ok := legacyHelpCommandDetail(interaction, query, m.helpDefinitions); ok {
				msg = detail
			} else {
				msg = legacyHelpInvalidCommand()
			}
		}
		if err := responder.FollowUp(ctx, msg); err != nil {
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
			if category, ok := legacyHelpCategoryMessage(interaction, selected, m.helpDefinitions); ok {
				msg = category
			} else if detail, ok := legacyHelpCommandDetail(interaction, selected, m.helpDefinitions); ok {
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

func helpQuery(interaction interactions.Interaction) (string, bool) {
	for _, key := range []string{"指令名稱", "command", "name"} {
		if value := interaction.Options[key]; value != "" {
			return strings.Split(value, " ")[0], true
		}
	}
	return "", false
}
