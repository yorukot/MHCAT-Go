package documents

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestTextXPChannelDocumentRoundTrip(t *testing.T) {
	document := TextXPChannelDocumentFromDomain(domain.TextXPConfig{
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		Color:     "#00ff00",
		Message:   "{user} 升到 {level}",
	})
	if document.Guild != "guild-1" || document.Channel != "channel-1" || document.Color != "#00ff00" || document.Message == "" {
		t.Fatalf("document = %#v", document)
	}
	config := document.ToDomain()
	if config.GuildID != "guild-1" || config.ChannelID != "channel-1" || config.Color != "#00ff00" || config.Message == "" {
		t.Fatalf("config = %#v", config)
	}
}

func TestVoiceXPChannelDocumentRoundTrip(t *testing.T) {
	document := VoiceXPChannelDocumentFromDomain(domain.VoiceXPConfig{
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		Color:     "#00ff00",
		Message:   "{user} 升到 {level}",
	})
	if document.Guild != "guild-1" || document.Channel != "channel-1" || document.Color != "#00ff00" || document.Message == "" {
		t.Fatalf("document = %#v", document)
	}
	config := document.ToDomain()
	if config.GuildID != "guild-1" || config.ChannelID != "channel-1" || config.Color != "#00ff00" || config.Message == "" {
		t.Fatalf("config = %#v", config)
	}
}

func TestXPRewardRoleDocumentWritesLegacyStringLevel(t *testing.T) {
	document := XPRewardRoleDocumentFromDomain(domain.XPRewardRoleConfig{
		GuildID:       " guild-1 ",
		Level:         12,
		RoleID:        " role-1 ",
		DeleteWhenNot: true,
	})
	encoded, err := bson.Marshal(document)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	level, ok := bson.Raw(encoded).Lookup("leavel").StringValueOK()
	if !ok || level != "12" {
		t.Fatalf("leavel = %q/%t", level, ok)
	}
	if got := document.ToDomain(); got.GuildID != "guild-1" || got.Level != 12 || got.RoleID != "role-1" || !got.DeleteWhenNot {
		t.Fatalf("round trip = %#v", got)
	}
}

func TestXPRewardRoleDocumentReadsLegacyNumericLevels(t *testing.T) {
	for _, value := range []any{"12", int32(12), int64(12), float64(12)} {
		t.Run(bsonTypeName(value), func(t *testing.T) {
			encoded, err := bson.Marshal(bson.D{
				{Key: "guild", Value: "guild-1"},
				{Key: "leavel", Value: value},
				{Key: "role", Value: "role-1"},
				{Key: "delete_when_not", Value: true},
			})
			if err != nil {
				t.Fatalf("marshal fixture: %v", err)
			}
			var document XPRewardRoleDocument
			if err := bson.Unmarshal(encoded, &document); err != nil {
				t.Fatalf("unmarshal fixture: %v", err)
			}
			if got := document.ToDomain(); got.GuildID != "guild-1" || got.Level != 12 || got.RoleID != "role-1" || !got.DeleteWhenNot {
				t.Fatalf("decoded config = %#v", got)
			}
		})
	}
}

func TestXPProfileDocumentReadsLegacyMixedNumericFields(t *testing.T) {
	for _, value := range []any{"12", int32(12), int64(12), float64(12)} {
		t.Run(bsonTypeName(value), func(t *testing.T) {
			encoded, err := bson.Marshal(bson.D{
				{Key: "guild", Value: "guild-1"},
				{Key: "member", Value: "user-1"},
				{Key: "xp", Value: value},
				{Key: "leavel", Value: value},
				{Key: "leavejoin", Value: "leave"},
			})
			if err != nil {
				t.Fatalf("marshal fixture: %v", err)
			}
			var document XPProfileDocument
			if err := bson.Unmarshal(encoded, &document); err != nil {
				t.Fatalf("unmarshal fixture: %v", err)
			}
			if got := document.ToDomain(); got.GuildID != "guild-1" || got.UserID != "user-1" || got.XP != 12 || got.Level != 12 || got.LeaveJoin != "leave" {
				t.Fatalf("decoded profile = %#v", got)
			}
		})
	}
}

func bsonTypeName(value any) string {
	switch value.(type) {
	case string:
		return "string"
	case int32:
		return "int32"
	case int64:
		return "int64"
	case float64:
		return "double"
	default:
		return "unknown"
	}
}
