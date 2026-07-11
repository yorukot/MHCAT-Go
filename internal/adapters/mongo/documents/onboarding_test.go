package documents

import (
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestJoinRoleDocumentRoundTrip(t *testing.T) {
	doc := JoinRoleDocumentFromDomain(domain.JoinRoleConfig{
		GuildID: "guild",
		RoleID:  "role",
		GiveTo:  domain.JoinRoleGiveBots,
	})
	if doc.Guild != "guild" || doc.Role != "role" || doc.GiveToWho != domain.JoinRoleGiveBots {
		t.Fatalf("document = %#v", doc)
	}
	back := doc.ToDomain()
	if back.GuildID != "guild" || back.RoleID != "role" || back.GiveTo != domain.JoinRoleGiveBots {
		t.Fatalf("domain = %#v", back)
	}
}

func TestJoinRoleDocumentDefaultsGiveTo(t *testing.T) {
	back := JoinRoleDocument{Guild: "guild", Role: "role"}.ToDomain()
	if back.GiveTo != domain.JoinRoleGiveAllUsers {
		t.Fatalf("give to = %q", back.GiveTo)
	}
}

func TestJoinRoleDocumentPreservesUnknownStoredAudience(t *testing.T) {
	back := JoinRoleDocument{Guild: "guild", Role: "role", GiveToWho: " unknown "}.ToDomain()
	if back.GiveTo != "unknown" {
		t.Fatalf("give to = %q", back.GiveTo)
	}
}

func TestJoinMessageDocumentToDomain(t *testing.T) {
	enabled := true
	content := "welcome"
	color := "#53FF53"
	image := "https://example.test/welcome.png"
	back := JoinMessageDocument{
		Guild:          "guild",
		Enable:         &enabled,
		MessageContent: &content,
		Color:          &color,
		Channel:        "channel",
		Image:          &image,
	}.ToDomain()
	if back.GuildID != "guild" || !back.Enabled || back.ChannelID != "channel" || back.MessageContent != "welcome" || back.Color != "#53FF53" || back.ImageURL != image {
		t.Fatalf("domain = %#v", back)
	}
}

func TestJoinMessageDocumentMissingEnableDefaultsEnabled(t *testing.T) {
	back := JoinMessageDocument{Guild: "guild", Channel: "channel"}.ToDomain()
	if !back.Enabled {
		t.Fatalf("missing legacy enable should not disable delivery: %#v", back)
	}
}

func TestJoinMessageDocumentDecodesLegacyEnableShapes(t *testing.T) {
	for _, tc := range []struct {
		name        string
		enable      any
		include     bool
		wantEnabled bool
	}{
		{name: "missing", wantEnabled: true},
		{name: "null", enable: nil, include: true, wantEnabled: true},
		{name: "false", enable: false, include: true},
		{name: "true", enable: true, include: true, wantEnabled: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := bson.D{
				{Key: "guild", Value: "guild"},
				{Key: "channel", Value: "channel"},
				{Key: "message_content", Value: "   "},
				{Key: "color", Value: "Green"},
			}
			if tc.include {
				raw = append(raw, bson.E{Key: "enable", Value: tc.enable})
			}
			encoded, err := bson.Marshal(raw)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var document JoinMessageDocument
			if err := bson.Unmarshal(encoded, &document); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			config := document.ToDomain()
			if config.Enabled != tc.wantEnabled || config.MessageContent != "   " || config.Color != "Green" {
				t.Fatalf("config = %#v", config)
			}
		})
	}
}

func TestLeaveMessageDocumentRoundTrip(t *testing.T) {
	doc := LeaveMessageDocumentFromDomain(domain.LeaveMessageConfig{
		GuildID:        "guild",
		ChannelID:      "channel",
		MessageContent: "bye",
		Title:          "bye title",
		Color:          "#df1f2f",
	})
	if doc.Guild != "guild" || doc.Channel != "channel" || doc.MessageContent == nil || *doc.MessageContent != "bye" {
		t.Fatalf("document = %#v", doc)
	}
	back := doc.ToDomain()
	if back.GuildID != "guild" || back.ChannelID != "channel" || back.MessageContent != "bye" || back.Title != "bye title" || back.Color != "#df1f2f" {
		t.Fatalf("domain = %#v", back)
	}
}

func TestLeaveMessageDocumentNullFieldsDecodeEmpty(t *testing.T) {
	back := LeaveMessageDocument{Guild: "guild", Channel: "channel"}.ToDomain()
	if back.MessageContent != "" || back.Title != "" || back.Color != "" {
		t.Fatalf("domain = %#v", back)
	}
}

func TestLeaveMessageDocumentDecodesNullableLegacyFields(t *testing.T) {
	encoded, err := bson.Marshal(bson.D{
		{Key: "guild", Value: "guild"},
		{Key: "channel", Value: "channel"},
		{Key: "message_content", Value: "   "},
		{Key: "title", Value: nil},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var document LeaveMessageDocument
	if err := bson.Unmarshal(encoded, &document); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	config := document.ToDomain()
	if config.MessageContent != "   " || config.Title != "" || config.Color != "" || config.ChannelID != "channel" {
		t.Fatalf("config = %#v", config)
	}
}

func TestVerificationDocumentDecodesLegacyRenameShapes(t *testing.T) {
	tests := []struct {
		name       string
		document   bson.D
		wantRename string
	}{
		{
			name:       "missing name",
			document:   bson.D{{Key: "guild", Value: "guild"}, {Key: "role", Value: "role"}},
			wantRename: "",
		},
		{
			name:       "null name",
			document:   bson.D{{Key: "guild", Value: "guild"}, {Key: "role", Value: "role"}, {Key: "name", Value: nil}},
			wantRename: "",
		},
		{
			name:       "string name",
			document:   bson.D{{Key: "guild", Value: "guild"}, {Key: "role", Value: "role"}, {Key: "name", Value: "  {name} | MHCAT  "}},
			wantRename: "  {name} | MHCAT  ",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := bson.Marshal(tc.document)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var document VerificationDocument
			if err := bson.Unmarshal(encoded, &document); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			config := document.ToDomain()
			if config.GuildID != "guild" || config.RoleID != "role" || config.RenameTemplate != tc.wantRename {
				t.Fatalf("config = %#v", config)
			}
		})
	}
}

func TestAccountAgeDocumentDecodesLegacyShapes(t *testing.T) {
	tests := []struct {
		name        string
		document    bson.D
		wantChannel string
	}{
		{
			name:     "missing channel",
			document: bson.D{{Key: "guild", Value: "guild"}, {Key: "hours", Value: "3600"}},
		},
		{
			name:     "null channel",
			document: bson.D{{Key: "guild", Value: "guild"}, {Key: "hours", Value: "3600"}, {Key: "channel", Value: nil}},
		},
		{
			name:        "string channel",
			document:    bson.D{{Key: "guild", Value: "guild"}, {Key: "hours", Value: "3600"}, {Key: "channel", Value: "channel"}},
			wantChannel: "channel",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := bson.Marshal(tc.document)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var document AccountAgeDocument
			if err := bson.Unmarshal(encoded, &document); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			config, err := document.ToDomain()
			if err != nil {
				t.Fatalf("to domain: %v", err)
			}
			if config.GuildID != "guild" || config.RequiredSeconds != 3600 || config.ChannelID != tc.wantChannel {
				t.Fatalf("config = %#v", config)
			}
		})
	}
}

func TestAccountAgeDocumentRejectsInvalidLegacyHours(t *testing.T) {
	for _, tc := range []struct {
		name  string
		hours any
	}{
		{name: "null", hours: nil},
		{name: "empty", hours: ""},
		{name: "non-numeric", hours: "not-a-number"},
		{name: "zero", hours: "0"},
		{name: "negative", hours: "-3600"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := bson.Marshal(bson.D{{Key: "guild", Value: "guild"}, {Key: "hours", Value: tc.hours}})
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var document AccountAgeDocument
			if err := bson.Unmarshal(encoded, &document); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if _, err := document.ToDomain(); !errors.Is(err, domain.ErrInvalidAccountAgeConfig) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestAccountAgeReadDocumentUsesMongooseStringAndNumberCoercion(t *testing.T) {
	tests := []struct {
		name        string
		hours       any
		channel     any
		wantSeconds float64
		wantChannel string
	}{
		{name: "numeric scalar", hours: int32(3600), channel: int64(123), wantSeconds: 3600, wantChannel: "123"},
		{name: "exponent string", hours: "3.6e3", channel: true, wantSeconds: 3600, wantChannel: "true"},
		{name: "fractional string", hours: "3600.5", channel: nil, wantSeconds: 3600.5},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := bson.Marshal(bson.D{
				{Key: "guild", Value: "guild"},
				{Key: "hours", Value: tc.hours},
				{Key: "channel", Value: tc.channel},
			})
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var document AccountAgeReadDocument
			if err := bson.Unmarshal(encoded, &document); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			config, err := document.ToDomain()
			if err != nil {
				t.Fatalf("to domain: %v", err)
			}
			if config.GuildID != "guild" || config.RequiredSeconds != tc.wantSeconds || config.ChannelID != tc.wantChannel {
				t.Fatalf("config = %#v", config)
			}
		})
	}
}

func TestAccountAgeReadDocumentRejectsUnusableHoursWithoutRejectingChannel(t *testing.T) {
	for _, hours := range []any{nil, "", "not-a-number", "0", "-1", "Infinity", bson.D{{Key: "bad", Value: true}}, bson.A{3600}} {
		encoded, err := bson.Marshal(bson.D{
			{Key: "guild", Value: "guild"},
			{Key: "hours", Value: hours},
			{Key: "channel", Value: bson.D{{Key: "bad", Value: true}}},
		})
		if err != nil {
			t.Fatalf("marshal %#v: %v", hours, err)
		}
		var document AccountAgeReadDocument
		if err := bson.Unmarshal(encoded, &document); err != nil {
			t.Fatalf("unmarshal %#v: %v", hours, err)
		}
		if document.ChannelID() != "" {
			t.Fatalf("compound channel should remain unusable: %#v", document.Channel)
		}
		if _, err := document.ToDomain(); !errors.Is(err, domain.ErrInvalidAccountAgeConfig) {
			t.Fatalf("hours %#v error = %v", hours, err)
		}
	}
}

func TestAccountAgeWriteDocumentRemainsTyped(t *testing.T) {
	payload, err := bson.Marshal(AccountAgeDocumentFromDomain(domain.AccountAgeConfig{
		GuildID: "guild", RequiredSeconds: 3600.5, ChannelID: "channel",
	}))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	raw := bson.Raw(payload)
	if raw.Lookup("guild").Type != bson.TypeString || raw.Lookup("hours").Type != bson.TypeString || raw.Lookup("hours").StringValue() != "3600.5" || raw.Lookup("channel").Type != bson.TypeString {
		t.Fatalf("payload = %#v", raw)
	}

	nullPayload, err := bson.Marshal(AccountAgeDocumentFromDomain(domain.AccountAgeConfig{GuildID: "guild", RequiredSeconds: 3600}))
	if err != nil {
		t.Fatalf("marshal null channel: %v", err)
	}
	if got := bson.Raw(nullPayload).Lookup("channel").Type; got != 0 {
		t.Fatalf("omitted channel type = %s", got)
	}
}
