package documents

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"

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
