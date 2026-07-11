package moderation

import (
	"context"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type WarningHistoryService struct {
	Repository ports.WarningHistoryRepository
}

type WarningSettingsService struct {
	Repository ports.WarningSettingsRepository
}

type WarningIssueService struct {
	Repository ports.WarningIssueRepository
}

type WarningRemovalService struct {
	Repository ports.WarningRemovalRepository
}

func (s WarningHistoryService) History(ctx context.Context, guildID string, userID string) (domain.WarningHistory, error) {
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return domain.WarningHistory{}, domain.ErrInvalidWarningQuery
	}
	if s.Repository == nil {
		return domain.WarningHistory{}, ports.ErrWarningsNotFound
	}
	history, err := s.Repository.GetWarningHistory(ctx, guildID, userID)
	if err != nil {
		return domain.WarningHistory{}, err
	}
	if len(history.Entries) == 0 {
		return domain.WarningHistory{}, ports.ErrWarningsNotFound
	}
	return history, nil
}

func (s WarningSettingsService) Configure(ctx context.Context, settings domain.WarningSettings) error {
	settings.GuildID = strings.TrimSpace(settings.GuildID)
	settings.Action = strings.TrimSpace(settings.Action)
	if err := settings.Validate(); err != nil {
		return err
	}
	if s.Repository == nil {
		return ports.ErrWarningSettingsUnavailable
	}
	return s.Repository.SaveWarningSettings(ctx, settings)
}

func (s WarningSettingsService) Settings(ctx context.Context, guildID string) (domain.WarningSettings, error) {
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.WarningSettings{}, domain.ErrInvalidWarningSettings
	}
	if s.Repository == nil {
		return domain.WarningSettings{}, ports.ErrWarningSettingsUnavailable
	}
	return s.Repository.GetWarningSettings(ctx, guildID)
}

func (s WarningIssueService) Issue(ctx context.Context, issue domain.WarningIssue) (domain.WarningIssueResult, error) {
	issue.GuildID = strings.TrimSpace(issue.GuildID)
	issue.UserID = strings.TrimSpace(issue.UserID)
	issue.ModeratorID = strings.TrimSpace(issue.ModeratorID)
	issue.Time = strings.TrimSpace(issue.Time)
	if err := issue.Validate(); err != nil {
		return domain.WarningIssueResult{}, err
	}
	if s.Repository == nil {
		return domain.WarningIssueResult{}, ports.ErrWarningIssueUnavailable
	}
	return s.Repository.AddWarning(ctx, issue)
}

func (s WarningRemovalService) RemoveOne(ctx context.Context, removal domain.WarningRemoval) error {
	removal.GuildID = strings.TrimSpace(removal.GuildID)
	removal.UserID = strings.TrimSpace(removal.UserID)
	if err := removal.ValidateSingle(); err != nil {
		return err
	}
	if s.Repository == nil {
		return ports.ErrWarningRemovalUnavailable
	}
	return s.Repository.RemoveWarning(ctx, removal)
}

func (s WarningRemovalService) RemoveAll(ctx context.Context, removal domain.WarningRemoval) error {
	removal.GuildID = strings.TrimSpace(removal.GuildID)
	removal.UserID = strings.TrimSpace(removal.UserID)
	if err := removal.ValidateAll(); err != nil {
		return err
	}
	if s.Repository == nil {
		return ports.ErrWarningRemovalUnavailable
	}
	return s.Repository.RemoveAllWarnings(ctx, removal)
}
