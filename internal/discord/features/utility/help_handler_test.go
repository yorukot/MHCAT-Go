package utility_test

import (
	"context"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	featureutility "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/utility"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeinteractions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestHelpHandlerOverviewMatchesLegacyMenu(t *testing.T) {
	module := featureutility.NewModule(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), nil, nil, nil)
	responder := fakediscord.NewResponder()
	if err := module.HelpHandler()(context.Background(), fakediscord.SlashInteraction("help"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 || len(responder.Edits) != 1 {
		t.Fatalf("defers=%d edits=%d", len(responder.Defers), len(responder.Edits))
	}
	msg := responder.Edits[0]
	if msg.Content != "" {
		t.Fatalf("legacy help overview should use embeds, got content %q", msg.Content)
	}
	if len(msg.Embeds) != 1 {
		t.Fatalf("legacy help overview embeds = %d", len(msg.Embeds))
	}
	embed := msg.Embeds[0]
	if embed.Author == nil || embed.Author.Name != "MHCAT" {
		t.Fatalf("legacy help author = %#v", embed.Author)
	}
	if embed.Timestamp.IsZero() || embed.Color < 0 || embed.Color > 0xFFFFFF {
		t.Fatalf("legacy help presentation = %#v", embed)
	}
	if len(embed.Fields) != 0 {
		t.Fatalf("legacy overview did not attach category fields: %#v", embed.Fields)
	}
	if !strings.Contains(embed.Description, "嗨嗨，你發現了酷東西") || !strings.Contains(embed.Description, "/help 指令名稱") {
		t.Fatalf("unexpected legacy help description:\n%s", embed.Description)
	}
	if len(msg.Components) != 2 || len(msg.Components[0].Components) != 1 {
		t.Fatalf("legacy help components = %#v", msg.Components)
	}
	selectMenu := msg.Components[0].Components[0]
	if selectMenu.CustomID != "helphelphelphelpmenu" || selectMenu.Placeholder != "📜 選擇命令類別" {
		t.Fatalf("legacy select menu = %#v", selectMenu)
	}
	if len(selectMenu.Options) == 0 {
		t.Fatal("legacy select menu has no category options")
	}
	if len(msg.Components[1].Components) != 3 {
		t.Fatalf("legacy link buttons = %#v", msg.Components[1].Components)
	}
}

func TestHelpHandlerCommandDetail(t *testing.T) {
	module := featureutility.NewModule(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), nil, nil, nil)
	interaction := fakediscord.SlashInteractionWithOptions("help", "", map[string]string{"指令名稱": "ping"})
	responder := fakediscord.NewResponder()
	if err := module.HelpHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	msg := responder.Edits[0]
	if len(msg.Embeds) != 1 || !strings.Contains(msg.Embeds[0].Title, "指令資料") {
		t.Fatalf("detail embed = %#v", msg.Embeds)
	}
	if msg.Embeds[0].Timestamp.IsZero() || msg.Embeds[0].Color < 0 || msg.Embeds[0].Color > 0xFFFFFF {
		t.Fatalf("detail presentation = %#v", msg.Embeds[0])
	}
	foundName := false
	foundDescription := false
	for _, field := range msg.Embeds[0].Fields {
		if strings.Contains(field.Name, "指令名稱") && strings.Contains(field.Value, "ping") {
			foundName = true
		}
		if strings.Contains(field.Name, "指令說明") && strings.Contains(field.Value, "查看我的ping") {
			foundDescription = true
		}
	}
	if !foundName || !foundDescription {
		t.Fatalf("detail fields = %#v", msg.Embeds[0].Fields)
	}
}

func TestHelpHandlerUnknownCommandSafe(t *testing.T) {
	module := featureutility.NewModule(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), nil, nil, nil)
	interaction := fakediscord.SlashInteractionWithOptions("help", "", map[string]string{"指令名稱": "ticket"})
	responder := fakediscord.NewResponder()
	if err := module.HelpHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	msg := responder.Edits[0]
	if len(msg.Embeds) != 1 || !strings.Contains(msg.Embeds[0].Title, "無效的指令") || strings.Contains(msg.Embeds[0].Title, "help command not found") {
		t.Fatalf("unsafe not-found response: %#v", msg.Embeds)
	}
}

func TestHelpHandlerTracksUsage(t *testing.T) {
	tracker := &fakeusage.Tracker{}
	module := featureutility.NewModule(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), nil, nil, tracker)
	if err := module.HelpHandler()(context.Background(), fakediscord.SlashInteraction("help"), fakediscord.NewResponder()); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(tracker.Events) != 1 || tracker.Events[0].CommandName != "help" {
		t.Fatalf("usage events = %#v", tracker.Events)
	}
}

func TestHelpComponentRoutesByParsedVersionedID(t *testing.T) {
	module := featureutility.NewModule(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), nil, nil, nil)
	router := interactions.NewRouter()
	router.SetCustomIDParser(interactions.DefaultCustomIDParser{})
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	responder := fakediscord.NewResponder()
	if err := router.Handle(context.Background(), fakeinteractions.Component("mhcat:v1:help:category:overview"), responder); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Description, "嗨嗨，你發現了酷東西") {
		t.Fatalf("component response = %#v", responder.Edits)
	}
	if !responder.State.Ephemeral() {
		t.Fatal("help component response should preserve ephemeral defer")
	}
}

func TestHelpComponentRoutesByParsedLegacyID(t *testing.T) {
	module := featureutility.NewModule(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), nil, nil, nil)
	router := interactions.NewRouter()
	router.SetCustomIDParser(interactions.DefaultCustomIDParser{})
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	responder := fakediscord.NewResponder()
	if err := router.Handle(context.Background(), fakeinteractions.ComponentWithValues("helphelphelphelpmenu", "實用工具"), responder); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(responder.Edits) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	msg := responder.Edits[0]
	if !msg.Ephemeral || len(msg.Embeds) != 1 || !strings.Contains(msg.Embeds[0].Title, "實用工具 指令") {
		t.Fatalf("legacy category response = %#v", msg)
	}
	foundPing := false
	for _, field := range msg.Embeds[0].Fields {
		if strings.Contains(field.Name, "ping") && strings.Contains(field.Value, "查看我的ping") {
			foundPing = true
		}
	}
	if !foundPing {
		t.Fatalf("legacy category fields = %#v", msg.Embeds[0].Fields)
	}
}
