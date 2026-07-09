package balance

import (
	"context"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestHandlerRendersLegacyEphemeralBalance(t *testing.T) {
	repo := fakemongo.NewBalanceRepository()
	repo.Balances["guild-1"] = domain.Balance{GuildID: "guild-1", Amount: "88"}
	usage := &fakeusage.Tracker{}
	module := NewModule(repo, usage)
	responder := fakediscord.NewResponder()

	if err := module.Handler()(context.Background(), fakediscord.SlashInteraction(CommandName), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 || !responder.Defers[0].Ephemeral {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Author == nil || embed.Author.Name != "伺服器目前剩於餘額: 88" || embed.Author.IconURL != successIconURL || embed.Color != successColor {
		t.Fatalf("embed = %#v", embed)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != CommandName || usage.Events[0].Feature != "balance-query" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestHandlerDefaultsMissingBalanceToZero(t *testing.T) {
	module := NewModule(fakemongo.NewBalanceRepository(), nil)
	responder := fakediscord.NewResponder()

	if err := module.Handler()(context.Background(), fakediscord.SlashInteraction(CommandName), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Author == nil || embed.Author.Name != "伺服器目前剩於餘額: 0" {
		t.Fatalf("embed = %#v", embed)
	}
}
