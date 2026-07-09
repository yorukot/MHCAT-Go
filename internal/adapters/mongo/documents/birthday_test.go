package documents

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestBirthdayConfigDocumentRoundTripDomain(t *testing.T) {
	config := domain.BirthdayConfig{
		GuildID:                    "guild-1",
		Message:                    "{user} 生日快樂",
		UTCOffset:                  "+08:00",
		ChannelID:                  "channel-1",
		EveryoneCanSetBirthdayDate: true,
		RoleID:                     "role-1",
	}
	document := BirthdayConfigDocumentFromDomain(config)
	if document.Guild != "guild-1" || document.Message != "{user} 生日快樂" || document.UTCOffset != "+08:00" || document.Channel != "channel-1" || document.Role == nil || *document.Role != "role-1" {
		t.Fatalf("document = %#v", document)
	}
	if got := document.ToDomain(); got != config {
		t.Fatalf("round trip = %#v, want %#v", got, config)
	}
}

func TestBirthdayConfigDocumentNilRoleDecodesEmpty(t *testing.T) {
	payload, err := bson.Marshal(bson.D{
		{Key: "guild", Value: "guild-1"},
		{Key: "msg", Value: "happy"},
		{Key: "utc", Value: "+08:00"},
		{Key: "channel", Value: "channel-1"},
		{Key: "everyone_can_set_birthday_date", Value: false},
		{Key: "role", Value: nil},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var document BirthdayConfigDocument
	if err := bson.Unmarshal(payload, &document); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := document.ToDomain(); got.RoleID != "" || got.GuildID != "guild-1" {
		t.Fatalf("domain = %#v", got)
	}
}
