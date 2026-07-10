package documents

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestAnnouncementReadDocumentsUseMongooseStringCoercion(t *testing.T) {
	payload, err := bson.Marshal(bson.D{
		{Key: "guild", Value: "guild-1"},
		{Key: "announcement_id", Value: int64(123)},
		{Key: "tag", Value: true},
		{Key: "color", Value: int32(15548997)},
		{Key: "title", Value: false},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var guildDocument GuildAnnouncementReadDocument
	if err := bson.Unmarshal(payload, &guildDocument); err != nil {
		t.Fatalf("unmarshal guild: %v", err)
	}
	guildConfig := guildDocument.ToDomain()
	if guildConfig.GuildID != "guild-1" || guildConfig.ChannelID != "123" {
		t.Fatalf("guild config = %#v", guildConfig)
	}

	var boundDocument BoundAnnouncementReadDocument
	if err := bson.Unmarshal(payload, &boundDocument); err != nil {
		t.Fatalf("unmarshal bound: %v", err)
	}
	boundConfig := boundDocument.ToDomain()
	if boundConfig.GuildID != "guild-1" || boundConfig.ChannelID != "123" || boundConfig.Tag != "true" || boundConfig.Color != "15548997" || boundConfig.Title != "false" {
		t.Fatalf("bound config = %#v", boundConfig)
	}
}

func TestAnnouncementReadDocumentsRejectCompoundStringShapes(t *testing.T) {
	payload, err := bson.Marshal(bson.D{
		{Key: "guild", Value: bson.D{{Key: "bad", Value: true}}},
		{Key: "announcement_id", Value: bson.A{"bad"}},
		{Key: "tag", Value: bson.D{{Key: "bad", Value: true}}},
		{Key: "color", Value: bson.A{"bad"}},
		{Key: "title", Value: bson.D{{Key: "bad", Value: true}}},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var document BoundAnnouncementReadDocument
	if err := bson.Unmarshal(payload, &document); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if config := document.ToDomain(); config.GuildID != "" || config.ChannelID != "" || config.Tag != "" || config.Color != "" || config.Title != "" {
		t.Fatalf("compound values should remain unusable: %#v", config)
	}
}

func TestAnnouncementWriteDocumentsRemainTyped(t *testing.T) {
	guildPayload, err := bson.Marshal(GuildAnnouncementDocument{Guild: "guild-1", AnnouncementID: "channel-1"})
	if err != nil {
		t.Fatalf("marshal guild: %v", err)
	}
	guildRaw := bson.Raw(guildPayload)
	if guildRaw.Lookup("guild").Type != bson.TypeString || guildRaw.Lookup("announcement_id").Type != bson.TypeString {
		t.Fatalf("guild payload = %#v", guildRaw)
	}

	boundPayload, err := bson.Marshal(BoundAnnouncementDocument{Guild: "guild-1", AnnouncementID: "channel-1", Tag: "@here", Color: "Random", Title: "公告"})
	if err != nil {
		t.Fatalf("marshal bound: %v", err)
	}
	boundRaw := bson.Raw(boundPayload)
	for _, field := range []string{"guild", "announcement_id", "tag", "color", "title"} {
		if boundRaw.Lookup(field).Type != bson.TypeString {
			t.Fatalf("field %s type = %s", field, boundRaw.Lookup(field).Type)
		}
	}
}
