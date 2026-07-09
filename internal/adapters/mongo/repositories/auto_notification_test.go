package repositories

import "testing"

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
