package documents

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestAutoChatPaidDocumentDecodesLegacyNullableConversationIDs(t *testing.T) {
	encoded, err := bson.Marshal(bson.D{
		{Key: "_id", Value: bson.NewObjectID()},
		{Key: "guild", Value: "guild-1"},
		{Key: "resid_c", Value: nil},
		{Key: "resid_p", Value: "parent-1"},
		{Key: "reply", Value: true},
		{Key: "message", Value: "answer"},
		{Key: "time", Value: float64(1_700_000_000_123)},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var document AutoChatPaidDocument
	if err := bson.Unmarshal(encoded, &document); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if document.ResidC.Type != bson.TypeNull || document.ResidP.StringValue() != "parent-1" {
		t.Fatalf("conversation ids = %#v / %#v", document.ResidC, document.ResidP)
	}
	response, ok := document.ToResponse()
	if !ok || response.GuildID != "guild-1" || response.Content != "answer" || response.RequestTimeMilli != 1_700_000_000_123 || !response.Reply {
		t.Fatalf("response = %#v ok=%v", response, ok)
	}
}

func TestLegacyExactInt64RejectsMalformedValues(t *testing.T) {
	for _, value := range []any{"1700000000123", 1.5, true} {
		typeValue, encoded, err := bson.MarshalValue(value)
		if err != nil {
			t.Fatalf("marshal %T: %v", value, err)
		}
		if parsed, ok := LegacyExactInt64(bson.RawValue{Type: typeValue, Value: encoded}); ok {
			t.Fatalf("value %v parsed as %d", value, parsed)
		}
	}
	if parsed, ok := LegacyExactInt64(bson.RawValue{Type: bson.TypeNull}); ok {
		t.Fatalf("null parsed as %d", parsed)
	}
}
