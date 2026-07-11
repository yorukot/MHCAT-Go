package repositories

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestAutoNotificationScheduleCollectionName(t *testing.T) {
	if AutoNotificationScheduleCollectionName != "cron_sets" {
		t.Fatalf("auto-notification schedule collection = %s, want cron_sets", AutoNotificationScheduleCollectionName)
	}
}

func TestNewAutoNotificationScheduleRepositoryRequiresCollection(t *testing.T) {
	if _, err := NewAutoNotificationScheduleRepository(nil); err == nil {
		t.Fatal("expected nil collection error")
	}
}

func TestNewAutoNotificationScheduleRepositoryFromDatabaseRequiresDatabase(t *testing.T) {
	if _, err := NewAutoNotificationScheduleRepositoryFromDatabase(nil); err == nil {
		t.Fatal("expected nil database error")
	}
}

func TestAutoNotificationDeliveryFilterRequiresActiveLegacyPayloadShape(t *testing.T) {
	filter := autoNotificationDeliveryFilter()
	if len(filter) != 3 {
		t.Fatalf("filter = %#v", filter)
	}
	guildType, ok := autoNotificationFilterValue(t, filter, "guild").(bson.D)
	if !ok || autoNotificationFilterValue(t, guildType, "$type") != "string" {
		t.Fatalf("guild filter = %#v", guildType)
	}
	cronValue, ok := autoNotificationFilterValue(t, filter, "cron").(bson.D)
	if !ok || autoNotificationFilterValue(t, cronValue, "$ne") != nil {
		t.Fatalf("cron filter = %#v", cronValue)
	}
	messageType, ok := autoNotificationFilterValue(t, filter, "message").(bson.D)
	if !ok || autoNotificationFilterValue(t, messageType, "$type") != "object" {
		t.Fatalf("message filter = %#v", messageType)
	}
}

func autoNotificationFilterValue(t *testing.T, document bson.D, key string) any {
	t.Helper()
	for _, element := range document {
		if element.Key == key {
			return element.Value
		}
	}
	t.Fatalf("missing key %q in %#v", key, document)
	return nil
}
