package fakeinteractions

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"

func Component(customID string) interactions.Interaction {
	return interactions.Interaction{
		Type:     interactions.TypeComponent,
		CustomID: customID,
		Actor:    interactions.Actor{UserID: "user-1", GuildID: "guild-1"},
	}
}

func ComponentWithValues(customID string, values ...string) interactions.Interaction {
	interaction := Component(customID)
	interaction.Values = append([]string(nil), values...)
	return interaction
}
