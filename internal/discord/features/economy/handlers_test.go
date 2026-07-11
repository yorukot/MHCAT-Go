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

func TestCoinQuerySelfUsesLegacyDefaultConfigFooter(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 125})
	usage := &fakeusage.Tracker{}
	module := NewModule(repo, nil, usage)
	module.color = func() int { return 0x123456 }
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction("代幣查詢")
	interaction.Actor.AvatarURL = "https://cdn.example/avatar.png"

	if err := module.CoinQueryHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handle coin query: %v", err)
	}
	if len(responder.Defers) != 1 || !responder.Defers[0].Ephemeral {
		t.Fatalf("expected one ephemeral defer, got %#v", responder.Defers)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("expected one edit embed, got %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != "<:money:997374193026994236>你目前有:`125`個代幣!" {
		t.Fatalf("unexpected title: %q", embed.Title)
	}
	if embed.Color != 0x123456 {
		t.Fatalf("success color = %#x", embed.Color)
	}
	if !strings.Contains(embed.Description, "代幣數到了500可以進行扭蛋喔") {
		t.Fatalf("description did not include default gacha cost: %q", embed.Description)
	}
	if embed.Footer == nil || embed.Footer.Text != "你還差:500" || embed.Footer.IconURL != "https://cdn.example/avatar.png" {
		t.Fatalf("unexpected footer: %#v", embed.Footer)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != "代幣查詢" || usage.Events[0].Feature != "economy-query" {
		t.Fatalf("unexpected usage events: %#v", usage.Events)
	}
}

func TestCoinQuerySelectedUserUsesFetchedUsername(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-2", Coins: 800})
	repo.PutConfig(domain.EconomyConfig{GuildID: "guild-1", GachaCost: 700})
	discordInfo := &fakebotinfo.DiscordInfoProvider{User: ports.DiscordUserInfo{ID: "user-2", Username: "TargetUser"}}
	module := NewModule(repo, discordInfo, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions("代幣查詢", "", map[string]string{"使用者": "user-2"})

	if err := module.CoinQueryHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handle coin query: %v", err)
	}
	if len(discordInfo.UserCalls) != 1 || discordInfo.UserCalls[0] != "guild-1:user-2" {
		t.Fatalf("user lookup calls = %#v", discordInfo.UserCalls)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != "<:money:997374193026994236>TargetUser目前有:`800`個代幣!" {
		t.Fatalf("unexpected title: %q", embed.Title)
	}
	if embed.Footer == nil || embed.Footer.Text != "TargetUser還差:你可以扭蛋了!!使用`/扭蛋`進行扭蛋" {
		t.Fatalf("unexpected footer: %#v", embed.Footer)
	}
}

func TestCoinQuerySelectedUserPrefersResolvedUsername(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-2", Coins: 10})
	module := NewModule(repo, nil, nil)
	interaction := fakediscord.SlashInteractionWithOptions("代幣查詢", "", map[string]string{"使用者": "user-2"})
	interaction.CommandOptions = map[string]interactions.CommandOptionValue{
		"使用者": {Type: interactions.CommandOptionUser, String: "user-2", UserName: "ResolvedUser"},
	}
	responder := fakediscord.NewResponder()
	if err := module.CoinQueryHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handle coin query: %v", err)
	}
	if got := responder.Edits[0].Embeds[0].Title; got != "<:money:997374193026994236>ResolvedUser目前有:`10`個代幣!" {
		t.Fatalf("title = %q", got)
	}
}

func TestCoinQueryNoBalanceUsesLegacyErrorEmbed(t *testing.T) {
	module := NewModule(fakemongo.NewEconomyRepository(), nil, nil)
	responder := fakediscord.NewResponder()
	if err := module.CoinQueryHandler()(context.Background(), fakediscord.SlashInteraction("代幣查詢"), responder); err != nil {
		t.Fatalf("handle coin query: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("expected one error edit, got %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != legacyCoinNoBalanceTitle || embed.Color != coinQueryErrorColor {
		t.Fatalf("unexpected no-balance embed: %#v", embed)
	}
}

func TestCoinQueryRendersLegacyConfiguredNumberEdges(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		wantCost   string
		wantFooter string
	}{
		{name: "undefined", text: "undefined", wantCost: "代幣數到了undefined", wantFooter: "你還差:你可以扭蛋了!!使用`/扭蛋`進行扭蛋"},
		{name: "null", text: "null", wantCost: "代幣數到了null", wantFooter: "你還差:你可以扭蛋了!!使用`/扭蛋`進行扭蛋"},
		{name: "decimal", text: "700.5", wantCost: "代幣數到了700.5", wantFooter: "你還差:你還差575.5就可以扭蛋了，加油!!"},
		{name: "infinity", text: "Infinity", wantCost: "代幣數到了Infinity", wantFooter: "你還差:你還差Infinity就可以扭蛋了，加油!!"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := fakemongo.NewEconomyRepository()
			repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 125})
			repo.PutConfig(domain.EconomyConfig{GuildID: "guild-1", GachaCostText: test.text})
			module := NewModule(repo, nil, nil)
			responder := fakediscord.NewResponder()
			if err := module.CoinQueryHandler()(context.Background(), fakediscord.SlashInteraction("代幣查詢"), responder); err != nil {
				t.Fatalf("handle coin query: %v", err)
			}
			embed := responder.Edits[0].Embeds[0]
			if !strings.Contains(embed.Description, test.wantCost) || embed.Footer == nil || embed.Footer.Text != test.wantFooter {
				t.Fatalf("embed = %#v", embed)
			}
		})
	}
}

func TestCoinQueryModuleRegistersRoute(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 1})
	module := NewModule(repo, nil, nil)
	router := interactions.NewRouter()
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	responder := fakediscord.NewResponder()
	if err := router.Handle(context.Background(), fakediscord.SlashInteraction("代幣查詢"), responder); err != nil {
		t.Fatalf("route coin query: %v", err)
	}
	if len(responder.Edits) != 1 {
		t.Fatalf("expected routed edit, got %#v", responder.Edits)
	}
}
