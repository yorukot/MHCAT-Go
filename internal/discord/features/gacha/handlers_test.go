package gacha

import (
	"context"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakebotinfo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestPrizeListRendersLegacyEmbed(t *testing.T) {
	repo := fakemongo.NewGachaRepository()
	repo.Prizes["guild-1"] = []domain.GachaPrize{
		{GuildID: "guild-1", Name: "大獎", Chance: 12.5, Count: 3},
		{GuildID: "guild-1", Name: "小獎", Chance: 87.5, Count: 9},
	}
	repo.Configs["guild-1"] = domain.EconomyConfig{GuildID: "guild-1", GachaCost: 700, SignCoins: 40, XPMultiple: 2.5}
	discordInfo := &fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{Name: "測試群"}}
	usage := &fakeusage.Tracker{}
	module := NewModuleWithColor(repo, discordInfo, usage, func() int { return 0x123456 })
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction(GachaPrizeListCommandName)
	interaction.Actor.UserTag = "Yoru#0001"
	interaction.Actor.AvatarURL = "https://example.invalid/avatar.png"

	if err := module.PrizeListHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != "<:list:992002476360343602> 以下是測試群的獎池" || embed.Color != 0x123456 {
		t.Fatalf("embed = %#v", embed)
	}
	for _, want := range []string{
		"**<:money:997374193026994236> 扭蛋所需代幣:**`700`個",
		"<:calendar:990254384812290048> **簽到給予代幣數:**`40`個",
		"**<:levelup:990254382845157406> 等級提升給予倍數:**`2.5`倍",
	} {
		if !strings.Contains(embed.Description, want) {
			t.Fatalf("description missing %q: %q", want, embed.Description)
		}
	}
	if len(embed.Fields) != 2 || embed.Fields[0].Name != "<:id:985950321975128094> 獎品名: 大獎" {
		t.Fatalf("fields = %#v", embed.Fields)
	}
	if !strings.Contains(embed.Fields[0].Value, "`12.5`%") || !strings.Contains(embed.Fields[0].Value, "'<:counter:994585977207140423> **獎品數量:** `3`") {
		t.Fatalf("field value = %q", embed.Fields[0].Value)
	}
	if embed.Footer == nil || embed.Footer.Text != "Yoru#0001的查詢" || embed.Footer.IconURL != "https://example.invalid/avatar.png" {
		t.Fatalf("footer = %#v", embed.Footer)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != GachaPrizeListCommandName || usage.Events[0].Feature != "gacha-prize-list" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestPrizeListUsesLegacyDefaultsWhenConfigMissing(t *testing.T) {
	repo := fakemongo.NewGachaRepository()
	repo.Prizes["guild-1"] = []domain.GachaPrize{{GuildID: "guild-1", Name: "大獎", Chance: 10, Count: 1}}
	module := NewModuleWithColor(repo, nil, nil, func() int { return 0 })
	responder := fakediscord.NewResponder()

	if err := module.PrizeListHandler()(context.Background(), fakediscord.SlashInteraction(GachaPrizeListCommandName), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	description := responder.Edits[0].Embeds[0].Description
	if !strings.Contains(description, "扭蛋所需代幣:**`500`個") || !strings.Contains(description, "簽到給予代幣數:**`25`個") || !strings.Contains(description, "等級提升給予倍數:**`0`倍") {
		t.Fatalf("description = %q", description)
	}
}

func TestPrizeListEmptyPoolReturnsLegacyError(t *testing.T) {
	module := NewModule(fakemongo.NewGachaRepository(), nil, nil)
	responder := fakediscord.NewResponder()
	if err := module.PrizeListHandler()(context.Background(), fakediscord.SlashInteraction(GachaPrizeListCommandName), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "目前獎池沒有任何獎品喔") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestPrizeListSplitsMoreThanDiscordFieldLimit(t *testing.T) {
	repo := fakemongo.NewGachaRepository()
	for i := 0; i < 26; i++ {
		repo.Prizes["guild-1"] = append(repo.Prizes["guild-1"], domain.GachaPrize{GuildID: "guild-1", Name: "獎品", Chance: 1, Count: 1})
	}
	module := NewModuleWithColor(repo, nil, nil, func() int { return 0x123456 })
	responder := fakediscord.NewResponder()
	if err := module.PrizeListHandler()(context.Background(), fakediscord.SlashInteraction(GachaPrizeListCommandName), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 2 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	if len(responder.Edits[0].Embeds[0].Fields) != 25 || len(responder.Edits[0].Embeds[1].Fields) != 1 {
		t.Fatalf("field counts = %d/%d", len(responder.Edits[0].Embeds[0].Fields), len(responder.Edits[0].Embeds[1].Fields))
	}
	if responder.Edits[0].Embeds[1].Description != "" || responder.Edits[0].Embeds[1].Footer != nil {
		t.Fatalf("second embed should not duplicate description/footer: %#v", responder.Edits[0].Embeds[1])
	}
}
