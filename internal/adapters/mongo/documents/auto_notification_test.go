package documents

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestAutoNotificationScheduleDocumentDecodesActiveSchedule(t *testing.T) {
	raw, err := bson.Marshal(bson.D{
		{Key: "guild", Value: "guild-1"},
		{Key: "id", Value: "schedule-1"},
		{Key: "cron", Value: "*/30 * * * *"},
		{Key: "channel", Value: "channel-1"},
		{Key: "message", Value: bson.D{{Key: "content", Value: "hello"}}},
	})
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	var document AutoNotificationScheduleDocument
	if err := bson.Unmarshal(raw, &document); err != nil {
		t.Fatalf("decode document: %v", err)
	}
	schedule := document.ToDomain()
	if schedule.GuildID != "guild-1" || schedule.ID != "schedule-1" || schedule.Cron != "*/30 * * * *" || schedule.ChannelID != "channel-1" || schedule.Pending {
		t.Fatalf("schedule = %#v", schedule)
	}
}

func TestAutoNotificationScheduleDocumentDecodesNullCronAsPending(t *testing.T) {
	raw, err := bson.Marshal(bson.D{
		{Key: "guild", Value: "guild-1"},
		{Key: "id", Value: "draft-1"},
		{Key: "cron", Value: nil},
		{Key: "channel", Value: "channel-1"},
	})
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	var document AutoNotificationScheduleDocument
	if err := bson.Unmarshal(raw, &document); err != nil {
		t.Fatalf("decode document: %v", err)
	}
	schedule := document.ToDomain()
	if !schedule.Pending || schedule.Cron != "" {
		t.Fatalf("schedule = %#v", schedule)
	}
}

func TestAutoNotificationScheduleDocumentUsesMongooseStringCoercion(t *testing.T) {
	for _, test := range []struct {
		value any
		want  string
	}{
		{value: true, want: "true"},
		{value: false, want: "false"},
		{value: 0.000001, want: "0.000001"},
		{value: 0.0000001, want: "1e-7"},
		{value: 1e21, want: "1e+21"},
	} {
		raw, err := bson.Marshal(bson.D{{Key: "guild", Value: "guild-1"}, {Key: "id", Value: "schedule-1"}, {Key: "channel", Value: "channel-1"}, {Key: "cron", Value: test.value}})
		if err != nil {
			t.Fatalf("marshal %T: %v", test.value, err)
		}
		var document AutoNotificationScheduleDocument
		if err := bson.Unmarshal(raw, &document); err != nil {
			t.Fatalf("decode %T: %v", test.value, err)
		}
		schedule := document.ToDomain()
		if schedule.Pending || schedule.Cron != test.want {
			t.Fatalf("value %#v decoded as %#v, want cron %q", test.value, schedule, test.want)
		}
	}
}

func TestAutoNotificationDocumentsCoerceLegacyScalarIdentityFields(t *testing.T) {
	raw, err := bson.Marshal(bson.D{
		{Key: "guild", Value: "guild-1"},
		{Key: "id", Value: int64(123)},
		{Key: "cron", Value: bson.Binary{Data: []byte("*/30 * * * *")}},
		{Key: "channel", Value: bson.ObjectID{0x50, 0x7f, 0x1f, 0x77, 0xbc, 0xf8, 0x6c, 0xd7, 0x99, 0x43, 0x90, 0x11}},
		{Key: "message", Value: bson.D{{Key: "content", Value: "hello"}}},
	})
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}

	var scheduleDocument AutoNotificationScheduleDocument
	if err := bson.Unmarshal(raw, &scheduleDocument); err != nil {
		t.Fatalf("decode schedule: %v", err)
	}
	schedule := scheduleDocument.ToDomain()
	if schedule.ID != "123" || schedule.Cron != "*/30 * * * *" || schedule.ChannelID != "507f1f77bcf86cd799439011" || schedule.Pending {
		t.Fatalf("schedule = %#v", schedule)
	}

	var deliveryDocument AutoNotificationDeliveryDocument
	if err := bson.Unmarshal(raw, &deliveryDocument); err != nil {
		t.Fatalf("decode delivery: %v", err)
	}
	delivery := deliveryDocument.ToDomain()
	if delivery.ID != schedule.ID || delivery.Cron != schedule.Cron || delivery.ChannelID != schedule.ChannelID || delivery.Message.Content != "hello" {
		t.Fatalf("delivery = %#v", delivery)
	}
}

func TestAutoNotificationDocumentsKeepUncastableIdentityFieldsRowLocal(t *testing.T) {
	raw, err := bson.Marshal(bson.D{
		{Key: "guild", Value: "guild-1"},
		{Key: "id", Value: bson.A{"invalid"}},
		{Key: "cron", Value: bson.D{{Key: "invalid", Value: true}}},
		{Key: "channel", Value: bson.D{{Key: "invalid", Value: true}}},
		{Key: "message", Value: bson.D{{Key: "content", Value: "hello"}}},
	})
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	var document AutoNotificationDeliveryDocument
	if err := bson.Unmarshal(raw, &document); err != nil {
		t.Fatalf("decode delivery: %v", err)
	}
	schedule := document.ToDomain()
	if schedule.ID != "" || schedule.Cron != "" || schedule.ChannelID != "" || schedule.Pending {
		t.Fatalf("schedule = %#v", schedule)
	}
}

func TestAutoNotificationMessageBSONPreservesLegacyEmbedDataShape(t *testing.T) {
	payload := AutoNotificationMessageBSON(domain.AutoNotificationMessage{
		Content:          "hello",
		EmbedTitle:       "Title",
		EmbedDescription: "Content",
		EmbedColor:       "#123456",
	})
	raw, err := bson.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	var decoded struct {
		Content *string `bson:"content"`
		Embeds  []struct {
			Data struct {
				Title       string `bson:"title"`
				Description string `bson:"description"`
				Color       int32  `bson:"color"`
			} `bson:"data"`
		} `bson:"embeds"`
	}
	if err := bson.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if decoded.Content == nil || *decoded.Content != "hello" || len(decoded.Embeds) != 1 {
		t.Fatalf("decoded = %#v", decoded)
	}
	if decoded.Embeds[0].Data.Title != "Title" || decoded.Embeds[0].Data.Description != "Content" || decoded.Embeds[0].Data.Color != 0x123456 {
		t.Fatalf("embed data = %#v", decoded.Embeds[0].Data)
	}
}

func TestAutoNotificationMessageBSONPreservesLegacyWhitespace(t *testing.T) {
	payload := AutoNotificationMessageBSON(domain.AutoNotificationMessage{
		Content:          "   ",
		EmbedTitle:       "  Title  ",
		EmbedDescription: "  Content  ",
	})
	raw, err := bson.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if content := bson.Raw(raw).Lookup("content"); content.Type != bson.TypeString || content.StringValue() != "   " {
		t.Fatalf("content = %#v", content)
	}
	var decoded AutoNotificationMessageDocument
	if err := bson.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if len(decoded.Embeds) != 1 || decoded.Embeds[0].Data.Title.StringValue() != "  Title  " || decoded.Embeds[0].Data.Description.StringValue() != "  Content  " {
		t.Fatalf("decoded = %#v", decoded)
	}
}

func TestAutoNotificationDeliveryDocumentDecodesLegacyNumericEmbedColor(t *testing.T) {
	raw, err := bson.Marshal(bson.D{
		{Key: "guild", Value: "guild-1"},
		{Key: "id", Value: "schedule-1"},
		{Key: "cron", Value: "0 9 * * 1"},
		{Key: "channel", Value: "channel-1"},
		{Key: "message", Value: bson.D{
			{Key: "content", Value: "hello"},
			{Key: "embeds", Value: bson.A{bson.D{{Key: "data", Value: bson.D{
				{Key: "title", Value: "Title"},
				{Key: "description", Value: "Content"},
				{Key: "color", Value: int32(0xA6FFA6)},
			}}}}},
		}},
	})
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	var document AutoNotificationDeliveryDocument
	if err := bson.Unmarshal(raw, &document); err != nil {
		t.Fatalf("decode document: %v", err)
	}
	message := document.ToDomain().Message
	if message.Content != "hello" || message.EmbedTitle != "Title" || message.EmbedDescription != "Content" || message.EmbedColor != "#A6FFA6" {
		t.Fatalf("message = %#v", message)
	}
}

func TestAutoNotificationDeliveryDocumentKeepsMixedMessageTypesRowLocal(t *testing.T) {
	for _, test := range []struct {
		name        string
		message     any
		wantContent string
	}{
		{
			name: "malformed embeds retain content",
			message: bson.D{
				{Key: "content", Value: "hello"},
				{Key: "embeds", Value: bson.D{{Key: "not", Value: "an array"}}},
			},
			wantContent: "hello",
		},
		{name: "numeric nested content is not mongoose cast", message: bson.D{{Key: "content", Value: int32(123)}}},
		{name: "scalar message is unusable", message: "hello"},
	} {
		t.Run(test.name, func(t *testing.T) {
			raw, err := bson.Marshal(bson.D{
				{Key: "guild", Value: "guild-1"},
				{Key: "id", Value: "schedule-1"},
				{Key: "cron", Value: "0 9 * * 1"},
				{Key: "channel", Value: "channel-1"},
				{Key: "message", Value: test.message},
			})
			if err != nil {
				t.Fatalf("marshal fixture: %v", err)
			}
			var document AutoNotificationDeliveryDocument
			if err := bson.Unmarshal(raw, &document); err != nil {
				t.Fatalf("decode document: %v", err)
			}
			message := document.ToDomain().Message
			if message.Content != test.wantContent || message.HasEmbed() {
				t.Fatalf("message = %#v", message)
			}
		})
	}
}
