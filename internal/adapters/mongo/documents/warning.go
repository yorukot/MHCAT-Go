package documents

import (
	"fmt"
	"strconv"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

type WarningDocument struct {
	Guild   string                 `bson:"guild" json:"guild"`
	User    string                 `bson:"user" json:"user"`
	Content []WarningEntryDocument `bson:"content" json:"content"`
}

type WarningEntryDocument struct {
	Moderator string `bson:"moderator" json:"moderator"`
	Reason    string `bson:"reason" json:"reason"`
	Time      string `bson:"time" json:"time"`
}

type WarningSettingsDocument struct {
	Guild    string `bson:"guild" json:"guild"`
	BanCount string `bson:"ban_count" json:"ban_count"`
	Move     string `bson:"move" json:"move"`
}

func (d WarningDocument) ToDomain() domain.WarningHistory {
	entries := make([]domain.WarningEntry, 0, len(d.Content))
	for _, entry := range d.Content {
		entries = append(entries, domain.WarningEntry{
			ModeratorID: entry.Moderator,
			Reason:      entry.Reason,
			Time:        entry.Time,
		})
	}
	return domain.WarningHistory{
		GuildID: d.Guild,
		UserID:  d.User,
		Entries: entries,
	}
}

func WarningSettingsDocumentFromDomain(settings domain.WarningSettings) WarningSettingsDocument {
	return WarningSettingsDocument{
		Guild:    settings.GuildID,
		BanCount: strconv.FormatInt(settings.Threshold, 10),
		Move:     settings.Action,
	}
}

func WarningEntryDocumentFromIssue(issue domain.WarningIssue) WarningEntryDocument {
	return WarningEntryDocument{
		Moderator: issue.ModeratorID,
		Reason:    issue.Reason,
		Time:      issue.Time,
	}
}

func (d WarningSettingsDocument) ToDomain() (domain.WarningSettings, error) {
	threshold, err := strconv.ParseInt(d.BanCount, 10, 64)
	if err != nil {
		return domain.WarningSettings{}, fmt.Errorf("%w: ban_count", domain.ErrInvalidWarningSettings)
	}
	settings := domain.WarningSettings{
		GuildID:   d.Guild,
		Threshold: threshold,
		Action:    d.Move,
	}
	if err := settings.Validate(); err != nil {
		return domain.WarningSettings{}, err
	}
	return settings, nil
}
