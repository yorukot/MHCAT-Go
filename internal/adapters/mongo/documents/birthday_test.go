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

func TestBirthdayProfileDocumentRoundTripDomain(t *testing.T) {
	year, month, day, hour, minute := 1999, 2, 28, 8, 5
	profile := domain.BirthdayProfile{
		GuildID:       "guild-1",
		UserID:        "user-1",
		BirthdayYear:  &year,
		BirthdayMonth: &month,
		BirthdayDay:   &day,
		SendHour:      &hour,
		SendMinute:    &minute,
		AllowAdmin:    true,
	}

	document := BirthdayProfileDocumentFromDomain(profile)
	if document.Guild != "guild-1" || document.User != "user-1" || document.BirthdayYear == nil || *document.BirthdayYear != 1999 || !document.Allow {
		t.Fatalf("document = %#v", document)
	}
	roundTrip := document.ToDomain()
	if roundTrip.GuildID != "guild-1" || roundTrip.UserID != "user-1" || roundTrip.BirthdayMonth == nil || *roundTrip.BirthdayMonth != 2 || !roundTrip.AllowAdmin {
		t.Fatalf("round trip = %#v", roundTrip)
	}
}

func TestBirthdayProfileDocumentPreservesNullDateFields(t *testing.T) {
	document := BirthdayProfileDocumentFromDomain(domain.BirthdayProfile{GuildID: "guild-1", UserID: "user-1", AllowAdmin: false})
	if document.BirthdayYear != nil || document.BirthdayMonth != nil || document.BirthdayDay != nil || document.SendHour != nil || document.SendMinute != nil {
		t.Fatalf("document = %#v", document)
	}
}

func TestBirthdayReadDocumentsUseMongooseScalarCoercion(t *testing.T) {
	configPayload, err := bson.Marshal(bson.D{
		{Key: "guild", Value: int64(123)},
		{Key: "msg", Value: true},
		{Key: "utc", Value: int32(8)},
		{Key: "channel", Value: false},
		{Key: "everyone_can_set_birthday_date", Value: int32(1)},
		{Key: "role", Value: nil},
	})
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	var config BirthdayConfigReadDocument
	if err := bson.Unmarshal(configPayload, &config); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	configDomain := config.ToDomain()
	if configDomain.GuildID != "123" || configDomain.Message != "true" || configDomain.UTCOffset != "8" || configDomain.ChannelID != "false" || !configDomain.EveryoneCanSetBirthdayDate || configDomain.RoleID != "" {
		t.Fatalf("config = %#v", configDomain)
	}

	profilePayload, err := bson.Marshal(bson.D{
		{Key: "guild", Value: int64(123)},
		{Key: "user", Value: true},
		{Key: "birthday_year", Value: "2000"},
		{Key: "birthday_month", Value: "0x2"},
		{Key: "birthday_day", Value: int64(29)},
		{Key: "send_msg_hour", Value: true},
		{Key: "send_msg_min", Value: nil},
		{Key: "allow", Value: float64(1)},
	})
	if err != nil {
		t.Fatalf("marshal profile: %v", err)
	}
	var profile BirthdayProfileReadDocument
	if err := bson.Unmarshal(profilePayload, &profile); err != nil {
		t.Fatalf("unmarshal profile: %v", err)
	}
	profileDomain := profile.ToDomain()
	if profileDomain.GuildID != "123" || profileDomain.UserID != "true" ||
		profileDomain.BirthdayYear == nil || *profileDomain.BirthdayYear != 2000 ||
		profileDomain.BirthdayMonth == nil || *profileDomain.BirthdayMonth != 2 ||
		profileDomain.BirthdayDay == nil || *profileDomain.BirthdayDay != 29 ||
		profileDomain.SendHour == nil || *profileDomain.SendHour != 1 ||
		profileDomain.SendMinute != nil || !profileDomain.AllowAdmin {
		t.Fatalf("profile = %#v", profileDomain)
	}
}

func TestBirthdayReadDocumentsRejectCompoundAndNonIntegralValues(t *testing.T) {
	payload, err := bson.Marshal(bson.D{
		{Key: "guild", Value: bson.D{{Key: "bad", Value: true}}},
		{Key: "user", Value: bson.A{"bad"}},
		{Key: "birthday_year", Value: float64(2000.5)},
		{Key: "birthday_month", Value: bson.D{{Key: "bad", Value: true}}},
		{Key: "birthday_day", Value: bson.A{29}},
		{Key: "send_msg_hour", Value: "not-a-number"},
		{Key: "send_msg_min", Value: nil},
		{Key: "allow", Value: bson.A{true}},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var document BirthdayProfileReadDocument
	if err := bson.Unmarshal(payload, &document); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	profile := document.ToDomain()
	if profile.GuildID != "" || profile.UserID != "" || profile.BirthdayYear != nil || profile.BirthdayMonth != nil || profile.BirthdayDay != nil || profile.SendHour != nil || profile.SendMinute != nil || profile.AllowAdmin {
		t.Fatalf("compound values should remain unusable: %#v", profile)
	}
}

func TestBirthdayWriteDocumentsRemainTyped(t *testing.T) {
	configPayload, err := bson.Marshal(BirthdayConfigDocumentFromDomain(domain.BirthdayConfig{
		GuildID: "guild-1", Message: "happy", UTCOffset: "+08:00", ChannelID: "channel-1", EveryoneCanSetBirthdayDate: true,
	}))
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	configRaw := bson.Raw(configPayload)
	for _, field := range []string{"guild", "msg", "utc", "channel"} {
		if configRaw.Lookup(field).Type != bson.TypeString {
			t.Fatalf("config field %s type = %s", field, configRaw.Lookup(field).Type)
		}
	}
	if configRaw.Lookup("everyone_can_set_birthday_date").Type != bson.TypeBoolean {
		t.Fatalf("config boolean type = %s", configRaw.Lookup("everyone_can_set_birthday_date").Type)
	}

	year := 2000
	profilePayload, err := bson.Marshal(BirthdayProfileDocumentFromDomain(domain.BirthdayProfile{GuildID: "guild-1", UserID: "user-1", BirthdayYear: &year, AllowAdmin: true}))
	if err != nil {
		t.Fatalf("marshal profile: %v", err)
	}
	profileRaw := bson.Raw(profilePayload)
	if profileRaw.Lookup("guild").Type != bson.TypeString || profileRaw.Lookup("user").Type != bson.TypeString || profileRaw.Lookup("birthday_year").Type != bson.TypeInt32 || profileRaw.Lookup("allow").Type != bson.TypeBoolean {
		t.Fatalf("profile payload = %#v", profileRaw)
	}
}
