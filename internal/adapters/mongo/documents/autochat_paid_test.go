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

func TestAutoChatPaidDocumentCoercesLegacyScalarWorkerResponses(t *testing.T) {
	for _, test := range []struct {
		name    string
		message any
		want    string
	}{
		{name: "string", message: "answer", want: "answer"},
		{name: "boolean", message: true, want: "true"},
		{name: "integer", message: int32(123), want: "123"},
		{name: "small decimal", message: 0.000001, want: "0.000001"},
		{name: "scientific decimal", message: 0.0000001, want: "1e-7"},
		{name: "large decimal", message: 1e21, want: "1e+21"},
	} {
		t.Run(test.name, func(t *testing.T) {
			encoded, err := bson.Marshal(bson.D{
				{Key: "guild", Value: "guild-1"},
				{Key: "reply", Value: "legacy-invalid-reply"},
				{Key: "message", Value: test.message},
				{Key: "time", Value: int64(1_700_000_000_123)},
			})
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var document AutoChatPaidDocument
			if err := bson.Unmarshal(encoded, &document); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			response, ok := document.ToResponse()
			if !ok || response.Content != test.want || response.Reply {
				t.Fatalf("response = %#v ok=%v", response, ok)
			}
		})
	}
}

func TestAutoChatPaidDocumentRejectsMissingOrNullWorkerMessage(t *testing.T) {
	for _, message := range []bson.RawValue{{}, {Type: bson.TypeNull}} {
		document := AutoChatPaidDocument{
			Guild:   "guild-1",
			Message: message,
			Time:    rawBSONValue(t, int64(1_700_000_000_123)),
		}
		if response, ok := document.ToResponse(); ok {
			t.Fatalf("response = %#v, want rejected message", response)
		}
	}
}

func rawBSONValue(t *testing.T, value any) bson.RawValue {
	t.Helper()
	valueType, encoded, err := bson.MarshalValue(value)
	if err != nil {
		t.Fatalf("marshal %T: %v", value, err)
	}
	return bson.RawValue{Type: valueType, Value: encoded}
}
