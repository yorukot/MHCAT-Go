package documents

import (
	"fmt"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type AutoNotificationScheduleDocument struct {
	Guild   string        `bson:"guild" json:"guild"`
	Channel bson.RawValue `bson:"channel" json:"channel"`
	ID      bson.RawValue `bson:"id" json:"id"`
	Cron    bson.RawValue `bson:"cron" json:"cron"`
}

type AutoNotificationDeliveryDocument struct {
	Guild   string        `bson:"guild" json:"guild"`
	Channel bson.RawValue `bson:"channel" json:"channel"`
	ID      bson.RawValue `bson:"id" json:"id"`
	Cron    bson.RawValue `bson:"cron" json:"cron"`
	Message bson.RawValue `bson:"message" json:"message"`
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
	id, _ := legacyMongooseString(d.ID)
	channel, _ := legacyMongooseString(d.Channel)
	cron, pending := legacyNullableString(d.Cron)
	return domain.AutoNotificationSchedule{
		GuildID:   d.Guild,
		ID:        id,
		Cron:      cron,
		ChannelID: channel,
		Pending:   pending,
	}
}

func (d AutoNotificationDeliveryDocument) ToDomain() domain.AutoNotificationSchedule {
	id, _ := legacyMongooseString(d.ID)
	channel, _ := legacyMongooseString(d.Channel)
	cron, pending := legacyNullableString(d.Cron)
	message := legacyAutoNotificationMessage(d.Message)
	return domain.AutoNotificationSchedule{
		GuildID:   d.Guild,
		ID:        id,
		Cron:      cron,
		ChannelID: channel,
		Message:   message,
		Pending:   pending,
	}
}

func legacyAutoNotificationMessage(value bson.RawValue) domain.AutoNotificationMessage {
	document, ok := value.DocumentOK()
	if !ok {
		return domain.AutoNotificationMessage{}
	}
	message := domain.AutoNotificationMessage{Content: legacyAutoNotificationMixedString(document.Lookup("content"))}
	embeds, ok := document.Lookup("embeds").ArrayOK()
	if !ok {
		return message
	}
	values, err := embeds.Values()
	if err != nil || len(values) == 0 {
		return message
	}
	envelope, ok := values[0].DocumentOK()
	if !ok {
		return message
	}
	data, ok := envelope.Lookup("data").DocumentOK()
	if !ok {
		return message
	}
	message.EmbedTitle = legacyAutoNotificationMixedString(data.Lookup("title"))
	message.EmbedDescription = legacyAutoNotificationMixedString(data.Lookup("description"))
	message.EmbedColor = legacyAutoNotificationColor(data.Lookup("color"))
	return message
}

func legacyAutoNotificationMixedString(value bson.RawValue) string {
	if text, ok := value.StringValueOK(); ok {
		return text
	}
	if text, ok := value.SymbolOK(); ok {
		return text
	}
	return ""
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
