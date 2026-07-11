package balance

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
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
	if !reflect.DeepEqual(responder.Defers, []responses.DeferOptions{{Ephemeral: true}}) {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	want := responses.Message{
		Embeds: []responses.Embed{{
			Author: &responses.EmbedAuthor{
				Name:    "伺服器目前剩於餘額: 88",
				IconURL: "https://media.discordapp.net/attachments/991337796960784424/1078883215462383697/success.gif",
			},
			Color: 0x57F287,
		}},
		Ephemeral:       true,
		AllowedMentions: &responses.AllowedMentions{},
	}
	if len(responder.Edits) != 1 || !reflect.DeepEqual(responder.Edits[0], want) {
		t.Fatalf("edits = %#v, want %#v", responder.Edits, want)
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

func TestHandlerReturnsControlledRepositoryError(t *testing.T) {
	repo := fakemongo.NewBalanceRepository()
	repo.Err = errors.New("mongo credential secret")
	module := NewModule(repo, nil)
	responder := fakediscord.NewResponder()

	if err := module.Handler()(context.Background(), fakediscord.SlashInteraction(CommandName), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	want := responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | 很抱歉，出現了未知的錯誤，請重試!",
			Color: 0xED4245,
		}},
		Ephemeral:       true,
		AllowedMentions: &responses.AllowedMentions{},
	}
	if len(responder.Edits) != 1 || !reflect.DeepEqual(responder.Edits[0], want) || strings.Contains(responder.Edits[0].Embeds[0].Title, "credential") {
		t.Fatalf("edits = %#v, want %#v", responder.Edits, want)
	}
}
