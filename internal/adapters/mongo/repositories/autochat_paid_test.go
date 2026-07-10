package repositories

import (
	"context"
	"testing"

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
	for _, value := range []any{int32(2), int64(3), 4.5} {
		typeValue, encoded, err := bson.MarshalValue(value)
		if err != nil {
			t.Fatalf("marshal %T: %v", value, err)
		}
		if parsed, ok := autoChatPaidNumeric(bson.RawValue{Type: typeValue, Value: encoded}); !ok || parsed <= 0 {
			t.Fatalf("value %v parsed as %f ok=%v", value, parsed, ok)
		}
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
