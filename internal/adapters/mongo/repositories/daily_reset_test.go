package repositories

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestDailyResetCollectionNames(t *testing.T) {
	if WorkSetCollectionName != "work_sets" {
		t.Fatalf("work_set collection = %s", WorkSetCollectionName)
	}
	if WorkUserCollectionName != "work_users" {
		t.Fatalf("work_user collection = %s", WorkUserCollectionName)
	}
}

func TestNewDailyResetRepositoryRequiresCollections(t *testing.T) {
	if _, err := NewDailyResetRepository(nil, nil, nil, nil); err == nil {
		t.Fatal("expected nil collection error")
	}
}

func TestNewDailyResetRepositoryFromDatabaseRequiresDatabase(t *testing.T) {
	if _, err := NewDailyResetRepositoryFromDatabase(nil); err == nil {
		t.Fatal("expected nil database error")
	}
}

func TestDailyCoinResetFilterWithNoExcludedGuilds(t *testing.T) {
	filter := dailyCoinResetFilter(nil)
	if len(filter) != 0 {
		t.Fatalf("expected empty filter, got %#v", filter)
	}
}

func TestDailyCoinResetFilterExcludesRollingGuilds(t *testing.T) {
	filter := dailyCoinResetFilter([]string{"guild-a", "guild-b"})
	if len(filter) != 1 || filter[0].Key != "guild" {
		t.Fatalf("unexpected filter: %#v", filter)
	}
	raw, err := bson.Marshal(filter[0].Value)
	if err != nil {
		t.Fatalf("marshal filter value: %v", err)
	}
	var decoded bson.D
	if err := bson.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal filter value: %v", err)
	}
	if len(decoded) != 1 || decoded[0].Key != "$nin" {
		t.Fatalf("expected $nin filter, got %#v", decoded)
	}
}

func TestRawInt64NormalizesLegacyNumberShapes(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  int64
	}{
		{name: "int32", value: int32(12), want: 12},
		{name: "int64", value: int64(34), want: 34},
		{name: "double", value: float64(56.7), want: 56},
		{name: "string", value: "78", want: 78},
		{name: "blank string", value: " ", want: 0},
		{name: "bool", value: true, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := rawValue(t, tt.value)
			if got := rawInt64(raw); got != tt.want {
				t.Fatalf("rawInt64(%#v) = %d, want %d", tt.value, got, tt.want)
			}
		})
	}
}

func rawValue(t *testing.T, value any) bson.RawValue {
	t.Helper()
	raw, err := bson.Marshal(bson.D{{Key: "value", Value: value}})
	if err != nil {
		t.Fatalf("marshal raw value: %v", err)
	}
	lookup := bson.Raw(raw).Lookup("value")
	return lookup
}
