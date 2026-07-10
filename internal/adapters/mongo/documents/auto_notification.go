package documents

import (
	"fmt"
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

type AutoNotificationDeliveryDocument struct {
	Guild   string                           `bson:"guild" json:"guild"`
	Channel string                           `bson:"channel" json:"channel"`
	ID      string                           `bson:"id" json:"id"`
	Cron    bson.RawValue                    `bson:"cron" json:"cron"`
	Message *AutoNotificationMessageDocument `bson:"message" json:"message"`
}

type AutoNotificationMessageDocument struct {
	Content bson.RawValue                   `bson:"content" json:"content"`
	Embeds  []AutoNotificationEmbedEnvelope `bson:"embeds" json:"embeds"`
}

type AutoNotificationEmbedEnvelope struct {
	Data AutoNotificationEmbedDocument `bson:"data" json:"data"`
}

type AutoNotificationEmbedDocument struct {
	Title       bson.RawValue `bson:"title" json:"title"`
	Description bson.RawValue `bson:"description" json:"description"`
	Color       bson.RawValue `bson:"color" json:"color"`
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

func (d AutoNotificationDeliveryDocument) ToDomain() domain.AutoNotificationSchedule {
	cron, pending := legacyNullableString(d.Cron)
	message := domain.AutoNotificationMessage{}
	if d.Message != nil {
		message.Content, _ = legacyNullableString(d.Message.Content)
		if len(d.Message.Embeds) > 0 {
			embed := d.Message.Embeds[0].Data
			message.EmbedTitle, _ = legacyNullableString(embed.Title)
			message.EmbedDescription, _ = legacyNullableString(embed.Description)
			message.EmbedColor = legacyAutoNotificationColor(embed.Color)
		}
	}
	return domain.AutoNotificationSchedule{
		GuildID:   d.Guild,
		ID:        d.ID,
		Cron:      cron,
		ChannelID: d.Channel,
		Message:   message,
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
		color := any(message.EmbedColor)
		if parsed, ok := domain.ParseLegacyColorValue(message.EmbedColor); ok {
			color = int32(parsed)
		}
		embedData = append(embedData, bson.E{Key: "color", Value: color})
	}
	return bson.D{
		{Key: "content", Value: nullableString(message.Content)},
		{Key: "embeds", Value: bson.A{bson.D{{Key: "data", Value: embedData}}}},
	}
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func legacyNullableString(value bson.RawValue) (string, bool) {
	if value.Type == 0 || value.Type == bson.TypeNull || value.Type == bson.TypeUndefined {
		return "", true
	}
	if text, ok := legacyMongooseString(value); ok {
		return text, false
	}
	return "", false
}

func legacyAutoNotificationColor(value bson.RawValue) string {
	if text, ok := value.StringValueOK(); ok {
		return strings.TrimSpace(text)
	}
	if parsed, ok := value.AsInt64OK(); ok {
		return fmt.Sprintf("#%06X", parsed&0xFFFFFF)
	}
	if parsed, ok := value.DoubleOK(); ok {
		return fmt.Sprintf("#%06X", int64(parsed)&0xFFFFFF)
	}
	return ""
}
