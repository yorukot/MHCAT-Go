package components

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
)

type SelectOption struct {
	Label       string
	Value       string
	Description string
}

type HelpMenu struct {
	CustomID string
	Options  []SelectOption
}

func BuildHelpMenu(registry commands.Registry) (HelpMenu, error) {
	payload, err := customid.TokenPayload("overview")
	if err != nil {
		return HelpMenu{}, err
	}
	encoded, err := customid.Encode(customid.InteractionKindComponent, "help", "category", payload)
	if err != nil {
		return HelpMenu{}, err
	}
	definitions := commands.EnabledDefinitions(registry)
	options := make([]SelectOption, 0, len(definitions))
	for _, definition := range definitions {
		options = append(options, SelectOption{
			Label:       definition.Name,
			Value:       definition.Name,
			Description: definition.Description,
		})
	}
	return HelpMenu{CustomID: encoded, Options: options}, nil
}
