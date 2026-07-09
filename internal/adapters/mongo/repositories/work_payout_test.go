package repositories

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestNewWorkPayoutRepositoryRequiresCollections(t *testing.T) {
	if _, err := NewWorkPayoutRepository(nil, nil, nil); err == nil {
		t.Fatal("expected nil collection error")
	}
}

func TestNewWorkPayoutRepositoryFromDatabaseRequiresDatabase(t *testing.T) {
	if _, err := NewWorkPayoutRepositoryFromDatabase(nil); err == nil {
		t.Fatal("expected nil database error")
	}
}

func TestWorkPayoutEligibleFilterUsesEffectiveLegacyGuard(t *testing.T) {
	filter := workPayoutEligibleFilter(123)
	state, ok := lookupD(filter, "state").(bson.D)
	if !ok || lookupD(state, "$ne") != LegacyIdleWorkState {
		t.Fatalf("state filter = %#v", lookupD(filter, "state"))
	}
	endTime, ok := lookupD(filter, "end_time").(bson.D)
	if !ok || lookupD(endTime, "$lt") != int64(123) {
		t.Fatalf("end_time filter = %#v", lookupD(filter, "end_time"))
	}
}

func TestWorkPayoutStateResetFilterTargetsDocumentIDWhenAvailable(t *testing.T) {
	filter := workPayoutStateResetFilter(workUserPayoutDocument{ID: "id-1", Guild: "guild", User: "user"}, 123)
	if lookupD(filter, "_id") != "id-1" {
		t.Fatalf("expected _id filter, got %#v", filter)
	}
	if lookupD(filter, "guild") != nil || lookupD(filter, "user") != nil {
		t.Fatalf("did not expect guild/user fallback with _id: %#v", filter)
	}
}

func TestWorkPayoutStateResetFilterFallsBackToGuildUser(t *testing.T) {
	filter := workPayoutStateResetFilter(workUserPayoutDocument{Guild: "guild", User: "user"}, 123)
	if lookupD(filter, "guild") != "guild" || lookupD(filter, "user") != "user" {
		t.Fatalf("expected guild/user filter, got %#v", filter)
	}
}

func TestWorkPayoutTodayFromConfig(t *testing.T) {
	if got := workPayoutTodayFromConfig(false, 0, 999); got != 1 {
		t.Fatalf("missing config today = %d", got)
	}
	if got := workPayoutTodayFromConfig(true, 0, 999); got != 1 {
		t.Fatalf("daily config today = %d", got)
	}
	if got := workPayoutTodayFromConfig(true, 3600, 999); got != 999 {
		t.Fatalf("rolling config today = %d", got)
	}
}

func TestValidWorkPayoutDocument(t *testing.T) {
	valid := workUserPayoutDocument{
		Guild:   "guild",
		User:    "user",
		State:   "working",
		EndTime: rawValue(t, int64(100)),
		GetCoin: rawValue(t, int64(50)),
	}
	if !validWorkPayoutDocument(valid) {
		t.Fatalf("expected valid document")
	}
	invalid := valid
	invalid.State = LegacyIdleWorkState
	if validWorkPayoutDocument(invalid) {
		t.Fatalf("idle work state must be invalid")
	}
	invalid = valid
	invalid.EndTime = rawValue(t, int64(0))
	if validWorkPayoutDocument(invalid) {
		t.Fatalf("zero end_time must be invalid")
	}
	zeroCoin := valid
	zeroCoin.GetCoin = rawValue(t, int64(0))
	if !validWorkPayoutDocument(zeroCoin) {
		t.Fatalf("zero get_coin is legacy-compatible and must not invalidate the row")
	}
}

func lookupD(document bson.D, key string) any {
	for _, element := range document {
		if element.Key == key {
			return element.Value
		}
	}
	return nil
}
