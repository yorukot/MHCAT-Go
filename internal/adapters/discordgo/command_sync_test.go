package discordgo

import (
	"testing"

	dgo "github.com/bwmarrin/discordgo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestCommandOptionChoicesAndChannelTypesRoundTrip(t *testing.T) {
	definition := commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        "translate",
		Description: "Translate",
		Options: []commands.Option{
			{
				Type:        commands.OptionTypeString,
				Name:        "target",
				Description: "language",
				Choices: []commands.Choice{
					{Name: "繁體中文", Value: "zh-TW"},
					{Name: "English", Value: "en"},
				},
			},
			{
				Type:         commands.OptionTypeChannel,
				Name:         "channel",
				Description:  "channel",
				ChannelTypes: []int{int(dgo.ChannelTypeGuildText), int(dgo.ChannelTypeGuildNews)},
			},
		},
	}
	discordCommand := toDiscordCommand(definition)
	if len(discordCommand.Options) != 2 || len(discordCommand.Options[0].Choices) != 2 {
		t.Fatalf("discord options = %#v", discordCommand.Options)
	}
	if got := discordCommand.Options[1].ChannelTypes; len(got) != 2 || got[0] != dgo.ChannelTypeGuildText || got[1] != dgo.ChannelTypeGuildNews {
		t.Fatalf("channel types = %#v", got)
	}
	roundTrip := fromDiscordCommand(&dgo.ApplicationCommand{
		Type:        dgo.ChatApplicationCommand,
		Name:        discordCommand.Name,
		Description: discordCommand.Description,
		Options:     discordCommand.Options,
	}).Definition
	if len(roundTrip.Options[0].Choices) != 2 || roundTrip.Options[0].Choices[0].Value != "zh-TW" {
		t.Fatalf("round-trip choices = %#v", roundTrip.Options[0].Choices)
	}
	if len(roundTrip.Options[1].ChannelTypes) != 2 || roundTrip.Options[1].ChannelTypes[1] != int(dgo.ChannelTypeGuildNews) {
		t.Fatalf("round-trip channel types = %#v", roundTrip.Options[1].ChannelTypes)
	}
}
