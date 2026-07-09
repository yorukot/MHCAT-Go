package fakediscord

import (
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

func SlashInteraction(name string) interactions.Interaction {
	return interactions.Interaction{
		Type:        interactions.TypeSlash,
		CommandName: name,
		Options:     map[string]string{},
		Actor:       interactions.Actor{UserID: "user-1", Username: "User", UserTag: "User#0001", GuildID: "guild-1"},
	}
}

func SlashInteractionWithOptions(name string, subcommand string, options map[string]string) interactions.Interaction {
	copied := make(map[string]string, len(options))
	for key, value := range options {
		copied[key] = value
	}
	return interactions.Interaction{
		Type:        interactions.TypeSlash,
		CommandName: name,
		Subcommand:  subcommand,
		Options:     copied,
		Actor:       interactions.Actor{UserID: "user-1", Username: "User", UserTag: "User#0001", GuildID: "guild-1"},
	}
}

func SlashInteractionCreatedAt(name string, createdAt time.Time) interactions.Interaction {
	interaction := SlashInteraction(name)
	interaction.CreatedAt = createdAt
	return interaction
}

func ComponentInteraction(key interactions.ComponentKey) interactions.Interaction {
	return interactions.Interaction{
		Type:         interactions.TypeComponent,
		ComponentKey: key,
		Actor:        interactions.Actor{UserID: "user-1", Username: "User", UserTag: "User#0001", GuildID: "guild-1"},
	}
}

func ComponentInteractionFromID(customID string) interactions.Interaction {
	return interactions.Interaction{
		Type:     interactions.TypeComponent,
		CustomID: customID,
		Actor:    interactions.Actor{UserID: "user-1", Username: "User", UserTag: "User#0001", GuildID: "guild-1"},
	}
}

func ModalInteraction(key interactions.ModalKey) interactions.Interaction {
	return interactions.Interaction{
		Type:     interactions.TypeModal,
		ModalKey: key,
		Actor:    interactions.Actor{UserID: "user-1", Username: "User", UserTag: "User#0001", GuildID: "guild-1"},
	}
}
