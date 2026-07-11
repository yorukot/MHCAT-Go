package documents

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestStatsConfigReadDocumentUsesMongooseStringCoercion(t *testing.T) {
	document := StatsConfigDocument{
		Guild:             statsRawValue(t, int64(123)),
		Parent:            statsRawValue(t, true),
		MemberNumber:      statsRawValue(t, bson.NewObjectID()),
		MemberNumberName:  statsRawValue(t, 12.5),
		UserNumber:        statsRawValue(t, bson.Binary{Data: []byte("user-channel")}),
		UserNumberName:    statsRawValue(t, false),
		BotNumber:         statsRawValue(t, bson.Regex{Pattern: "bot", Options: "i"}),
		BotNumberName:     statsRawValue(t, int32(2)),
		ChannelNumber:     statsRawValue(t, bson.D{{Key: "invalid", Value: true}}),
		ChannelNumberName: bson.RawValue{Type: bson.TypeNull},
	}

	got := document.ToDomain()
	if got.GuildID != "123" || got.ParentID != "true" || got.MemberNumberID == "" || got.MemberNumberName != "12.5" {
		t.Fatalf("base scalar coercion = %#v", got)
	}
	if got.UserNumberID != "user-channel" || got.UserNumberName != "false" || got.BotNumberID != "/bot/i" || got.BotNumberName != "2" {
		t.Fatalf("counter scalar coercion = %#v", got)
	}
	if got.ChannelNumberID != "" || got.ChannelNumberName != "" || got.TextNumberID != "" {
		t.Fatalf("invalid/null/missing values = %#v", got)
	}
}

func TestStatsRoleConfigReadDocumentUsesMongooseStringCoercion(t *testing.T) {
	document := StatsRoleConfigDocument{
		Guild:       statsRawValue(t, int32(42)),
		Channel:     statsRawValue(t, bson.Binary{Data: []byte("channel")}),
		ChannelName: statsRawValue(t, true),
		Role:        statsRawValue(t, bson.A{"invalid"}),
	}

	got := document.ToDomain()
	if got.GuildID != "42" || got.ChannelID != "channel" || got.ChannelName != "true" || got.RoleID != "" {
		t.Fatalf("role scalar coercion = %#v", got)
	}
}

func TestStatsRoleConfigWriteDocumentRemainsTyped(t *testing.T) {
	document := StatsRoleConfigDocumentFromDomain(domain.StatsRoleConfig{
		GuildID: " guild ", ChannelID: " channel ", ChannelName: " 4 ", RoleID: " role ",
	})
	if document.Guild != "guild" || document.Channel != "channel" || document.ChannelName != "4" || document.Role != "role" {
		t.Fatalf("write document = %#v", document)
	}
}

func statsRawValue(t *testing.T, value any) bson.RawValue {
	t.Helper()
	valueType, encoded, err := bson.MarshalValue(value)
	if err != nil {
		t.Fatalf("marshal raw value %#v: %v", value, err)
	}
	return bson.RawValue{Type: valueType, Value: encoded}
}
