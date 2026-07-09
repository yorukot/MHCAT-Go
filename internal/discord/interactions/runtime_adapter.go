package interactions

import "fmt"

func ValidateRuntimeInteraction(interaction Interaction) error {
	switch interaction.Type {
	case TypeSlash:
		if interaction.CommandName == "" {
			return fmt.Errorf("%w: slash command name is required", ErrInvalidCommandOption)
		}
	case TypeComponent, TypeModal, TypeAutocomplete:
		return nil
	default:
		return fmt.Errorf("%w: unsupported interaction type %q", ErrInvalidCommandOption, interaction.Type)
	}
	return nil
}
