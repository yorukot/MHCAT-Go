package documents

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestGoodWebConfigDocumentRoundTrip(t *testing.T) {
	document := GoodWebConfigDocumentFromDomain(domain.AntiScamConfig{
		GuildID: "guild-1",
		Open:    true,
	})
	if document.Guild != "guild-1" || !document.Open {
		t.Fatalf("document = %#v", document)
	}
	config := document.ToDomain()
	if config.GuildID != "guild-1" || !config.Open {
		t.Fatalf("config = %#v", config)
	}
}

func TestGoodWebConfigReadDocumentUsesMongooseBooleanCoercion(t *testing.T) {
	for _, test := range []struct {
		value any
		want  bool
	}{
		{value: true, want: true},
		{value: int32(1), want: true},
		{value: "true", want: true},
		{value: "1", want: true},
		{value: "yes", want: true},
		{value: false},
		{value: int32(0)},
		{value: "false"},
		{value: "0"},
		{value: "no"},
		{value: "invalid"},
		{value: nil},
	} {
		raw, err := bson.Marshal(bson.D{{Key: "guild", Value: "guild-1"}, {Key: "open", Value: test.value}})
		if err != nil {
			t.Fatalf("marshal %#v: %v", test.value, err)
		}
		var document GoodWebConfigReadDocument
		if err := bson.Unmarshal(raw, &document); err != nil {
			t.Fatalf("decode %#v: %v", test.value, err)
		}
		config := document.ToDomain()
		if config.GuildID != "guild-1" || config.Open != test.want {
			t.Fatalf("value %#v decoded as %#v, want open %v", test.value, config, test.want)
		}
	}
}
