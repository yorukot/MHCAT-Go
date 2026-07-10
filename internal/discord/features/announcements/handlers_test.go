package announcements

import (
	"context"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestConfigHandlerSetsOnceAnnouncementChannel(t *testing.T) {
	repo := fakemongo.NewAnnouncementConfigRepository()
	module := NewModule(repo)
	interaction := announcementInteraction(subcommandOnce, map[string]string{optionChannel: "channel-1"})
	responder := fakediscord.NewResponder()

	if err := module.ConfigHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 || len(responder.Edits) != 1 {
		t.Fatalf("defers=%#v edits=%#v", responder.Defers, responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != "<:megaphone:985943890148327454> 公告系統" || !strings.Contains(embed.Description, "成功__創建__!!") || !strings.Contains(embed.Description, "<#channel-1>") {
		t.Fatalf("embed = %#v", embed)
	}
	if repo.AnnouncementChannels["guild-1"] != "channel-1" {
		t.Fatalf("repo state = %#v", repo.AnnouncementChannels)
	}
}

func TestConfigHandlerUpdatesOnceAnnouncementChannel(t *testing.T) {
	repo := fakemongo.NewAnnouncementConfigRepository()
	repo.AnnouncementChannels["guild-1"] = "old-channel"
	module := NewModule(repo)
	interaction := announcementInteraction(subcommandOnce, map[string]string{optionChannel: "channel-2"})
	responder := fakediscord.NewResponder()

	if err := module.ConfigHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Description, "成功__更新__!!") {
		t.Fatalf("edit = %#v", responder.Edits[0])
	}
	if repo.AnnouncementChannels["guild-1"] != "channel-2" {
		t.Fatalf("repo state = %#v", repo.AnnouncementChannels)
	}
}

func TestConfigHandlerSetsBoundAnnouncement(t *testing.T) {
	repo := fakemongo.NewAnnouncementConfigRepository()
	module := NewModule(repo)
	interaction := announcementInteraction(subcommandBound, map[string]string{
		optionChannel: "channel-1",
		optionTag:     "@here",
		optionColor:   "#53FF53",
		optionTitle:   "公告",
	})
	responder := fakediscord.NewResponder()

	if err := module.ConfigHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != "<:megaphone:985943890148327454> 綁定型公告系統" || !strings.Contains(embed.Description, "成功__創建__!!") {
		t.Fatalf("embed = %#v", embed)
	}
	saved := repo.BoundAnnouncements["guild-1:channel-1"]
	if saved != (domain.BoundAnnouncementConfig{GuildID: "guild-1", ChannelID: "channel-1", Tag: "@here", Color: "#53FF53", Title: "公告"}) {
		t.Fatalf("saved = %#v", saved)
	}
}

func TestConfigHandlerAcceptsLegacyRandomColor(t *testing.T) {
	repo := fakemongo.NewAnnouncementConfigRepository()
	module := NewModule(repo)
	interaction := announcementInteraction(subcommandBound, map[string]string{
		optionChannel: "channel-1",
		optionTag:     "@everyone",
		optionColor:   "Random",
		optionTitle:   "公告",
	})
	responder := fakediscord.NewResponder()

	if err := module.ConfigHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if repo.BoundAnnouncements["guild-1:channel-1"].Color != "Random" {
		t.Fatalf("saved = %#v", repo.BoundAnnouncements)
	}
}

func TestConfigHandlerPreservesLegacyRawBoundValues(t *testing.T) {
	repo := fakemongo.NewAnnouncementConfigRepository()
	module := NewModule(repo)
	interaction := announcementInteraction(subcommandBound, map[string]string{
		optionChannel: "channel-1",
		optionTag:     " ",
		optionColor:   "rgb(0 0 0)",
		optionTitle:   "\t",
	})
	responder := fakediscord.NewResponder()

	if err := module.ConfigHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	saved := repo.BoundAnnouncements["guild-1:channel-1"]
	if saved.Tag != " " || saved.Color != "rgb(0 0 0)" || saved.Title != "\t" {
		t.Fatalf("raw values = %#v", saved)
	}
}

func TestConfigHandlerDoesNotTrimColorIntoValidity(t *testing.T) {
	module := NewModule(fakemongo.NewAnnouncementConfigRepository())
	for _, color := range []string{"53FF53", " #53FF53"} {
		interaction := announcementInteraction(subcommandBound, map[string]string{
			optionChannel: "channel-1",
			optionTag:     "@here",
			optionColor:   color,
			optionTitle:   "公告",
		})
		responder := fakediscord.NewResponder()
		if err := module.ConfigHandler()(context.Background(), interaction, responder); err != nil {
			t.Fatalf("handler %q: %v", color, err)
		}
		if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<a:Discord_AnimatedNo:1015989839809757295> | 你傳送的並不是顏色(色碼)" {
			t.Fatalf("color %q response = %#v", color, responder.Edits)
		}
	}
}

func TestConfigHandlerRejectsPermissionAndInvalidColor(t *testing.T) {
	module := NewModule(fakemongo.NewAnnouncementConfigRepository())
	interaction := fakediscord.SlashInteractionWithOptions(ConfigCommandName, subcommandOnce, map[string]string{optionChannel: "channel-1"})
	responder := fakediscord.NewResponder()

	if err := module.ConfigHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Title, "你需要有`undefined`才能使用此指令") || responder.Edits[0].Embeds[0].Color != 0xED4245 {
		t.Fatalf("permission response = %#v", responder.Edits)
	}

	interaction = announcementInteraction(subcommandBound, map[string]string{
		optionChannel: "channel-1",
		optionTag:     "@here",
		optionColor:   "not-a-color",
		optionTitle:   "公告",
	})
	responder = fakediscord.NewResponder()
	if err := module.ConfigHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Title, "你傳送的並不是顏色(色碼)") || responder.Edits[0].Embeds[0].Color != 0xED4245 {
		t.Fatalf("color response = %#v", responder.Edits)
	}
}

func TestConfigHandlerDeletesBoundAnnouncement(t *testing.T) {
	repo := fakemongo.NewAnnouncementConfigRepository()
	repo.BoundAnnouncements["guild-1:channel-1"] = domain.BoundAnnouncementConfig{GuildID: "guild-1", ChannelID: "channel-1"}
	module := NewModule(repo)
	interaction := announcementInteraction(subcommandDeleteBound, map[string]string{optionChannel: "channel-1"})
	responder := fakediscord.NewResponder()

	if err := module.ConfigHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if _, exists := repo.BoundAnnouncements["guild-1:channel-1"]; exists {
		t.Fatal("bound announcement config was not deleted")
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Description, "成功__刪除__!!") || !strings.Contains(responder.Edits[0].Embeds[0].Description, "<#channel-1>") {
		t.Fatalf("delete response = %#v", responder.Edits)
	}

	responder = fakediscord.NewResponder()
	if err := module.ConfigHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler missing: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Title, "你沒有對這個頻道設定過綁定型公告!") {
		t.Fatalf("missing response = %#v", responder.Edits)
	}
}

func announcementInteraction(subcommand string, options map[string]string) interactions.Interaction {
	interaction := fakediscord.SlashInteractionWithOptions(ConfigCommandName, subcommand, options)
	interaction.Actor.PermissionBits = permissionManageMessages
	return interaction
}
