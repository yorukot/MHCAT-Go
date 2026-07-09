package gacha

import (
	"context"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
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

func TestPrizeDeleteRemovesPrizeAndRendersLegacySuccess(t *testing.T) {
	repo := fakemongo.NewGachaRepository()
	repo.Prizes["guild-1"] = []domain.GachaPrize{
		{GuildID: "guild-1", Name: "大獎", Chance: 10, Count: 1},
		{GuildID: "guild-1", Name: "小獎", Chance: 90, Count: 9},
	}
	usage := &fakeusage.Tracker{}
	module := NewDeleteModule(repo, usage)
	responder := fakediscord.NewResponder()
	interaction := gachaPrizeDeleteInteraction("大獎")

	if err := module.PrizeDeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 || responder.Defers[0].Ephemeral {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != "<a:green_tick:994529015652163614>成功刪除!" || embed.Description != "獎品名:大獎" || embed.Color != gachaPrizeDeleteSuccessColor {
		t.Fatalf("embed = %#v", embed)
	}
	if len(repo.Prizes["guild-1"]) != 1 || repo.Prizes["guild-1"][0].Name != "小獎" {
		t.Fatalf("remaining prizes = %#v", repo.Prizes["guild-1"])
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != GachaPrizeDeleteCommandName || usage.Events[0].Feature != "gacha-prize-delete" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestPrizeDeleteMissingPrizeReturnsLegacyError(t *testing.T) {
	module := NewDeleteModule(fakemongo.NewGachaRepository(), nil)
	responder := fakediscord.NewResponder()

	if err := module.PrizeDeleteHandler()(context.Background(), gachaPrizeDeleteInteraction("不存在"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<a:Discord_AnimatedNo:1015989839809757295> | 找不到這個獎品!" {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestPrizeDeleteRequiresManageMessages(t *testing.T) {
	repo := fakemongo.NewGachaRepository()
	repo.Prizes["guild-1"] = []domain.GachaPrize{{GuildID: "guild-1", Name: "大獎", Chance: 10, Count: 1}}
	module := NewDeleteModule(repo, nil)
	responder := fakediscord.NewResponder()
	interaction := gachaPrizeDeleteInteraction("大獎")
	interaction.Actor.PermissionBits = 0

	if err := module.PrizeDeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "訊息管理") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	if len(repo.Prizes["guild-1"]) != 1 {
		t.Fatalf("prize should not be deleted without permission: %#v", repo.Prizes["guild-1"])
	}
}

func TestPrizeCreateStoresPrizeAndRendersLegacySuccess(t *testing.T) {
	repo := fakemongo.NewGachaRepository()
	usage := &fakeusage.Tracker{}
	module := NewCreateModule(repo, usage)
	responder := fakediscord.NewResponder()
	interaction := gachaPrizeCreateInteraction("大獎")
	interaction.Options[gachaPrizeCodeOption] = "code-1"
	interaction.CommandOptions = map[string]interactions.CommandOptionValue{
		gachaPrizeChanceOption:     {Type: interactions.CommandOptionNumber, Float: 12.5},
		gachaPrizeAutoDeleteOption: {Type: interactions.CommandOptionBoolean, Bool: false},
		gachaPrizeCountOption:      {Type: interactions.CommandOptionInteger, Int: 3},
		gachaPrizeGiveCoinOption:   {Type: interactions.CommandOptionInteger, Int: 7},
	}

	if err := module.PrizeCreateHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 || !responder.Defers[0].Ephemeral {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != "<a:green_tick:994529015652163614>設置成功" || embed.Color != gachaPrizeCreateSuccessColor || len(embed.Fields) != 6 {
		t.Fatalf("embed = %#v", embed)
	}
	for _, want := range []struct {
		index int
		name  string
		value string
	}{
		{0, "<:id:985950321975128094> **獎品名:**", "大獎"},
		{1, "<:dice:997374185322057799> **獎品機率:**", "12.5"},
		{2, "<:security:997374179257102396> **獎品代碼:**", "code-1"},
		{3, "<:counter:994585977207140423> **獎品數量:**", "3個"},
		{4, "<:trashbin:995991389043163257> **自動刪除:**", "false"},
		{5, "<:givemoney:1019632789110399068> **給予代幣數:**", "7個"},
	} {
		field := embed.Fields[want.index]
		if field.Name != want.name || field.Value != want.value || !field.Inline {
			t.Fatalf("field %d = %#v", want.index, field)
		}
	}
	if len(repo.PrizeConfigs["guild-1"]) != 1 {
		t.Fatalf("prize configs = %#v", repo.PrizeConfigs["guild-1"])
	}
	saved := repo.PrizeConfigs["guild-1"][0]
	if saved.Name != "大獎" || saved.Code != "code-1" || saved.Chance != 12.5 || saved.AutoDelete || saved.Count != 3 || saved.GiveCoin != 7 {
		t.Fatalf("saved prize = %#v", saved)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != GachaPrizeCreateCommandName || usage.Events[0].Feature != "gacha-prize-create" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestPrizeCreateDefaultsOptionalLegacyValues(t *testing.T) {
	repo := fakemongo.NewGachaRepository()
	module := NewCreateModule(repo, nil)
	responder := fakediscord.NewResponder()
	interaction := gachaPrizeCreateInteraction("無代碼")
	interaction.Options[gachaPrizeChanceOption] = "0.1"
	interaction.Options[gachaPrizeCountOption] = "0"

	if err := module.PrizeCreateHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(repo.PrizeConfigs["guild-1"]) != 1 {
		t.Fatalf("prize configs = %#v", repo.PrizeConfigs["guild-1"])
	}
	saved := repo.PrizeConfigs["guild-1"][0]
	if saved.Chance != 0.1 || !saved.AutoDelete || saved.Count != 1 || saved.GiveCoin != 0 {
		t.Fatalf("saved defaults = %#v", saved)
	}
	fields := responder.Edits[0].Embeds[0].Fields
	if fields[2].Value != "該獎品無代碼" || fields[3].Value != "1個" || fields[4].Value != "true" || fields[5].Value != "0個" {
		t.Fatalf("fields = %#v", fields)
	}
}

func TestPrizeCreateRequiresManageMessages(t *testing.T) {
	repo := fakemongo.NewGachaRepository()
	module := NewCreateModule(repo, nil)
	responder := fakediscord.NewResponder()
	interaction := gachaPrizeCreateInteraction("大獎")
	interaction.Actor.PermissionBits = 0

	if err := module.PrizeCreateHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<a:Discord_AnimatedNo:1015989839809757295> | 你沒有權限使用這個指令" {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	if len(repo.Prizes["guild-1"]) != 0 {
		t.Fatalf("prize should not be created without permission: %#v", repo.Prizes["guild-1"])
	}
}

func TestPrizeCreateDuplicateReturnsLegacyError(t *testing.T) {
	repo := fakemongo.NewGachaRepository()
	repo.Prizes["guild-1"] = []domain.GachaPrize{{GuildID: "guild-1", Name: "大獎", Count: 1}}
	module := NewCreateModule(repo, nil)
	responder := fakediscord.NewResponder()

	if err := module.PrizeCreateHandler()(context.Background(), gachaPrizeCreateInteraction("大獎"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<a:Discord_AnimatedNo:1015989839809757295> | 獎品名稱重複，請將之前的刪除或等待被抽中!" {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestPrizeCreateFullPoolReturnsLegacyError(t *testing.T) {
	repo := fakemongo.NewGachaRepository()
	for i := 0; i < 25; i++ {
		repo.Prizes["guild-1"] = append(repo.Prizes["guild-1"], domain.GachaPrize{GuildID: "guild-1", Name: "獎品", Count: 1})
	}
	module := NewCreateModule(repo, nil)
	responder := fakediscord.NewResponder()

	if err := module.PrizeCreateHandler()(context.Background(), gachaPrizeCreateInteraction("新獎品"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<a:Discord_AnimatedNo:1015989839809757295> | 你的獎品數量已經過多了!!請先刪除部分獎品!" {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestPrizeCreateNegativeCountReturnsLegacyError(t *testing.T) {
	module := NewCreateModule(fakemongo.NewGachaRepository(), nil)
	responder := fakediscord.NewResponder()
	interaction := gachaPrizeCreateInteraction("大獎")
	interaction.Options[gachaPrizeCountOption] = "-1"

	if err := module.PrizeCreateHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<a:Discord_AnimatedNo:1015989839809757295> | 獎品必須大於1" {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestPrizeEditUpdatesPrizeAndRendersLegacySuccess(t *testing.T) {
	repo := fakemongo.NewGachaRepository()
	repo.Prizes["guild-1"] = []domain.GachaPrize{{GuildID: "guild-1", Name: "大獎", Chance: 10, Count: 2}}
	repo.PrizeConfigs["guild-1"] = []domain.GachaPrizeConfig{{
		GuildID:    "guild-1",
		Name:       "大獎",
		Code:       "old-code",
		Chance:     10,
		AutoDelete: false,
		Count:      2,
		GiveCoin:   5,
	}}
	usage := &fakeusage.Tracker{}
	module := NewEditModule(repo, usage)
	responder := fakediscord.NewResponder()
	interaction := gachaPrizeEditInteraction("大獎")
	interaction.Options[gachaPrizeCodeOption] = "new-code"
	interaction.CommandOptions = map[string]interactions.CommandOptionValue{
		gachaPrizeChanceOption:     {Type: interactions.CommandOptionNumber, Float: 12.5},
		gachaPrizeAutoDeleteOption: {Type: interactions.CommandOptionBoolean, Bool: true},
		gachaPrizeCountOption:      {Type: interactions.CommandOptionInteger, Int: 3},
		gachaPrizeGiveCoinOption:   {Type: interactions.CommandOptionInteger, Int: 7},
	}

	if err := module.PrizeEditHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 || !responder.Defers[0].Ephemeral {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != "<a:green_tick:994529015652163614>編輯成功成功" || embed.Color != gachaPrizeEditSuccessColor || len(embed.Fields) != 6 {
		t.Fatalf("embed = %#v", embed)
	}
	for _, want := range []struct {
		index int
		value string
	}{
		{0, "大獎"},
		{1, "12.5"},
		{2, "new-code"},
		{3, "3個"},
		{4, "true"},
		{5, "7個"},
	} {
		if embed.Fields[want.index].Value != want.value || !embed.Fields[want.index].Inline {
			t.Fatalf("field %d = %#v", want.index, embed.Fields[want.index])
		}
	}
	saved := repo.PrizeConfigs["guild-1"][0]
	if saved.Code != "new-code" || saved.Chance != 12.5 || !saved.AutoDelete || saved.Count != 3 || saved.GiveCoin != 7 {
		t.Fatalf("saved = %#v", saved)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != GachaPrizeEditCommandName || usage.Events[0].Feature != "gacha-prize-edit" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestPrizeEditDefaultsUIAndPreservesLegacyFalseyFields(t *testing.T) {
	repo := fakemongo.NewGachaRepository()
	repo.Prizes["guild-1"] = []domain.GachaPrize{{GuildID: "guild-1", Name: "大獎", Chance: 10, Count: 2}}
	repo.PrizeConfigs["guild-1"] = []domain.GachaPrizeConfig{{
		GuildID:    "guild-1",
		Name:       "大獎",
		Code:       "old-code",
		Chance:     10,
		AutoDelete: false,
		Count:      2,
		GiveCoin:   5,
	}}
	module := NewEditModule(repo, nil)
	responder := fakediscord.NewResponder()
	interaction := gachaPrizeEditInteraction("大獎")
	interaction.Options[gachaPrizeAutoDeleteOption] = "false"
	interaction.Options[gachaPrizeGiveCoinOption] = "0"

	if err := module.PrizeEditHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	fields := responder.Edits[0].Embeds[0].Fields
	for _, want := range []struct {
		index int
		value string
	}{
		{1, "null"},
		{2, "該獎品無代碼"},
		{3, "1個"},
		{4, "false"},
		{5, "0個"},
	} {
		if fields[want.index].Value != want.value {
			t.Fatalf("field %d = %#v", want.index, fields[want.index])
		}
	}
	saved := repo.PrizeConfigs["guild-1"][0]
	if saved.Code != "old-code" || saved.Chance != 10 || saved.AutoDelete || saved.Count != 1 || saved.GiveCoin != 5 {
		t.Fatalf("saved = %#v", saved)
	}
}

func TestPrizeEditMissingPrizeReturnsLegacyError(t *testing.T) {
	module := NewEditModule(fakemongo.NewGachaRepository(), nil)
	responder := fakediscord.NewResponder()

	if err := module.PrizeEditHandler()(context.Background(), gachaPrizeEditInteraction("不存在"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<a:Discord_AnimatedNo:1015989839809757295> | 找不到這個獎品!" {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestPrizeEditRequiresManageMessages(t *testing.T) {
	repo := fakemongo.NewGachaRepository()
	repo.Prizes["guild-1"] = []domain.GachaPrize{{GuildID: "guild-1", Name: "大獎", Chance: 10, Count: 2}}
	module := NewEditModule(repo, nil)
	responder := fakediscord.NewResponder()
	interaction := gachaPrizeEditInteraction("大獎")
	interaction.Actor.PermissionBits = 0

	if err := module.PrizeEditHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<a:Discord_AnimatedNo:1015989839809757295> | 你沒有權限使用這個指令" {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	if repo.Prizes["guild-1"][0].Count != 2 {
		t.Fatalf("prize should not be edited without permission: %#v", repo.Prizes["guild-1"])
	}
}

func TestPrizeEditNegativeCountReturnsLegacyError(t *testing.T) {
	module := NewEditModule(fakemongo.NewGachaRepository(), nil)
	responder := fakediscord.NewResponder()
	interaction := gachaPrizeEditInteraction("大獎")
	interaction.Options[gachaPrizeCountOption] = "-1"

	if err := module.PrizeEditHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<a:Discord_AnimatedNo:1015989839809757295> | 獎品必須大於1" {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func gachaPrizeDeleteInteraction(prizeName string) interactions.Interaction {
	interaction := fakediscord.SlashInteractionWithOptions(GachaPrizeDeleteCommandName, "", map[string]string{gachaPrizeNameOption: prizeName})
	interaction.Actor.PermissionBits = gachaManageMessagesPermissionBit
	return interaction
}

func gachaPrizeCreateInteraction(prizeName string) interactions.Interaction {
	interaction := fakediscord.SlashInteractionWithOptions(GachaPrizeCreateCommandName, "", map[string]string{
		gachaPrizeNameOption:   prizeName,
		gachaPrizeChanceOption: "10",
	})
	interaction.Actor.PermissionBits = gachaManageMessagesPermissionBit
	return interaction
}

func gachaPrizeEditInteraction(prizeName string) interactions.Interaction {
	interaction := fakediscord.SlashInteractionWithOptions(GachaPrizeEditCommandName, "", map[string]string{
		gachaPrizeNameOption: prizeName,
	})
	interaction.Actor.PermissionBits = gachaManageMessagesPermissionBit
	return interaction
}
