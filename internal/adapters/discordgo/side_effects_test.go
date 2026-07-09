package discordgo

import (
	"context"
	"testing"
	"time"

	dgo "github.com/bwmarrin/discordgo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

func TestSideEffectClientRequiresSession(t *testing.T) {
	client := SideEffectClient{}
	if _, err := client.SendMessage(context.Background(), "channel-1", ports.OutboundMessage{Content: "hello"}); err == nil {
		t.Fatal("expected send error")
	}
	if err := client.DeleteChannel(context.Background(), "channel-1"); err == nil {
		t.Fatal("expected delete channel error")
	}
	if err := client.AddRole(context.Background(), "guild-1", "user-1", "role-1"); err == nil {
		t.Fatal("expected add role error")
	}
}

func TestCoreAllowedMentionsSuppressesByDefault(t *testing.T) {
	allowed := coreAllowedMentions(ports.AllowedMentions{})
	if allowed == nil || len(allowed.Parse) != 0 || len(allowed.Users) != 0 || len(allowed.Roles) != 0 {
		t.Fatalf("allowed mentions = %#v", allowed)
	}
}

func TestOutboundMessageConversionIncludesEmbedsAndButtons(t *testing.T) {
	embeds := outboundEmbeds([]ports.OutboundEmbed{{
		Title:         "__**私人頻道**__",
		Description:   "你開啟了一個私人頻道，請等待客服人員的回復!",
		Color:         0x00DB00,
		FooterText:    "來自tester的公告",
		FooterIconURL: "https://example.invalid/avatar.png",
		ThumbnailURL:  "https://example.invalid/thumb.png",
		ImageURL:      "https://example.invalid/image.png",
		Timestamp:     time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC),
	}})
	if len(embeds) != 1 || embeds[0].Title != "__**私人頻道**__" || embeds[0].Color != 0x00DB00 {
		t.Fatalf("embeds = %#v", embeds)
	}
	if embeds[0].Footer == nil || embeds[0].Footer.Text != "來自tester的公告" {
		t.Fatalf("footer = %#v", embeds[0].Footer)
	}
	if embeds[0].Thumbnail == nil || embeds[0].Thumbnail.URL != "https://example.invalid/thumb.png" {
		t.Fatalf("thumbnail = %#v", embeds[0].Thumbnail)
	}
	if embeds[0].Image == nil || embeds[0].Image.URL != "https://example.invalid/image.png" {
		t.Fatalf("image = %#v", embeds[0].Image)
	}
	if embeds[0].Timestamp != "2026-07-04T00:00:00Z" {
		t.Fatalf("timestamp = %q", embeds[0].Timestamp)
	}

	components := outboundComponents([]ports.OutboundComponentRow{{Components: []ports.OutboundComponent{{
		Type:     "button",
		CustomID: "del",
		Label:    "🗑️ 刪除!",
		Style:    "danger",
	}}}})
	if len(components) != 1 {
		t.Fatalf("components = %#v", components)
	}
	row, ok := components[0].(dgo.ActionsRow)
	if !ok || len(row.Components) != 1 {
		t.Fatalf("row = %#v", components[0])
	}
	button, ok := row.Components[0].(dgo.Button)
	if !ok {
		t.Fatalf("button = %#v", row.Components[0])
	}
	if button.CustomID != "del" || button.Label != "🗑️ 刪除!" || button.Style != dgo.DangerButton {
		t.Fatalf("button = %#v", button)
	}
}

func TestOutboundMessageConversionIncludesSelectMenusAndEmojis(t *testing.T) {
	components := outboundComponents([]ports.OutboundComponentRow{{Components: []ports.OutboundComponent{{
		Type:        "select",
		CustomID:    "mhcat:v1:poll:owner_menu:",
		Placeholder: "🔧投票發起人操作",
		Options: []ports.OutboundSelectOption{{
			Label:       "公開投票結果",
			Description: "讓所有成員都可以查看該投票結果",
			Value:       "poll_public_result",
			Emoji:       "<:publicrelation:1023972880385585212>",
		}},
	}}}})
	if len(components) != 1 {
		t.Fatalf("components = %#v", components)
	}
	row, ok := components[0].(dgo.ActionsRow)
	if !ok || len(row.Components) != 1 {
		t.Fatalf("row = %#v", components[0])
	}
	menu, ok := row.Components[0].(dgo.SelectMenu)
	if !ok {
		t.Fatalf("menu = %#v", row.Components[0])
	}
	if menu.CustomID != "mhcat:v1:poll:owner_menu:" || menu.Placeholder != "🔧投票發起人操作" || len(menu.Options) != 1 {
		t.Fatalf("menu = %#v", menu)
	}
	if menu.Options[0].Emoji.Name != "publicrelation" || menu.Options[0].Emoji.ID != "1023972880385585212" {
		t.Fatalf("option emoji = %#v", menu.Options[0].Emoji)
	}
}
