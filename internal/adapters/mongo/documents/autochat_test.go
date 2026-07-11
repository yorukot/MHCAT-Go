package documents

import (
	"math"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestAutoChatConfigDocumentRoundTrip(t *testing.T) {
	document := AutoChatConfigDocumentFromDomain(domain.AutoChatConfig{
		GuildID:   "guild-1",
		ChannelID: "channel-1",
	})
	if document.Guild != "guild-1" || document.Channel != "channel-1" {
		t.Fatalf("document = %#v", document)
	}
	config := document.ToDomain()
	if config.GuildID != "guild-1" || config.ChannelID != "channel-1" {
		t.Fatalf("config = %#v", config)
	}
}

func TestAutoChatConfigReadDocumentUsesMongooseStringCoercion(t *testing.T) {
	for _, test := range []struct {
		name  string
		value any
		want  string
	}{
		{name: "string whitespace", value: " channel-1 ", want: " channel-1 "},
		{name: "int64 channel", value: int64(123456789012345678), want: "123456789012345678"},
		{name: "boolean", value: true, want: "true"},
		{name: "nan", value: math.NaN(), want: "NaN"},
		{name: "object id", value: bson.ObjectID{1, 2, 3}, want: "010203000000000000000000"},
		{name: "binary", value: bson.Binary{Data: []byte("channel-1")}, want: "channel-1"},
		{name: "null", value: nil},
		{name: "compound", value: bson.D{{Key: "bad", Value: true}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			payload, err := bson.Marshal(bson.D{{Key: "guild", Value: "guild-1"}, {Key: "channel", Value: test.value}})
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var document AutoChatConfigReadDocument
			if err := bson.Unmarshal(payload, &document); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			config := document.ToDomain()
			if config.GuildID != "guild-1" || config.ChannelID != test.want {
				t.Fatalf("config=%#v want channel=%q", config, test.want)
			}
		})
	}
}
