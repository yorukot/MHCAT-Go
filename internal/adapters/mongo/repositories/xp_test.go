package repositories

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
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

func TestXPRewardRoleFilterUsesLegacyMisspelledLevelField(t *testing.T) {
	filter := xpRewardRoleFilter(" guild-1 ", " 12 ", " role-1 ")
	if value := documentValue(t, filter, "guild"); value != "guild-1" {
		t.Fatalf("guild = %#v", value)
	}
	if value := documentValue(t, filter, "leavel"); value != "12" {
		t.Fatalf("leavel = %#v", value)
	}
	if value := documentValue(t, filter, "role"); value != "role-1" {
		t.Fatalf("role = %#v", value)
	}
}

func TestXPProfileUpdateStoresLegacyStringFields(t *testing.T) {
	update := xpProfileUpdate(domain.XPProfile{GuildID: " guild-1 ", UserID: " user-1 ", XP: 87, Level: 2}, false)
	set := documentValue(t, update, "$set")
	if value := documentValue(t, set, "xp"); value != "87" {
		t.Fatalf("xp = %#v", value)
	}
	if value := documentValue(t, set, "leavel"); value != "2" {
		t.Fatalf("leavel = %#v", value)
	}
	setOnInsert := documentValue(t, update, "$setOnInsert")
	if value := documentValue(t, setOnInsert, "guild"); value != "guild-1" {
		t.Fatalf("guild = %#v", value)
	}
	if value := documentValue(t, setOnInsert, "member"); value != "user-1" {
		t.Fatalf("member = %#v", value)
	}
	if hasKey(setOnInsert.(bson.D), "leavejoin") {
		t.Fatalf("text profile insert should not include leavejoin: %#v", setOnInsert)
	}
}

func TestVoiceXPProfileUpdateSetsLegacyLeaveJoinOnInsert(t *testing.T) {
	update := xpProfileUpdate(domain.XPProfile{GuildID: "guild-1", UserID: "user-1", XP: 12, Level: 1}, true)
	setOnInsert := documentValue(t, update, "$setOnInsert")
	if value := documentValue(t, setOnInsert, "leavejoin"); value != "leave" {
		t.Fatalf("leavejoin = %#v", value)
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
