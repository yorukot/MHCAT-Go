package discordgo

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	dgo "github.com/bwmarrin/discordgo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

type CommandSyncClient struct {
	session       *dgo.Session
	applicationID string
}

var errCommandSyncClientNotConfigured = errors.New("discord command sync client is not configured")

func NewCommandSyncClient(token, applicationID string) (*CommandSyncClient, error) {
	token = strings.TrimSpace(token)
	applicationID = strings.TrimSpace(applicationID)
	if token == "" || applicationID == "" {
		return nil, errCommandSyncClientNotConfigured
	}
	session, err := dgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("create discord rest session: %w", err)
	}
	return &CommandSyncClient{session: session, applicationID: applicationID}, nil
}

func (c *CommandSyncClient) ListCommands(ctx context.Context, scope commands.Scope) ([]commands.RemoteCommand, error) {
	if err := c.ready(ctx); err != nil {
		return nil, err
	}
	remote, err := c.session.ApplicationCommands(c.applicationID, scope.GuildID, dgo.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("list application commands: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	result := make([]commands.RemoteCommand, 0, len(remote))
	for _, command := range remote {
		if command == nil {
			continue
		}
		result = append(result, fromDiscordCommand(command))
	}
	commands.SortRemote(result)
	return result, nil
}

func (c *CommandSyncClient) CreateCommand(ctx context.Context, scope commands.Scope, definition commands.Definition) (commands.RemoteCommand, error) {
	if err := c.ready(ctx); err != nil {
		return commands.RemoteCommand{}, err
	}
	created, err := c.session.ApplicationCommandCreate(c.applicationID, scope.GuildID, toDiscordCommand(definition), dgo.WithContext(ctx))
	if err != nil {
		return commands.RemoteCommand{}, fmt.Errorf("create application command: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return commands.RemoteCommand{}, err
	}
	if created == nil {
		return commands.RemoteCommand{}, errors.New("create application command returned an empty response")
	}
	return fromDiscordCommand(created), nil
}

func (c *CommandSyncClient) UpdateCommand(ctx context.Context, scope commands.Scope, remoteID string, definition commands.Definition) (commands.RemoteCommand, error) {
	if err := c.ready(ctx); err != nil {
		return commands.RemoteCommand{}, err
	}
	updated, err := c.session.ApplicationCommandEdit(c.applicationID, scope.GuildID, remoteID, toDiscordCommand(definition), dgo.WithContext(ctx))
	if err != nil {
		return commands.RemoteCommand{}, fmt.Errorf("edit application command: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return commands.RemoteCommand{}, err
	}
	if updated == nil {
		return commands.RemoteCommand{}, errors.New("edit application command returned an empty response")
	}
	return fromDiscordCommand(updated), nil
}

func (c *CommandSyncClient) DeleteCommand(ctx context.Context, scope commands.Scope, remoteID string) error {
	if err := c.ready(ctx); err != nil {
		return err
	}
	if err := c.session.ApplicationCommandDelete(c.applicationID, scope.GuildID, remoteID, dgo.WithContext(ctx)); err != nil {
		return fmt.Errorf("delete application command: %w", err)
	}
	return ctx.Err()
}

func (c *CommandSyncClient) BulkOverwriteCommands(ctx context.Context, scope commands.Scope, definitions []commands.Definition) ([]commands.RemoteCommand, error) {
	if err := c.ready(ctx); err != nil {
		return nil, err
	}
	payload := make([]*dgo.ApplicationCommand, 0, len(definitions))
	for _, definition := range definitions {
		payload = append(payload, toDiscordCommand(definition))
	}
	created, err := c.session.ApplicationCommandBulkOverwrite(c.applicationID, scope.GuildID, payload, dgo.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("bulk overwrite application commands: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	result := make([]commands.RemoteCommand, 0, len(created))
	for _, command := range created {
		if command == nil {
			continue
		}
		result = append(result, fromDiscordCommand(command))
	}
	commands.SortRemote(result)
	return result, nil
}

func (c *CommandSyncClient) ready(ctx context.Context) error {
	if ctx == nil {
		return errCommandSyncClientNotConfigured
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if c == nil || c.session == nil || strings.TrimSpace(c.applicationID) == "" {
		return errCommandSyncClientNotConfigured
	}
	return nil
}

func toDiscordCommand(definition commands.Definition) *dgo.ApplicationCommand {
	command := &dgo.ApplicationCommand{
		Type:        dgo.ApplicationCommandType(definition.Type),
		Name:        definition.Name,
		Description: definition.Description,
		Options:     toDiscordOptions(definition.Options),
	}
	if len(definition.NameLocalizations) > 0 {
		command.NameLocalizations = toDiscordLocales(definition.NameLocalizations)
	}
	if len(definition.DescriptionLocalizations) > 0 {
		command.DescriptionLocalizations = toDiscordLocales(definition.DescriptionLocalizations)
	}
	if definition.DefaultMemberPermissions != nil && *definition.DefaultMemberPermissions != "" {
		if parsed, err := strconv.ParseInt(*definition.DefaultMemberPermissions, 10, 64); err == nil {
			command.DefaultMemberPermissions = &parsed
		}
	}
	if len(definition.Contexts) > 0 {
		contexts := make([]dgo.InteractionContextType, 0, len(definition.Contexts))
		for _, value := range definition.Contexts {
			contexts = append(contexts, dgo.InteractionContextType(value))
		}
		command.Contexts = &contexts
	}
	if len(definition.IntegrationTypes) > 0 {
		integrationTypes := make([]dgo.ApplicationIntegrationType, 0, len(definition.IntegrationTypes))
		for _, value := range definition.IntegrationTypes {
			integrationTypes = append(integrationTypes, dgo.ApplicationIntegrationType(value))
		}
		command.IntegrationTypes = &integrationTypes
	}
	if definition.NSFW {
		nsfw := true
		command.NSFW = &nsfw
	}
	return command
}

func toDiscordOptions(options []commands.Option) []*dgo.ApplicationCommandOption {
	result := make([]*dgo.ApplicationCommandOption, 0, len(options))
	for _, option := range options {
		converted := &dgo.ApplicationCommandOption{
			Type:        dgo.ApplicationCommandOptionType(option.Type),
			Name:        option.Name,
			Description: option.Description,
			Required:    option.Required,
			Options:     toDiscordOptions(option.Options),
			Choices:     toDiscordChoices(option.Choices),
		}
		for _, channelType := range option.ChannelTypes {
			converted.ChannelTypes = append(converted.ChannelTypes, dgo.ChannelType(channelType))
		}
		if len(option.NameLocalizations) > 0 {
			converted.NameLocalizations = valueDiscordLocales(option.NameLocalizations)
		}
		if len(option.DescriptionLocalizations) > 0 {
			converted.DescriptionLocalizations = valueDiscordLocales(option.DescriptionLocalizations)
		}
		result = append(result, converted)
	}
	return result
}

func toDiscordChoices(choices []commands.Choice) []*dgo.ApplicationCommandOptionChoice {
	result := make([]*dgo.ApplicationCommandOptionChoice, 0, len(choices))
	for _, choice := range choices {
		converted := &dgo.ApplicationCommandOptionChoice{
			Name:  choice.Name,
			Value: choice.Value,
		}
		if len(choice.NameLocalizations) > 0 {
			converted.NameLocalizations = valueDiscordLocales(choice.NameLocalizations)
		}
		result = append(result, converted)
	}
	return result
}

func fromDiscordCommand(command *dgo.ApplicationCommand) commands.RemoteCommand {
	definition := commands.Definition{
		Type:                     commands.CommandType(command.Type),
		Name:                     command.Name,
		Description:              command.Description,
		NameLocalizations:        fromDiscordLocales(command.NameLocalizations),
		DescriptionLocalizations: fromDiscordLocales(command.DescriptionLocalizations),
		Options:                  fromDiscordOptions(command.Options),
		NSFW:                     command.NSFW != nil && *command.NSFW,
	}
	if command.DefaultMemberPermissions != nil {
		value := strconv.FormatInt(*command.DefaultMemberPermissions, 10)
		definition.DefaultMemberPermissions = &value
	}
	if command.Contexts != nil {
		for _, value := range *command.Contexts {
			definition.Contexts = append(definition.Contexts, int(value))
		}
	}
	if command.IntegrationTypes != nil {
		for _, value := range *command.IntegrationTypes {
			definition.IntegrationTypes = append(definition.IntegrationTypes, int(value))
		}
	}
	return commands.RemoteCommand{
		ID:            command.ID,
		ApplicationID: command.ApplicationID,
		GuildID:       command.GuildID,
		Version:       command.Version,
		Definition:    definition,
	}
}

func fromDiscordOptions(options []*dgo.ApplicationCommandOption) []commands.Option {
	result := make([]commands.Option, 0, len(options))
	for _, option := range options {
		if option == nil {
			continue
		}
		result = append(result, commands.Option{
			Type:                     commands.OptionType(option.Type),
			Name:                     option.Name,
			Description:              option.Description,
			NameLocalizations:        fromDiscordLocalesMap(option.NameLocalizations),
			DescriptionLocalizations: fromDiscordLocalesMap(option.DescriptionLocalizations),
			Required:                 option.Required,
			Options:                  fromDiscordOptions(option.Options),
			Choices:                  fromDiscordChoices(option.Choices),
			ChannelTypes:             fromDiscordChannelTypes(option.ChannelTypes),
		})
	}
	return result
}

func fromDiscordChoices(choices []*dgo.ApplicationCommandOptionChoice) []commands.Choice {
	result := make([]commands.Choice, 0, len(choices))
	for _, choice := range choices {
		if choice == nil {
			continue
		}
		result = append(result, commands.Choice{
			Name:              choice.Name,
			NameLocalizations: fromDiscordLocalesMap(choice.NameLocalizations),
			Value:             choice.Value,
		})
	}
	return result
}

func fromDiscordChannelTypes(channelTypes []dgo.ChannelType) []int {
	result := make([]int, 0, len(channelTypes))
	for _, channelType := range channelTypes {
		result = append(result, int(channelType))
	}
	return result
}

func toDiscordLocales(values map[string]string) *map[dgo.Locale]string {
	locales := valueDiscordLocales(values)
	return &locales
}

func valueDiscordLocales(values map[string]string) map[dgo.Locale]string {
	locales := make(map[dgo.Locale]string, len(values))
	for key, value := range values {
		locales[dgo.Locale(key)] = value
	}
	return locales
}

func fromDiscordLocales(values *map[dgo.Locale]string) map[string]string {
	if values == nil {
		return nil
	}
	return fromDiscordLocalesMap(*values)
}

func fromDiscordLocalesMap(values map[dgo.Locale]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	locales := make(map[string]string, len(values))
	for key, value := range values {
		locales[string(key)] = value
	}
	return locales
}
