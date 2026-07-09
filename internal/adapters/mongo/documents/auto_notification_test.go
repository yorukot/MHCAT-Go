package documents

import (
	"testing"

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
