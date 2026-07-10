package documents

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestLoggingConfigDocumentRoundTripPreservesLegacyFields(t *testing.T) {
	config := domain.LoggingConfig{
		GuildID:           "guild-1",
		ChannelID:         "channel-1",
		MessageUpdate:     true,
		MessageDelete:     true,
		ChannelUpdate:     false,
		MemberVoiceUpdate: true,
	}
	document := LoggingConfigDocumentFromDomain(config)
	if document.Guild != "guild-1" || document.ChannelID != "channel-1" || !document.MessageUpdate || !document.MessageDelete || document.ChannelUpdate || !document.MemberVoiceUpdate {
		t.Fatalf("document = %#v", document)
	}
	got := document.ToDomain()
	if got != config {
		t.Fatalf("round trip = %#v want %#v", got, config)
	}
}

func TestLoggingConfigReadDocumentUsesMongooseScalarCoercion(t *testing.T) {
	raw, err := bson.Marshal(bson.D{
		{Key: "guild", Value: "guild-1"},
		{Key: "channel_id", Value: int64(1234)},
		{Key: "message_update", Value: "yes"},
		{Key: "message_delete", Value: int32(1)},
		{Key: "channel_update", Value: "false"},
		{Key: "member_voice_update", Value: int32(2)},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var document LoggingConfigReadDocument
	if err := bson.Unmarshal(raw, &document); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	config := document.ToDomain()
	if config.GuildID != "guild-1" || config.ChannelID != "1234" || !config.MessageUpdate || !config.MessageDelete || config.ChannelUpdate || config.MemberVoiceUpdate {
		t.Fatalf("config = %#v", config)
	}
}
