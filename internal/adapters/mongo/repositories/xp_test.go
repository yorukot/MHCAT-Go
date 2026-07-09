package repositories

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestXPChannelConfigUpdatePreservesMessageAndUnsetsBackground(t *testing.T) {
	update, err := xpChannelConfigUpdate("channel-1", " #00ff00 ", "  {user} 升到了 {level}  ", "guild-1", true)
	if err != nil {
		t.Fatalf("build update: %v", err)
	}
	set := documentValue(t, update, "$set")
	if value := documentValue(t, set, "channel"); value != "channel-1" {
		t.Fatalf("channel = %#v", value)
	}
	if value := documentValue(t, set, "color"); value != "#00ff00" {
		t.Fatalf("color = %#v", value)
	}
	if value := documentValue(t, set, "message"); value != "  {user} 升到了 {level}  " {
		t.Fatalf("message should preserve legacy spacing, got %#v", value)
	}
	unset := documentValue(t, update, "$unset")
	if value := documentValue(t, unset, "background"); value == nil {
		t.Fatalf("background unset missing in %#v", unset)
	}
	setOnInsert := documentValue(t, update, "$setOnInsert")
	if value := documentValue(t, setOnInsert, "guild"); value != "guild-1" {
		t.Fatalf("guild setOnInsert = %#v", value)
	}
}

func TestXPChannelConfigUpdateUsesNilForEmptyColorAndMessage(t *testing.T) {
	update, err := xpChannelConfigUpdate("channel-1", " ", "", "", false)
	if err != nil {
		t.Fatalf("build update: %v", err)
	}
	set := documentValue(t, update, "$set")
	if value := documentValue(t, set, "color"); value != nil {
		t.Fatalf("empty color should become nil, got %#v", value)
	}
	if value := documentValue(t, set, "message"); value != nil {
		t.Fatalf("empty message should become nil, got %#v", value)
	}
	if hasKey(update, "$setOnInsert") {
		t.Fatalf("non-upsert update should not include setOnInsert: %#v", update)
	}
}

func documentValue(t *testing.T, doc any, key string) any {
	t.Helper()
	switch typed := doc.(type) {
	case bson.D:
		for _, element := range typed {
			if element.Key == key {
				return element.Value
			}
		}
	case []bson.E:
		for _, element := range typed {
			if element.Key == key {
				return element.Value
			}
		}
	default:
		t.Fatalf("unsupported document type %T for key %s", doc, key)
	}
	t.Fatalf("missing key %s in %#v", key, doc)
	return nil
}

func hasKey(doc bson.D, key string) bool {
	for _, element := range doc {
		if element.Key == key {
			return true
		}
	}
	return false
}
