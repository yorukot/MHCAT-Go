package repositories

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestNewAutoChatPaidRepositoryRequiresDependencies(t *testing.T) {
	runner := &fakemongo.TransactionRunner{}
	if _, err := NewAutoChatPaidRepository(nil, nil, runner); err == nil {
		t.Fatal("expected missing balance collection error")
	}
	if _, err := NewAutoChatPaidRepositoryFromDatabase(nil, runner); err == nil {
		t.Fatal("expected missing database error")
	}
}

func TestAutoChatPaidNumericAcceptsLegacyNumberTypes(t *testing.T) {
	decimal, err := bson.ParseDecimal128("5.5")
	if err != nil {
		t.Fatalf("parse decimal: %v", err)
	}
	for _, value := range []any{int32(2), int64(3), 4.5, "6.5", true, time.UnixMilli(7), decimal} {
		typeValue, encoded, err := bson.MarshalValue(value)
		if err != nil {
			t.Fatalf("marshal %T: %v", value, err)
		}
		if parsed, ok := autoChatPaidNumeric(bson.RawValue{Type: typeValue, Value: encoded}); !ok || parsed <= 0 {
			t.Fatalf("value %v parsed as %f ok=%v", value, parsed, ok)
		}
	}
}

func TestAutoChatPaidNumericDoesNotTreatUnavailableValuesAsPositive(t *testing.T) {
	for _, value := range []any{0, -1, "not-a-number", false, math.NaN(), math.Inf(1), bson.D{{Key: "invalid", Value: 1}}} {
		parsed, ok := autoChatPaidNumeric(rawValue(t, value))
		if ok && parsed > 0 {
			t.Fatalf("value %#v parsed as %f", value, parsed)
		}
	}
}

func TestAutoChatPaidBalanceFilterPreservesRawPriceType(t *testing.T) {
	price := rawValue(t, "12.5")
	filter := bson.D{{Key: "_id", Value: bson.NewObjectID()}, {Key: "guild", Value: "guild-1"}, {Key: "price", Value: price}}
	encoded, err := bson.Marshal(filter)
	if err != nil {
		t.Fatalf("marshal filter: %v", err)
	}
	stored := bson.Raw(encoded).Lookup("price")
	if stored.Type != bson.TypeString || stored.StringValue() != "12.5" {
		t.Fatalf("stored price = %#v", stored)
	}
}

func TestNewAutoChatPaidIDIsStableAndGuildScoped(t *testing.T) {
	first := newAutoChatPaidID("guild-1")
	if first.IsZero() || first != newAutoChatPaidID("guild-1") || first == newAutoChatPaidID("guild-2") {
		t.Fatalf("ids first=%s same=%s other=%s", first.Hex(), newAutoChatPaidID("guild-1").Hex(), newAutoChatPaidID("guild-2").Hex())
	}
}

func TestAutoChatPaidRequestValidationRejectsInvalidValues(t *testing.T) {
	for _, request := range []domain.AutoChatPaidRequest{
		{},
		{GuildID: "guild-1", RequestedAtMilli: 1, Cost: -1},
	} {
		if _, err := (&AutoChatPaidRepository{}).QueuePaidAutoChat(context.Background(), request); err == nil {
			t.Fatalf("request %#v should fail", request)
		}
	}
}

func TestLegacyAutoChatTimingPreservesJavaScriptBoundaries(t *testing.T) {
	const now = int64(1_700_000_000_000)
	for _, test := range []struct {
		name      string
		previous  any
		wantBusy  bool
		wantReset bool
	}{
		{name: "under busy boundary", previous: float64(now) - 9_999.5, wantBusy: true},
		{name: "at busy boundary", previous: now - 10_000},
		{name: "at reset boundary", previous: now - 40_000},
		{name: "over reset boundary", previous: float64(now) - 40_000.5, wantReset: true},
		{name: "numeric string", previous: "1699999990001", wantBusy: true},
		{name: "null resets", previous: nil, wantReset: true},
		{name: "positive infinity stays busy", previous: math.Inf(1), wantBusy: true},
		{name: "nan preserves conversation", previous: math.NaN()},
		{name: "malformed preserves conversation", previous: bson.D{{Key: "invalid", Value: 1}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			busy, reset := legacyAutoChatTiming(now, rawValue(t, test.previous))
			if busy != test.wantBusy || reset != test.wantReset {
				t.Fatalf("timing = busy %v reset %v; want busy %v reset %v", busy, reset, test.wantBusy, test.wantReset)
			}
		})
	}
}
