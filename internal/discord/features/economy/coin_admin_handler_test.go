package economy

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

func TestCoinAdminDefinitionMatchesLegacyCommand(t *testing.T) {
	definition := CoinAdminDefinition()
	if definition.Name != "代幣增加" || definition.Description != "改變扭蛋數量" {
		t.Fatalf("definition = %#v", definition)
	}
	if len(definition.Options) != 3 {
		t.Fatalf("options = %#v", definition.Options)
	}
	if definition.Options[0].Name != "使用者" || !definition.Options[0].Required {
		t.Fatalf("user option = %#v", definition.Options[0])
	}
	action := definition.Options[1]
	if action.Name != "增加或減少" || action.Description != "輸入這個獎品叫甚麼，以及簡單概述" || len(action.Choices) != 2 {
		t.Fatalf("action option = %#v", action)
	}
	if definition.Options[2].Name != "數量" || !definition.Options[2].Required {
		t.Fatalf("amount option = %#v", definition.Options[2])
	}
}

func TestCoinAdminRequiresManageMessages(t *testing.T) {
	module := NewCoinAdminModule(fakemongo.NewEconomyRepository(), nil, nil)
	responder := fakediscord.NewResponder()
	interaction := coinAdminInteraction("target", "add", 10)
	if err := module.CoinAdminHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "你需要有`訊息管理`") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestCoinAdminAddCreatesBalanceAndRendersLegacySuccess(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	usage := &fakeusage.Tracker{}
	info := &fakebotinfo.DiscordInfoProvider{Users: map[string]ports.DiscordUserInfo{"target": {Username: "Target"}}}
	module := NewCoinAdminModule(repo, info, usage)
	module.color = func() int { return 0x123456 }
	responder := fakediscord.NewResponder()
	interaction := coinAdminInteraction("target", "add", 10)
	interaction.Actor.PermissionBits = coinAdminManageMessagesPermission
	interaction.Actor.AvatarURL = "https://example.test/mod.png"
	interaction.Actor.GuildAvatarURL = "https://example.test/guild-mod.gif"
	if err := module.CoinAdminHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	balance, err := repo.GetCoinBalance(context.Background(), "guild-1", "target")
	if err != nil || balance.Coins != 10 {
		t.Fatalf("balance = %#v err=%v", balance, err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<:money:997374193026994236>已為Target`增加`:`10`個代幣!" {
		t.Fatalf("success edit = %#v", responder.Edits)
	}
	if responder.Edits[0].Embeds[0].Footer == nil || responder.Edits[0].Embeds[0].Footer.Text != "增加10" {
		t.Fatalf("footer = %#v", responder.Edits[0].Embeds[0].Footer)
	}
	if responder.Edits[0].Embeds[0].Footer.IconURL != "https://example.test/guild-mod.gif" || responder.Edits[0].Embeds[0].Color != 0x123456 {
		t.Fatalf("success embed = %#v", responder.Edits[0].Embeds[0])
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != CoinAdminCommandName || usage.Events[0].Feature != "economy-coin-admin" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestCoinAdminReduceRejectsNegativeBalance(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "target", Coins: 3})
	module := NewCoinAdminModule(repo, nil, nil)
	responder := fakediscord.NewResponder()
	interaction := coinAdminInteraction("target", "reduce", 4)
	interaction.Actor.PermissionBits = coinAdminManageMessagesPermission
	if err := module.CoinAdminHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<a:error:980086028113182730> | 不可減到負數!" {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func coinAdminInteraction(userID string, operation string, amount int64) interactions.Interaction {
	interaction := fakediscord.SlashInteractionWithOptions(CoinAdminCommandName, "", map[string]string{
		coinAdminOptionUser:      userID,
		coinAdminOptionOperation: operation,
		coinAdminOptionAmount:    "0",
	})
	interaction.CommandOptions = map[string]interactions.CommandOptionValue{
		coinAdminOptionAmount: {Type: interactions.CommandOptionInteger, Int: amount},
	}
	return interaction
}
