package discordgo

import (
	"fmt"

	dgo "github.com/bwmarrin/discordgo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

func InteractionFromDiscord(event *dgo.InteractionCreate) (interactions.Interaction, error) {
	if event == nil || event.Interaction == nil {
		return interactions.Interaction{}, fmt.Errorf("discord interaction event is nil")
	}
	actor := actorFromDiscord(event.Interaction)
	switch event.Type {
	case dgo.InteractionApplicationCommand:
		data := event.ApplicationCommandData()
		subcommandGroup, subcommand, options, values, err := commandOptions(data.Options, resolvedChannels(data.Resolved), resolvedUsers(data.Resolved))
		if err != nil {
			return interactions.Interaction{}, err
		}
		return interactions.Interaction{
			ID:              event.ID,
			ApplicationID:   event.AppID,
			Type:            interactions.TypeSlash,
			CommandName:     data.Name,
			SubcommandGroup: subcommandGroup,
			Subcommand:      subcommand,
			Options:         options,
			CommandOptions:  values,
			ChannelID:       event.ChannelID,
			Locale:          string(event.Locale),
			GuildLocale:     guildLocale(event.Interaction),
			Actor:           actor,
		}, nil
	case dgo.InteractionApplicationCommandAutocomplete:
		data := event.ApplicationCommandData()
		return interactions.Interaction{
			ID:            event.ID,
			ApplicationID: event.AppID,
			Type:          interactions.TypeAutocomplete,
			CommandName:   data.Name,
			ChannelID:     event.ChannelID,
			Locale:        string(event.Locale),
			GuildLocale:   guildLocale(event.Interaction),
			Actor:         actor,
		}, nil
	case dgo.InteractionMessageComponent:
		data := event.MessageComponentData()
		input, err := ComponentInputFromDiscord(data)
		if err != nil {
			return interactions.Interaction{}, err
		}
		return interactions.Interaction{
			ID:                        event.ID,
			ApplicationID:             event.AppID,
			Type:                      interactions.TypeComponent,
			CustomID:                  input.CustomID,
			Values:                    input.Values,
			ChannelID:                 event.ChannelID,
			MessageID:                 interactionMessageID(event.Interaction),
			OriginalInteractionID:     originalInteractionID(event.Interaction),
			OriginalInteractionUserID: originalInteractionUserID(event.Interaction),
			Locale:                    string(event.Locale),
			GuildLocale:               guildLocale(event.Interaction),
			Actor:                     actor,
		}, nil
	case dgo.InteractionModalSubmit:
		data := event.ModalSubmitData()
		input, err := ModalInputFromDiscord(data)
		if err != nil {
			return interactions.Interaction{}, err
		}
		return interactions.Interaction{
			ID:            event.ID,
			ApplicationID: event.AppID,
			Type:          interactions.TypeModal,
			CustomID:      input.CustomID,
			ModalFields:   input.Fields,
			ChannelID:     event.ChannelID,
			Locale:        string(event.Locale),
			GuildLocale:   guildLocale(event.Interaction),
			Actor:         actor,
		}, nil
	default:
		return interactions.Interaction{}, fmt.Errorf("unsupported discord interaction type %d", event.Type)
	}
}

func resolvedChannels(resolved *dgo.ApplicationCommandInteractionDataResolved) map[string]*dgo.Channel {
	if resolved == nil {
		return nil
	}
	return resolved.Channels
}

func resolvedUsers(resolved *dgo.ApplicationCommandInteractionDataResolved) map[string]*dgo.User {
	if resolved == nil {
		return nil
	}
	return resolved.Users
}

func commandOptions(options []*dgo.ApplicationCommandInteractionDataOption, channels map[string]*dgo.Channel, users map[string]*dgo.User) (string, string, map[string]string, map[string]interactions.CommandOptionValue, error) {
	internalOptions := make([]interactions.CommandOption, 0, len(options))
	for _, option := range options {
		if option == nil {
			continue
		}
		internalOptions = append(internalOptions, fromDiscordOption(option, channels, users))
	}
	parsed, err := interactions.ParseCommandOptions(internalOptions)
	if err != nil {
		return "", "", nil, nil, err
	}
	return parsed.SubcommandGroup, parsed.Subcommand, parsed.Options, parsed.Values, nil
}

func fromDiscordOption(option *dgo.ApplicationCommandInteractionDataOption, channels map[string]*dgo.Channel, users map[string]*dgo.User) interactions.CommandOption {
	converted := interactions.CommandOption{
		Name:  option.Name,
		Type:  fromDiscordOptionType(option.Type),
		Value: option.Value,
	}
	if converted.Type == interactions.CommandOptionChannel {
		channelID := fmt.Sprint(option.Value)
		if channel := channels[channelID]; channel != nil {
			converted.ChannelName = channel.Name
			converted.ChannelType = int(channel.Type)
			converted.ChannelParentID = channel.ParentID
		}
	}
	if converted.Type == interactions.CommandOptionUser {
		userID := fmt.Sprint(option.Value)
		if user := users[userID]; user != nil {
			converted.UserName = user.Username
		}
	}
	for _, child := range option.Options {
		if child != nil {
			converted.Options = append(converted.Options, fromDiscordOption(child, channels, users))
		}
	}
	return converted
}

func fromDiscordOptionType(value dgo.ApplicationCommandOptionType) interactions.CommandOptionType {
	switch value {
	case dgo.ApplicationCommandOptionString:
		return interactions.CommandOptionString
	case dgo.ApplicationCommandOptionInteger:
		return interactions.CommandOptionInteger
	case dgo.ApplicationCommandOptionNumber:
		return interactions.CommandOptionNumber
	case dgo.ApplicationCommandOptionBoolean:
		return interactions.CommandOptionBoolean
	case dgo.ApplicationCommandOptionUser:
		return interactions.CommandOptionUser
	case dgo.ApplicationCommandOptionChannel:
		return interactions.CommandOptionChannel
	case dgo.ApplicationCommandOptionRole:
		return interactions.CommandOptionRole
	case dgo.ApplicationCommandOptionMentionable:
		return interactions.CommandOptionMentionable
	case dgo.ApplicationCommandOptionSubCommand:
		return interactions.CommandOptionSubcommand
	case dgo.ApplicationCommandOptionSubCommandGroup:
		return interactions.CommandOptionSubcommandGroup
	default:
		return interactions.CommandOptionType(fmt.Sprintf("unsupported:%d", value))
	}
}

func guildLocale(interaction *dgo.Interaction) string {
	if interaction == nil || interaction.GuildLocale == nil {
		return ""
	}
	return string(*interaction.GuildLocale)
}

func actorFromDiscord(interaction *dgo.Interaction) interactions.Actor {
	actor := interactions.Actor{GuildID: interaction.GuildID}
	if interaction.Member != nil && interaction.Member.User != nil {
		actor.UserID = interaction.Member.User.ID
		actor.Username = interaction.Member.User.Username
		actor.UserTag = userTag(interaction.Member.User)
		actor.AvatarURL = interaction.Member.User.AvatarURL("")
		member := *interaction.Member
		if member.GuildID == "" {
			member.GuildID = interaction.GuildID
		}
		actor.GuildAvatarURL = member.AvatarURL("")
		actor.RoleIDs = append([]string(nil), interaction.Member.Roles...)
		actor.PermissionBits = interaction.Member.Permissions
		return actor
	}
	if interaction.User != nil {
		actor.UserID = interaction.User.ID
		actor.Username = interaction.User.Username
		actor.UserTag = userTag(interaction.User)
		actor.AvatarURL = interaction.User.AvatarURL("")
	}
	return actor
}

func interactionMessageID(interaction *dgo.Interaction) string {
	if interaction == nil || interaction.Message == nil {
		return ""
	}
	return interaction.Message.ID
}

func originalInteractionID(interaction *dgo.Interaction) string {
	if interaction == nil || interaction.Message == nil {
		return ""
	}
	if metadata := interaction.Message.InteractionMetadata; metadata != nil {
		return metadata.ID
	}
	if legacy := interaction.Message.Interaction; legacy != nil {
		return legacy.ID
	}
	return ""
}

func originalInteractionUserID(interaction *dgo.Interaction) string {
	if interaction == nil || interaction.Message == nil {
		return ""
	}
	if metadata := interaction.Message.InteractionMetadata; metadata != nil && metadata.User != nil {
		return metadata.User.ID
	}
	if legacy := interaction.Message.Interaction; legacy != nil && legacy.User != nil {
		return legacy.User.ID
	}
	return ""
}

func userTag(user *dgo.User) string {
	if user == nil {
		return ""
	}
	if user.Discriminator != "" && user.Discriminator != "0" {
		return user.Username + "#" + user.Discriminator
	}
	return user.Username
}
