package documents

import (
	"strconv"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type AutoNotificationScheduleDocument struct {
	Guild   string        `bson:"guild" json:"guild"`
	Channel string        `bson:"channel" json:"channel"`
	ID      string        `bson:"id" json:"id"`
	Cron    bson.RawValue `bson:"cron" json:"cron"`
}

type AutoNotificationPendingWriteDocument struct {
	Guild   string `bson:"guild"`
	Channel string `bson:"channel"`
	ID      string `bson:"id"`
	Cron    any    `bson:"cron"`
	Message any    `bson:"message"`
}

func (d AutoNotificationScheduleDocument) ToDomain() domain.AutoNotificationSchedule {
	cron, pending := legacyNullableString(d.Cron)
	return domain.AutoNotificationSchedule{
		GuildID:   d.Guild,
		ID:        d.ID,
		Cron:      cron,
		ChannelID: d.Channel,
		Pending:   pending,
	}
}

func AutoNotificationPendingWriteDocumentFromDomain(draft domain.AutoNotificationSetupDraft) AutoNotificationPendingWriteDocument {
	draft = draft.Normalized()
	return AutoNotificationPendingWriteDocument{
		Guild:   draft.GuildID,
		Channel: draft.ChannelID,
		ID:      draft.ID,
		Cron:    nil,
		Message: nil,
	}
}

func AutoNotificationMessageBSON(message domain.AutoNotificationMessage) bson.D {
	message = message.Normalized()
	if !message.HasEmbed() {
		return bson.D{{Key: "content", Value: message.Content}}
	}
	embedData := bson.D{}
	if message.EmbedTitle != "" {
		embedData = append(embedData, bson.E{Key: "title", Value: message.EmbedTitle})
	}
	if message.EmbedDescription != "" {
		embedData = append(embedData, bson.E{Key: "description", Value: message.EmbedDescription})
	}
	if message.EmbedColor != "" {
		embedData = append(embedData, bson.E{Key: "color", Value: message.EmbedColor})
	}
	return bson.D{
		{Key: "content", Value: nullableString(message.Content)},
		{Key: "embeds", Value: bson.A{bson.D{{Key: "data", Value: embedData}}}},
	}
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func legacyNullableString(value bson.RawValue) (string, bool) {
	if value.Type == 0 || value.Type == bson.TypeNull || value.Type == bson.TypeUndefined {
		return "", true
	}
	if text, ok := value.StringValueOK(); ok {
		return text, false
	}
	if parsed, ok := value.AsInt64OK(); ok {
		return strconv.FormatInt(parsed, 10), false
	}
	if parsed, ok := value.DoubleOK(); ok {
		return strings.TrimRight(strings.TrimRight(strconv.FormatFloat(parsed, 'f', 6, 64), "0"), "."), false
	}
	return "", false
}
