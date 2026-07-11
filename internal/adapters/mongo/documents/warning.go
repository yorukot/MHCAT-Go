package documents

import (
	"fmt"
	"strconv"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
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

type WarningSettingsReadDocument struct {
	Guild    bson.RawValue `bson:"guild" json:"guild"`
	BanCount bson.RawValue `bson:"ban_count" json:"ban_count"`
	Move     bson.RawValue `bson:"move" json:"move"`
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
	return warningSettingsToDomain(d.Guild, d.BanCount, d.Move)
}

func (d WarningSettingsReadDocument) ToDomain() (domain.WarningSettings, error) {
	guild, _ := legacyMongooseString(d.Guild)
	banCount, _ := legacyMongooseString(d.BanCount)
	move, _ := legacyMongooseString(d.Move)
	return warningSettingsToDomain(guild, banCount, move)
}

func warningSettingsToDomain(guild string, banCount string, move string) (domain.WarningSettings, error) {
	threshold, err := strconv.ParseInt(banCount, 10, 64)
	if err != nil {
		return domain.WarningSettings{}, fmt.Errorf("%w: ban_count", domain.ErrInvalidWarningSettings)
	}
	settings := domain.WarningSettings{
		GuildID:   guild,
		Threshold: threshold,
		Action:    move,
	}
	if err := settings.Validate(); err != nil {
		return domain.WarningSettings{}, err
	}
	return settings, nil
}
