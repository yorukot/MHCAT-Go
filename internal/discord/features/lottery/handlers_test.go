package lottery

import (
	"context"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestCreateHandlerReturnsLegacyUnavailableEmbed(t *testing.T) {
	usage := &fakeusage.Tracker{}
	module := NewModule(usage)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction(LotteryCreateCommandName)

	if err := module.CreateHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 || !responder.Defers[0].Ephemeral {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != legacyUnavailableTitle || embed.Color != legacySuccessColor {
		t.Fatalf("embed = %#v", embed)
	}
	if len(responder.Replies) != 0 || len(responder.Follow) != 0 {
		t.Fatalf("unexpected responses replies=%#v followups=%#v", responder.Replies, responder.Follow)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != LotteryCreateCommandName || usage.Events[0].Feature != "lottery-disabled-command" {
		t.Fatalf("usage events = %#v", usage.Events)
	}
}

func TestCreateHandlerDoesNotRequireRuntimeManageMessagesPermission(t *testing.T) {
	module := NewModule(nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction(LotteryCreateCommandName)

	if err := module.CreateHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != legacyUnavailableTitle {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}
