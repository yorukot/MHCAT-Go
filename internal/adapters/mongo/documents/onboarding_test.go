package documents

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
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
