package stats

import (
	"context"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestQueryHandlerRendersLegacyStaticEmbed(t *testing.T) {
	usage := &fakeusage.Tracker{}
	module := NewModuleWithColor(usage, func() int { return 0x123456 })
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction(StatsQueryCommandName)

	if err := module.QueryHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Replies) != 1 {
		t.Fatalf("replies = %#v", responder.Replies)
	}
	reply := responder.Replies[0]
	if len(reply.Embeds) != 1 {
		t.Fatalf("embeds = %#v", reply.Embeds)
	}
	embed := reply.Embeds[0]
	if embed.Title != "統計系統查詢" || embed.Color != 0x123456 {
		t.Fatalf("embed = %#v", embed)
	}
	for _, want := range []string{
		"我的統計系統是每**10分鐘更新一次**",
		"輸入 /統計系統創建",
		"用戶總數 (伺服器的總人數)",
		"文字頻道數量 (文字頻道總數)",
		"語音頻道數量 (語音頻道總數)",
	} {
		if !strings.Contains(embed.Description, want) {
			t.Fatalf("description missing %q: %q", want, embed.Description)
		}
	}
	if reply.AllowedMentions == nil {
		t.Fatal("expected allowed mentions to be disabled explicitly")
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != StatsQueryCommandName || usage.Events[0].Feature != "stats-query" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestQueryHandlerDoesNotDeferLikeLegacy(t *testing.T) {
	module := NewModule(nil)
	responder := fakediscord.NewResponder()
	if err := module.QueryHandler()(context.Background(), fakediscord.SlashInteraction(StatsQueryCommandName), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 0 || len(responder.Edits) != 0 {
		t.Fatalf("unexpected defer/edit: defers=%#v edits=%#v", responder.Defers, responder.Edits)
	}
}
