package fakediscord

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

type CommandSyncClient struct {
	Remote         []commands.RemoteCommand
	ListErr        error
	CreateErr      error
	UpdateErr      error
	DeleteErr      error
	BulkErr        error
	Created        []commands.Definition
	Updated        []UpdatedCommand
	Deleted        []string
	BulkOverwrites [][]commands.Definition
}

type UpdatedCommand struct {
	RemoteID   string
	Definition commands.Definition
}

func (c *CommandSyncClient) ListCommands(ctx context.Context, scope commands.Scope) ([]commands.RemoteCommand, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if c.ListErr != nil {
		return nil, c.ListErr
	}
	remote := append([]commands.RemoteCommand(nil), c.Remote...)
	commands.SortRemote(remote)
	return remote, nil
}

func (c *CommandSyncClient) CreateCommand(ctx context.Context, scope commands.Scope, definition commands.Definition) (commands.RemoteCommand, error) {
	if err := ctx.Err(); err != nil {
		return commands.RemoteCommand{}, err
	}
	if c.CreateErr != nil {
		return commands.RemoteCommand{}, c.CreateErr
	}
	c.Created = append(c.Created, definition)
	remote := commands.RemoteCommand{
		ID:         fmt.Sprintf("created-%d", len(c.Created)),
		GuildID:    scope.GuildID,
		Definition: definition,
		Owned:      true,
	}
	c.Remote = append(c.Remote, remote)
	return remote, nil
}

func (c *CommandSyncClient) UpdateCommand(ctx context.Context, scope commands.Scope, remoteID string, definition commands.Definition) (commands.RemoteCommand, error) {
	if err := ctx.Err(); err != nil {
		return commands.RemoteCommand{}, err
	}
	if c.UpdateErr != nil {
		return commands.RemoteCommand{}, c.UpdateErr
	}
	c.Updated = append(c.Updated, UpdatedCommand{RemoteID: remoteID, Definition: definition})
	for index := range c.Remote {
		if c.Remote[index].ID == remoteID {
			c.Remote[index].Definition = definition
			return c.Remote[index], nil
		}
	}
	return commands.RemoteCommand{}, errors.New("remote command not found")
}

func (c *CommandSyncClient) DeleteCommand(ctx context.Context, scope commands.Scope, remoteID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if c.DeleteErr != nil {
		return c.DeleteErr
	}
	c.Deleted = append(c.Deleted, remoteID)
	for index := range c.Remote {
		if c.Remote[index].ID == remoteID {
			c.Remote = append(c.Remote[:index], c.Remote[index+1:]...)
			return nil
		}
	}
	return errors.New("remote command not found")
}

func (c *CommandSyncClient) BulkOverwriteCommands(ctx context.Context, scope commands.Scope, definitions []commands.Definition) ([]commands.RemoteCommand, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if c.BulkErr != nil {
		return nil, c.BulkErr
	}
	c.BulkOverwrites = append(c.BulkOverwrites, append([]commands.Definition(nil), definitions...))
	c.Remote = c.Remote[:0]
	for index, definition := range definitions {
		c.Remote = append(c.Remote, commands.RemoteCommand{
			ID:         "bulk-" + strconv.Itoa(index+1),
			GuildID:    scope.GuildID,
			Definition: definition,
			Owned:      true,
		})
	}
	return append([]commands.RemoteCommand(nil), c.Remote...), nil
}
