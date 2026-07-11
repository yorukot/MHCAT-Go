package documents

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type WarningDocument struct {
	Guild   string                 `bson:"guild" json:"guild"`
	User    string                 `bson:"user" json:"user"`
	Content []WarningEntryDocument `bson:"content" json:"content"`
}

type WarningReadDocument struct {
	Guild   bson.RawValue `bson:"guild" json:"guild"`
	User    bson.RawValue `bson:"user" json:"user"`
	Content bson.RawValue `bson:"content" json:"content"`
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

func (d WarningReadDocument) ToDomain() domain.WarningHistory {
	guild, _ := legacyMongooseString(d.Guild)
	user, _ := legacyMongooseString(d.User)
	return domain.WarningHistory{
		GuildID: guild,
		UserID:  user,
		Entries: warningReadEntries(d.Content),
	}
}

func warningReadEntries(content bson.RawValue) []domain.WarningEntry {
	values := []bson.RawValue{}
	if array, ok := content.ArrayOK(); ok {
		decoded, err := array.Values()
		if err != nil {
			return nil
		}
		values = decoded
	} else if content.Type != 0 && content.Type != bson.TypeNull {
		// Mongoose arrays wrap a stored non-array scalar during hydration.
		values = append(values, content)
	}
	entries := make([]domain.WarningEntry, 0, len(values))
	for _, value := range values {
		entry := domain.WarningEntry{ModeratorID: "undefined", Reason: "undefined", Time: "undefined"}
		if document, ok := value.DocumentOK(); ok {
			entry.ModeratorID = warningJavaScriptString(document.Lookup("moderator"))
			entry.Reason = warningJavaScriptString(document.Lookup("reason"))
			entry.Time = warningJavaScriptString(document.Lookup("time"))
		}
		entries = append(entries, entry)
	}
	return entries
}

func warningJavaScriptString(value bson.RawValue) string {
	if value.Type == 0 {
		return "undefined"
	}
	if value.Type == bson.TypeNull {
		return "null"
	}
	if text, ok := legacyMongooseString(value); ok {
		return text
	}
	if _, ok := value.DocumentOK(); ok {
		return "[object Object]"
	}
	if array, ok := value.ArrayOK(); ok {
		values, err := array.Values()
		if err != nil {
			return ""
		}
		parts := make([]string, 0, len(values))
		for _, item := range values {
			if item.Type == 0 || item.Type == bson.TypeNull {
				parts = append(parts, "")
				continue
			}
			parts = append(parts, warningJavaScriptString(item))
		}
		return strings.Join(parts, ",")
	}
	return ""
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
