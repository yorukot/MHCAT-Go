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

func TestJoinRoleReadDocumentUsesMongooseStringCoercion(t *testing.T) {
	tests := []struct {
		name       string
		document   bson.D
		wantGuild  string
		wantRole   string
		wantGiveTo string
	}{
		{
			name:       "typed strings",
			document:   bson.D{{Key: "guild", Value: "guild"}, {Key: "role", Value: "role"}, {Key: "give_to_who", Value: domain.JoinRoleGiveBots}},
			wantGuild:  "guild",
			wantRole:   "role",
			wantGiveTo: domain.JoinRoleGiveBots,
		},
		{
			name:       "scalar fields",
			document:   bson.D{{Key: "guild", Value: int64(123)}, {Key: "role", Value: int32(456)}, {Key: "give_to_who", Value: true}},
			wantGuild:  "123",
			wantRole:   "456",
			wantGiveTo: "true",
		},
		{
			name:       "missing audience defaults all users",
			document:   bson.D{{Key: "guild", Value: "guild"}, {Key: "role", Value: "role"}},
			wantGuild:  "guild",
			wantRole:   "role",
			wantGiveTo: domain.JoinRoleGiveAllUsers,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := bson.Marshal(tc.document)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var document JoinRoleReadDocument
			if err := bson.Unmarshal(payload, &document); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			config := document.ToDomain()
			if config.GuildID != tc.wantGuild || config.RoleID != tc.wantRole || config.GiveTo != tc.wantGiveTo {
				t.Fatalf("config = %#v", config)
			}
		})
	}
}

func TestJoinRoleReadDocumentLeavesCompoundRequiredValuesInvalid(t *testing.T) {
	payload, err := bson.Marshal(bson.D{
		{Key: "guild", Value: "guild"},
		{Key: "role", Value: bson.D{{Key: "bad", Value: true}}},
		{Key: "give_to_who", Value: bson.A{"bad"}},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var document JoinRoleReadDocument
	if err := bson.Unmarshal(payload, &document); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	config := document.ToDomain()
	if config.RoleID != "" || config.GiveTo != "invalid" {
		t.Fatalf("config = %#v", config)
	}
	if !errors.Is(config.Validate(), domain.ErrInvalidJoinRoleConfig) {
		t.Fatalf("compound required role should remain invalid: %#v", config)
	}
}

func TestJoinRoleWriteDocumentRemainsTyped(t *testing.T) {
	payload, err := bson.Marshal(JoinRoleDocumentFromDomain(domain.JoinRoleConfig{
		GuildID: "guild", RoleID: "role", GiveTo: domain.JoinRoleGiveMembers,
	}))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	raw := bson.Raw(payload)
	for _, field := range []string{"guild", "role", "give_to_who"} {
		if raw.Lookup(field).Type != bson.TypeString {
			t.Fatalf("field %s type = %s", field, raw.Lookup(field).Type)
		}
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

func TestJoinMessageReadDocumentUsesMongooseScalarCoercion(t *testing.T) {
	objectID := bson.NewObjectID()
	for _, tc := range []struct {
		name        string
		document    bson.D
		wantEnabled bool
		wantGuild   string
		wantChannel string
		wantContent string
		wantColor   string
		wantImage   string
	}{
		{
			name:        "typed fields and missing enable",
			document:    bson.D{{Key: "guild", Value: "guild"}, {Key: "channel", Value: "channel"}, {Key: "message_content", Value: "   "}, {Key: "color", Value: "Green"}, {Key: "img", Value: "image"}},
			wantEnabled: true, wantGuild: "guild", wantChannel: "channel", wantContent: "   ", wantColor: "Green", wantImage: "image",
		},
		{
			name:      "legacy scalar fields",
			document:  bson.D{{Key: "guild", Value: int64(123)}, {Key: "enable", Value: "false"}, {Key: "channel", Value: objectID}, {Key: "message_content", Value: true}, {Key: "color", Value: int32(456)}, {Key: "img", Value: 1.5}},
			wantGuild: "123", wantChannel: objectID.Hex(), wantContent: "true", wantColor: "456", wantImage: "1.5",
		},
		{
			name:        "null optional fields",
			document:    bson.D{{Key: "guild", Value: "guild"}, {Key: "enable", Value: nil}, {Key: "channel", Value: "channel"}, {Key: "message_content", Value: nil}, {Key: "color", Value: nil}, {Key: "img", Value: nil}},
			wantEnabled: true, wantGuild: "guild", wantChannel: "channel",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := bson.Marshal(tc.document)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var document JoinMessageReadDocument
			if err := bson.Unmarshal(encoded, &document); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			config := document.ToDomain()
			if config.Enabled != tc.wantEnabled || config.GuildID != tc.wantGuild || config.ChannelID != tc.wantChannel || config.MessageContent != tc.wantContent || config.Color != tc.wantColor || config.ImageURL != tc.wantImage {
				t.Fatalf("config = %#v", config)
			}
		})
	}
}

func TestJoinMessageReadDocumentPreservesLegacyEnableSemantics(t *testing.T) {
	for _, tc := range []struct {
		name   string
		enable any
		want   bool
	}{
		{name: "false boolean", enable: false},
		{name: "zero integer", enable: int32(0)},
		{name: "zero double", enable: float64(0)},
		{name: "false string", enable: "false"},
		{name: "zero string", enable: "0"},
		{name: "no string", enable: "no"},
		{name: "true boolean", enable: true, want: true},
		{name: "one integer", enable: int64(1), want: true},
		{name: "yes string", enable: "yes", want: true},
		{name: "unknown scalar remains enabled", enable: "unknown", want: true},
		{name: "compound remains enabled", enable: bson.D{{Key: "bad", Value: true}}, want: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := bson.Marshal(bson.D{{Key: "enable", Value: tc.enable}})
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var document JoinMessageReadDocument
			if err := bson.Unmarshal(encoded, &document); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if got := document.ToDomain().Enabled; got != tc.want {
				t.Fatalf("enabled = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestJoinMessageReadDocumentLeavesCompoundStringsUnusable(t *testing.T) {
	encoded, err := bson.Marshal(bson.D{
		{Key: "guild", Value: bson.D{{Key: "bad", Value: true}}},
		{Key: "channel", Value: bson.A{"bad"}},
		{Key: "message_content", Value: bson.D{{Key: "bad", Value: true}}},
		{Key: "color", Value: bson.A{"bad"}},
		{Key: "img", Value: bson.D{{Key: "bad", Value: true}}},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var document JoinMessageReadDocument
	if err := bson.Unmarshal(encoded, &document); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	config := document.ToDomain()
	if config.GuildID != "" || config.ChannelID != "" || config.MessageContent != "" || config.Color != "" || config.ImageURL != "" || !config.Enabled {
		t.Fatalf("config = %#v", config)
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

func TestLeaveMessageReadDocumentUsesMongooseScalarCoercion(t *testing.T) {
	objectID := bson.NewObjectID()
	encoded, err := bson.Marshal(bson.D{
		{Key: "guild", Value: int64(123)},
		{Key: "channel", Value: objectID},
		{Key: "message_content", Value: "   "},
		{Key: "title", Value: true},
		{Key: "color", Value: int32(456)},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var document LeaveMessageReadDocument
	if err := bson.Unmarshal(encoded, &document); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	config := document.ToDomain()
	if config.GuildID != "123" || config.ChannelID != objectID.Hex() || config.MessageContent != "   " || config.Title != "true" || config.Color != "456" {
		t.Fatalf("config = %#v", config)
	}
}

func TestLeaveMessageReadDocumentTreatsMissingNullAndCompoundFieldsAsEmpty(t *testing.T) {
	encoded, err := bson.Marshal(bson.D{
		{Key: "guild", Value: "guild"},
		{Key: "channel", Value: bson.A{"bad"}},
		{Key: "message_content", Value: bson.D{{Key: "bad", Value: true}}},
		{Key: "title", Value: nil},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var document LeaveMessageReadDocument
	if err := bson.Unmarshal(encoded, &document); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	config := document.ToDomain()
	if config.GuildID != "guild" || config.ChannelID != "" || config.MessageContent != "" || config.Title != "" || config.Color != "" {
		t.Fatalf("config = %#v", config)
	}
}

func TestWelcomeMessageWriteDocumentsRemainTyped(t *testing.T) {
	enabled := false
	content := "  message  "
	color := "Random"
	image := "https://example.test/image.png"
	joinPayload, err := bson.Marshal(JoinMessageDocument{
		Guild: "guild", Enable: &enabled, MessageContent: &content, Color: &color, Channel: "channel", Image: &image,
	})
	if err != nil {
		t.Fatalf("marshal join: %v", err)
	}
	joinRaw := bson.Raw(joinPayload)
	for _, field := range []string{"guild", "message_content", "color", "channel", "img"} {
		if joinRaw.Lookup(field).Type != bson.TypeString {
			t.Fatalf("join field %s type = %s", field, joinRaw.Lookup(field).Type)
		}
	}
	if joinRaw.Lookup("enable").Type != bson.TypeBoolean {
		t.Fatalf("join enable type = %s", joinRaw.Lookup("enable").Type)
	}

	leavePayload, err := bson.Marshal(LeaveMessageDocumentFromDomain(domain.LeaveMessageConfig{
		GuildID: "guild", ChannelID: "channel", MessageContent: content, Title: "title", Color: color,
	}))
	if err != nil {
		t.Fatalf("marshal leave: %v", err)
	}
	leaveRaw := bson.Raw(leavePayload)
	for _, field := range []string{"guild", "message_content", "title", "color", "channel"} {
		if leaveRaw.Lookup(field).Type != bson.TypeString {
			t.Fatalf("leave field %s type = %s", field, leaveRaw.Lookup(field).Type)
		}
	}
}

func TestVerificationReadDocumentUsesMongooseStringCoercion(t *testing.T) {
	tests := []struct {
		name       string
		document   bson.D
		wantGuild  string
		wantRole   string
		wantRename string
	}{
		{
			name:       "missing name",
			document:   bson.D{{Key: "guild", Value: "guild"}, {Key: "role", Value: "role"}},
			wantGuild:  "guild",
			wantRole:   "role",
			wantRename: "",
		},
		{
			name:       "null name",
			document:   bson.D{{Key: "guild", Value: "guild"}, {Key: "role", Value: "role"}, {Key: "name", Value: nil}},
			wantGuild:  "guild",
			wantRole:   "role",
			wantRename: "",
		},
		{
			name:       "string name",
			document:   bson.D{{Key: "guild", Value: "guild"}, {Key: "role", Value: "role"}, {Key: "name", Value: "  {name} | MHCAT  "}},
			wantGuild:  "guild",
			wantRole:   "role",
			wantRename: "  {name} | MHCAT  ",
		},
		{
			name:       "scalar fields",
			document:   bson.D{{Key: "guild", Value: int64(123)}, {Key: "role", Value: int32(456)}, {Key: "name", Value: true}},
			wantGuild:  "123",
			wantRole:   "456",
			wantRename: "true",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := bson.Marshal(tc.document)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var document VerificationReadDocument
			if err := bson.Unmarshal(encoded, &document); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			config := document.ToDomain()
			if config.GuildID != tc.wantGuild || config.RoleID != tc.wantRole || config.RenameTemplate != tc.wantRename {
				t.Fatalf("config = %#v", config)
			}
		})
	}
}

func TestVerificationReadDocumentRejectsCompoundAndNullRoleShapes(t *testing.T) {
	for _, role := range []any{nil, bson.D{{Key: "bad", Value: true}}, bson.A{"bad"}} {
		encoded, err := bson.Marshal(bson.D{
			{Key: "guild", Value: "guild"},
			{Key: "role", Value: role},
			{Key: "name", Value: bson.D{{Key: "bad", Value: true}}},
		})
		if err != nil {
			t.Fatalf("marshal %#v: %v", role, err)
		}
		var document VerificationReadDocument
		if err := bson.Unmarshal(encoded, &document); err != nil {
			t.Fatalf("unmarshal %#v: %v", role, err)
		}
		config := document.ToDomain()
		if config.RoleID != "" || config.RenameTemplate != "" {
			t.Fatalf("compound values should remain unusable: %#v", config)
		}
		if !errors.Is(config.Validate(), domain.ErrInvalidVerificationConfig) {
			t.Fatalf("role %#v should produce invalid config: %#v", role, config)
		}
	}
}

func TestVerificationWriteDocumentRemainsTyped(t *testing.T) {
	payload, err := bson.Marshal(VerificationDocumentFromDomain(domain.VerificationConfig{
		GuildID: "guild", RoleID: "role", RenameTemplate: "  {name} | MHCAT  ",
	}))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	raw := bson.Raw(payload)
	for _, field := range []string{"guild", "role", "name"} {
		if raw.Lookup(field).Type != bson.TypeString {
			t.Fatalf("field %s type = %s", field, raw.Lookup(field).Type)
		}
	}
	if raw.Lookup("name").StringValue() != "  {name} | MHCAT  " {
		t.Fatalf("name = %q", raw.Lookup("name").StringValue())
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
		{name: "raw channel whitespace", hours: "3600", channel: " channel ", wantSeconds: 3600, wantChannel: " channel "},
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
