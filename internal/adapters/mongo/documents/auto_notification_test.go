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

func TestAutoNotificationMessageBSONPreservesLegacyEmbedDataShape(t *testing.T) {
	payload := AutoNotificationMessageBSON(domain.AutoNotificationMessage{
		Content:          "hello",
		EmbedTitle:       "Title",
		EmbedDescription: "Content",
		EmbedColor:       "Random",
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
				Color       string `bson:"color"`
			} `bson:"data"`
		} `bson:"embeds"`
	}
	if err := bson.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if decoded.Content == nil || *decoded.Content != "hello" || len(decoded.Embeds) != 1 {
		t.Fatalf("decoded = %#v", decoded)
	}
	if decoded.Embeds[0].Data.Title != "Title" || decoded.Embeds[0].Data.Description != "Content" || decoded.Embeds[0].Data.Color != "Random" {
		t.Fatalf("embed data = %#v", decoded.Embeds[0].Data)
	}
}
