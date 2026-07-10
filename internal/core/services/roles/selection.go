package roles

import (
	"context"
	"errors"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

var ErrRoleAlreadyAssigned = errors.New("role already assigned")
var ErrRoleNotAssigned = errors.New("role not assigned")

type ReactionSetCommand struct {
	GuildID    string
	MessageURL string
	RoleID     string
	Emoji      string
}

type ReactionDeleteCommand struct {
	GuildID    string
	MessageURL string
	Emoji      string
}

type ButtonPrepareCommand struct {
	GuildID string
	RoleID  string
	BaseID  string
}

type ButtonPrepareResult struct {
	AddID    string
	RemoveID string
}

type ButtonApplyCommand struct {
	GuildID      string
	UserID       string
	Number       string
	Remove       bool
	ActorRoleIDs []string
}

type ReactionApplyCommand struct {
	GuildID   string
	MessageID string
	React     string
	UserID    string
	Remove    bool
}

type SelectionService struct {
	Repository    ports.RoleSelectionRepository
	RoleInspector ports.DiscordRoleInspector
	Roles         ports.DiscordRolePort
	Reactions     ports.DiscordReactionPort
}

func (s SelectionService) ConfigureReaction(ctx context.Context, command ReactionSetCommand) (domain.RoleReactionConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.RoleReactionConfig{}, err
	}
	if s.Repository == nil || s.RoleInspector == nil || s.Reactions == nil {
		return domain.RoleReactionConfig{}, domain.ErrInvalidRoleSelectionConfig
	}
	guildID := strings.TrimSpace(command.GuildID)
	roleID := strings.TrimSpace(command.RoleID)
	if guildID == "" || roleID == "" {
		return domain.RoleReactionConfig{}, domain.ErrInvalidRoleSelectionConfig
	}
	assignable, err := s.RoleInspector.CanAssignRole(ctx, guildID, roleID)
	if err != nil {
		return domain.RoleReactionConfig{}, err
	}
	if !assignable {
		return domain.RoleReactionConfig{}, ports.ErrDiscordRoleNotAssignable
	}
	target, err := domain.ParseLegacyDiscordMessageURL(command.MessageURL, true)
	if err != nil {
		return domain.RoleReactionConfig{}, err
	}
	if _, err := s.Reactions.FindCachedChannelByID(ctx, guildID, target.ChannelID); err != nil {
		return domain.RoleReactionConfig{}, err
	}
	reaction, err := domain.NormalizeLegacyReaction(command.Emoji)
	if err != nil {
		return domain.RoleReactionConfig{}, err
	}
	if reaction.CustomEmojiID != "" {
		exists, err := s.Reactions.CachedEmojiExists(ctx, reaction.CustomEmojiID)
		if err != nil {
			return domain.RoleReactionConfig{}, err
		}
		if !exists {
			return domain.RoleReactionConfig{}, domain.ErrInvalidRoleSelectionEmoji
		}
	} else if !domain.IsLegacyUnicodeEmoji(command.Emoji) {
		return domain.RoleReactionConfig{}, domain.ErrInvalidRoleSelectionEmoji
	}
	if _, err := s.Reactions.FetchMessage(ctx, target.ChannelID, target.MessageID); err != nil {
		return domain.RoleReactionConfig{}, err
	}
	if err := s.Reactions.AddReaction(ctx, target.ChannelID, target.MessageID, reaction.API); err != nil {
		return domain.RoleReactionConfig{}, err
	}
	config := domain.RoleReactionConfig{
		GuildID:   guildID,
		MessageID: target.MessageID,
		React:     reaction.Stored,
		RoleID:    roleID,
	}
	if err := s.Repository.SaveRoleReactionConfig(ctx, config); err != nil {
		return domain.RoleReactionConfig{}, err
	}
	return config, ctx.Err()
}

func (s SelectionService) DeleteReaction(ctx context.Context, command ReactionDeleteCommand) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.Repository == nil || s.Reactions == nil {
		return domain.ErrInvalidRoleSelectionConfig
	}
	guildID := strings.TrimSpace(command.GuildID)
	target, err := domain.ParseLegacyDiscordMessageURL(command.MessageURL, false)
	if err != nil {
		return err
	}
	if guildID == "" {
		return domain.ErrInvalidRoleSelectionConfig
	}
	if _, err := s.Reactions.FindCachedChannelByID(ctx, guildID, target.ChannelID); err != nil {
		return err
	}
	reaction, err := domain.NormalizeLegacyReaction(command.Emoji)
	if err != nil {
		return err
	}
	if _, err := s.Reactions.FetchMessage(ctx, target.ChannelID, target.MessageID); err != nil {
		return err
	}
	if err := s.Reactions.AddReaction(ctx, target.ChannelID, target.MessageID, reaction.API); err != nil {
		return err
	}
	if err := s.Repository.DeleteRoleReactionConfig(ctx, guildID, target.MessageID, reaction.Stored); err != nil {
		return err
	}
	return ctx.Err()
}

func (s SelectionService) PrepareButton(ctx context.Context, command ButtonPrepareCommand) (ButtonPrepareResult, error) {
	if err := ctx.Err(); err != nil {
		return ButtonPrepareResult{}, err
	}
	if s.Repository == nil || s.RoleInspector == nil {
		return ButtonPrepareResult{}, domain.ErrInvalidRoleSelectionConfig
	}
	guildID := strings.TrimSpace(command.GuildID)
	roleID := strings.TrimSpace(command.RoleID)
	baseID := strings.TrimSpace(command.BaseID)
	if guildID == "" || roleID == "" || baseID == "" {
		return ButtonPrepareResult{}, domain.ErrInvalidRoleSelectionConfig
	}
	assignable, err := s.RoleInspector.CanAssignRole(ctx, guildID, roleID)
	if err != nil {
		return ButtonPrepareResult{}, err
	}
	if !assignable {
		return ButtonPrepareResult{}, ports.ErrDiscordRoleNotAssignable
	}
	result := ButtonPrepareResult{AddID: baseID + "add", RemoveID: baseID + "delete"}
	if err := s.Repository.SaveRoleButtonConfigs(ctx,
		domain.RoleButtonConfig{GuildID: guildID, Number: result.AddID, RoleID: roleID},
		domain.RoleButtonConfig{GuildID: guildID, Number: result.RemoveID, RoleID: roleID},
	); err != nil {
		return ButtonPrepareResult{}, err
	}
	return result, ctx.Err()
}

func (s SelectionService) ApplyButton(ctx context.Context, command ButtonApplyCommand) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.Repository == nil || s.RoleInspector == nil || s.Roles == nil {
		return domain.ErrInvalidRoleSelectionConfig
	}
	guildID := strings.TrimSpace(command.GuildID)
	userID := strings.TrimSpace(command.UserID)
	number := strings.TrimSpace(command.Number)
	if guildID == "" || userID == "" || number == "" {
		return domain.ErrInvalidRoleSelectionConfig
	}
	config, err := s.Repository.GetRoleButtonConfig(ctx, guildID, number)
	if err != nil {
		return err
	}
	if command.Remove {
		if !hasRole(command.ActorRoleIDs, config.RoleID) {
			return ErrRoleNotAssigned
		}
		if err := s.ensureAssignable(ctx, guildID, config.RoleID); err != nil {
			return err
		}
		return s.Roles.RemoveRole(ctx, guildID, userID, config.RoleID)
	}
	if hasRole(command.ActorRoleIDs, config.RoleID) {
		return ErrRoleAlreadyAssigned
	}
	if err := s.ensureAssignable(ctx, guildID, config.RoleID); err != nil {
		return err
	}
	return s.Roles.AddRole(ctx, guildID, userID, config.RoleID)
}

func (s SelectionService) ApplyReaction(ctx context.Context, command ReactionApplyCommand) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.Repository == nil || s.Roles == nil {
		return domain.ErrInvalidRoleSelectionConfig
	}
	guildID := strings.TrimSpace(command.GuildID)
	messageID := strings.TrimSpace(command.MessageID)
	react := strings.TrimSpace(command.React)
	userID := strings.TrimSpace(command.UserID)
	if guildID == "" || messageID == "" || react == "" || userID == "" {
		return domain.ErrInvalidRoleSelectionConfig
	}
	config, err := s.Repository.GetRoleReactionConfig(ctx, guildID, messageID, react)
	if err != nil {
		return err
	}
	if s.RoleInspector != nil {
		if err := s.ensureAssignable(ctx, guildID, config.RoleID); err != nil {
			return err
		}
	}
	if command.Remove {
		return s.Roles.RemoveRole(ctx, guildID, userID, config.RoleID)
	}
	return s.Roles.AddRole(ctx, guildID, userID, config.RoleID)
}

func (s SelectionService) ensureAssignable(ctx context.Context, guildID string, roleID string) error {
	assignable, err := s.RoleInspector.CanAssignRole(ctx, guildID, roleID)
	if err != nil {
		return err
	}
	if !assignable {
		return ports.ErrDiscordRoleNotAssignable
	}
	return nil
}

func hasRole(roleIDs []string, roleID string) bool {
	roleID = strings.TrimSpace(roleID)
	for _, current := range roleIDs {
		if strings.TrimSpace(current) == roleID {
			return true
		}
	}
	return false
}
