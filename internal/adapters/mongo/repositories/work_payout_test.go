package repositories

import (
	"math"
	"strings"
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

func TestWorkPayoutStateResetFilterTargetsExactJobSnapshot(t *testing.T) {
	document := workUserPayoutDocument{
		ID:      "id-1",
		Guild:   "guild",
		User:    "user",
		State:   "job-a",
		EndTime: rawValue(t, int64(100)),
		GetCoin: rawValue(t, int64(50)),
	}
	filter := workPayoutStateResetFilter(document, 123)
	if lookupD(filter, "_id") != "id-1" {
		t.Fatalf("expected _id filter, got %#v", filter)
	}
	if lookupD(filter, "guild") != "guild" || lookupD(filter, "user") != "user" || lookupD(filter, "state") != "job-a" {
		t.Fatalf("identity snapshot = %#v", filter)
	}
	endTime, ok := lookupD(filter, "end_time").(bson.D)
	endTimeEqual, ok := lookupD(endTime, "$eq").(bson.RawValue)
	if !ok || rawInt64(endTimeEqual) != int64(100) || lookupD(endTime, "$lt") != int64(123) {
		t.Fatalf("end_time snapshot = %#v", lookupD(filter, "end_time"))
	}
	getCoin, ok := lookupD(filter, "get_coin").(bson.RawValue)
	if !ok || rawInt64(getCoin) != int64(50) {
		t.Fatalf("get_coin snapshot = %#v", lookupD(filter, "get_coin"))
	}
}

func TestWorkPayoutIdentityIsDeterministicAndVersionedPerWorkRow(t *testing.T) {
	document := validWorkPayoutTestDocument(t)
	first, err := newWorkPayoutIdentity(document)
	if err != nil {
		t.Fatalf("first identity: %v", err)
	}
	second, err := newWorkPayoutIdentity(document)
	if err != nil {
		t.Fatalf("second identity: %v", err)
	}
	if first != second {
		t.Fatalf("identity is not deterministic: %#v != %#v", first, second)
	}
	if !strings.HasPrefix(first.MarkerKey, "v1_") || !strings.HasPrefix(first.Token, "v1_") {
		t.Fatalf("identity is not versioned: %#v", first)
	}
	changedJob := document
	changedJob.GetCoin = rawValue(t, int64(51))
	changedIdentity, err := newWorkPayoutIdentity(changedJob)
	if err != nil {
		t.Fatalf("changed identity: %v", err)
	}
	if changedIdentity.MarkerKey != first.MarkerKey || changedIdentity.Token == first.Token {
		t.Fatalf("job version must change token but retain work-row marker: first=%#v changed=%#v", first, changedIdentity)
	}
	typeChangedJob := document
	typeChangedJob.GetCoin = rawValue(t, "50")
	typeChangedIdentity, err := newWorkPayoutIdentity(typeChangedJob)
	if err != nil {
		t.Fatalf("type-changed identity: %v", err)
	}
	if typeChangedIdentity.Token == first.Token {
		t.Fatalf("raw BSON type is part of the job version: first=%#v changed=%#v", first, typeChangedIdentity)
	}
	duplicateRow := document
	duplicateRow.ID = "id-2"
	duplicateIdentity, err := newWorkPayoutIdentity(duplicateRow)
	if err != nil {
		t.Fatalf("duplicate-row identity: %v", err)
	}
	if duplicateIdentity.MarkerKey == first.MarkerKey || duplicateIdentity.Token == first.Token {
		t.Fatalf("duplicate work rows must have independent identities: first=%#v duplicate=%#v", first, duplicateIdentity)
	}
}

func TestWorkPayoutCoinIDIsDeterministic(t *testing.T) {
	first, err := newWorkPayoutCoinID("guild", "user")
	if err != nil {
		t.Fatalf("first id: %v", err)
	}
	second, err := newWorkPayoutCoinID("guild", "user")
	if err != nil {
		t.Fatalf("second id: %v", err)
	}
	other, err := newWorkPayoutCoinID("guild", "other")
	if err != nil {
		t.Fatalf("other id: %v", err)
	}
	if first != second || first == other || first.IsZero() {
		t.Fatalf("deterministic ids: first=%s second=%s other=%s", first.Hex(), second.Hex(), other.Hex())
	}
}

func TestWorkPayoutCoinFilterAllowsOnlyMissingSameOrNewerMarker(t *testing.T) {
	identity := workPayoutIdentity{MarkerKey: "v1_key", Token: "v1_token", EndTime: 123, Reward: 50}
	filter := workPayoutCoinFilter("coin-id", "guild", "user", identity)
	if lookupD(filter, "_id") != "coin-id" || lookupD(filter, "guild") != "guild" || lookupD(filter, "member") != "user" {
		t.Fatalf("coin identity filter = %#v", filter)
	}
	combined, ok := lookupD(filter, "$and").(bson.A)
	if !ok || len(combined) != 2 {
		t.Fatalf("combined guards = %#v", lookupD(filter, "$and"))
	}
	markerGuard, ok := combined[0].(bson.D)
	if !ok {
		t.Fatalf("marker guard = %#v", combined[0])
	}
	guards, ok := lookupD(markerGuard, "$or").(bson.A)
	if !ok || len(guards) != 3 {
		t.Fatalf("marker guards = %#v", markerGuard)
	}
	markerPath := WorkPayoutMarkerField + "." + identity.MarkerKey
	missing := guards[0].(bson.D)
	missingCondition, ok := lookupD(missing, markerPath).(bson.D)
	if !ok || lookupD(missingCondition, "$exists") != false {
		t.Fatalf("missing marker guard = %#v", missing)
	}
	same := guards[1].(bson.D)
	if lookupD(same, markerPath+".token") != identity.Token {
		t.Fatalf("same-token guard = %#v", same)
	}
	newer := guards[2].(bson.D)
	newerCondition, ok := lookupD(newer, markerPath+".end_time").(bson.D)
	if !ok || lookupD(newerCondition, "$lt") != identity.EndTime {
		t.Fatalf("newer-job guard = %#v", newer)
	}
	coinGuard, ok := combined[1].(bson.D)
	if !ok {
		t.Fatalf("coin guard = %#v", combined[1])
	}
	coinChoices, ok := lookupD(coinGuard, "$or").(bson.A)
	if !ok || len(coinChoices) != 2 {
		t.Fatalf("coin choices = %#v", coinGuard)
	}
	numericCoin := coinChoices[0].(bson.D)
	typeCondition, ok := lookupD(numericCoin, "coin").(bson.D)
	if !ok || lookupD(typeCondition, "$type") != "number" {
		t.Fatalf("numeric coin guard = %#v", numericCoin)
	}
}

func TestWorkPayoutCoinPipelineWritesMarkerWithConditionalIncrement(t *testing.T) {
	identity := workPayoutIdentity{MarkerKey: "v1_key", Token: "v1_token", EndTime: 123, Reward: 50}
	pipeline := workPayoutCoinPipeline("guild", "user", 1, identity)
	if len(pipeline) != 1 {
		t.Fatalf("pipeline = %#v", pipeline)
	}
	set, ok := lookupD(pipeline[0], "$set").(bson.D)
	if !ok {
		t.Fatalf("set stage = %#v", lookupD(pipeline[0], "$set"))
	}
	markerPath := WorkPayoutMarkerField + "." + identity.MarkerKey
	marker, ok := lookupD(set, markerPath).(bson.D)
	if !ok || lookupD(marker, "token") != identity.Token || lookupD(marker, "end_time") != identity.EndTime {
		t.Fatalf("marker update = %#v", lookupD(set, markerPath))
	}
	coin, ok := lookupD(set, "coin").(bson.D)
	if !ok {
		t.Fatalf("coin expression = %#v", lookupD(set, "coin"))
	}
	conditional, ok := lookupD(coin, "$cond").(bson.A)
	if !ok || len(conditional) != 3 {
		t.Fatalf("coin condition = %#v", coin)
	}
	sameToken, ok := conditional[0].(bson.D)
	if !ok {
		t.Fatalf("same-token expression = %#v", conditional[0])
	}
	equality, ok := lookupD(sameToken, "$eq").(bson.A)
	if !ok || len(equality) != 2 || equality[0] != "$"+markerPath+".token" || equality[1] != identity.Token {
		t.Fatalf("same-token equality = %#v", sameToken)
	}
	if conditional[1] != "$coin" {
		t.Fatalf("replay branch must preserve coin: %#v", conditional[1])
	}
	add, ok := conditional[2].(bson.D)
	if !ok {
		t.Fatalf("increment branch = %#v", conditional[2])
	}
	operands, ok := lookupD(add, "$add").(bson.A)
	if !ok || len(operands) != 2 || operands[1] != identity.Reward {
		t.Fatalf("increment operands = %#v", add)
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
	valid := validWorkPayoutTestDocument(t)
	if !validWorkPayoutDocument(valid, 123) {
		t.Fatalf("expected valid document")
	}
	invalid := valid
	invalid.State = LegacyIdleWorkState
	if validWorkPayoutDocument(invalid, 123) {
		t.Fatalf("idle work state must be invalid")
	}
	invalid = valid
	invalid.EndTime = rawValue(t, int64(123))
	if validWorkPayoutDocument(invalid, 123) {
		t.Fatalf("equal end_time is not strictly due")
	}
	zeroCoin := valid
	zeroCoin.GetCoin = rawValue(t, int64(0))
	if !validWorkPayoutDocument(zeroCoin, 123) {
		t.Fatalf("zero get_coin is legacy-compatible and must not invalidate the row")
	}
	missingID := valid
	missingID.ID = nil
	if validWorkPayoutDocument(missingID, 123) {
		t.Fatalf("missing _id cannot produce a safe payout identity")
	}
}

func TestWorkPayoutScalarCoercionAndDueComparison(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    float64
		wantDue bool
	}{
		{name: "decimal", value: 0.5, want: 0.5, wantDue: true},
		{name: "negative", value: -1.5, want: -1.5, wantDue: true},
		{name: "null", value: nil, want: 0, wantDue: true},
		{name: "infinity", value: math.Inf(1), want: math.Inf(1), wantDue: false},
		{name: "nan", value: math.NaN(), want: math.NaN(), wantDue: false},
		{name: "malformed", value: "bad", want: math.NaN(), wantDue: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			raw := rawValue(t, test.value)
			got := workPayoutNumber(raw)
			if !(got == test.want || math.IsNaN(got) && math.IsNaN(test.want)) {
				t.Fatalf("number = %v, want %v", got, test.want)
			}
			document := validWorkPayoutTestDocument(t)
			document.EndTime = raw
			if gotDue := validWorkPayoutDocument(document, 1); gotDue != test.wantDue {
				t.Fatalf("due = %t, want %t", gotDue, test.wantDue)
			}
		})
	}
}

func validWorkPayoutTestDocument(t *testing.T) workUserPayoutDocument {
	t.Helper()
	return workUserPayoutDocument{
		ID:      "id-1",
		Guild:   "guild",
		User:    "user",
		State:   "working",
		EndTime: rawValue(t, int64(100)),
		GetCoin: rawValue(t, int64(50)),
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
