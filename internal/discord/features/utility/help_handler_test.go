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
	if len(responder.Defers) != 1 || len(responder.Follow) != 1 || len(responder.Edits) != 0 {
		t.Fatalf("defers=%d follow=%d edits=%d", len(responder.Defers), len(responder.Follow), len(responder.Edits))
	}
	msg := responder.Follow[0]
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
	msg := responder.Follow[0]
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

func TestHelpHandlerCommandDetailPreservesLegacyPermissionsAndTutorials(t *testing.T) {
	module := featureutility.NewModule(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), nil, nil, nil)
	tests := []struct {
		command    string
		permission string
		tutorial   string
	}{
		{command: "代幣商店", permission: "```查詢跟購買大家都可用，剩下皆須訊息管理```", tutorial: "```此指令目前沒有教學```"},
		{command: "automatic-notification", permission: "```訊息管理```", tutorial: "[__**點我立刻前往教學頁面**__](https://youtu.be/D43zPrZU5Fw)"},
		{command: "公告頻道設置", permission: "```這個指令大家都可以用喔```", tutorial: "[__**點我立刻前往教學頁面**__](https://docsmhcat.yorukot.meocs/ann_set)"},
		{command: "ping", permission: "```這個指令大家都可以用喔```", tutorial: "```此指令目前沒有教學```"},
	}
	for _, test := range tests {
		t.Run(test.command, func(t *testing.T) {
			interaction := fakediscord.SlashInteractionWithOptions("help", "", map[string]string{"指令名稱": test.command})
			responder := fakediscord.NewResponder()
			if err := module.HelpHandler()(context.Background(), interaction, responder); err != nil {
				t.Fatalf("handler: %v", err)
			}
			if len(responder.Follow) != 1 || len(responder.Follow[0].Embeds) != 1 {
				t.Fatalf("response = %#v", responder.Follow)
			}
			fields := responder.Follow[0].Embeds[0].Fields
			if len(fields) != 4 || fields[2].Value != test.permission || fields[3].Value != test.tutorial {
				t.Fatalf("detail fields = %#v", fields)
			}
		})
	}
}

func TestHelpHandlerUnknownCommandSafe(t *testing.T) {
	module := featureutility.NewModule(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), nil, nil, nil)
	interaction := fakediscord.SlashInteractionWithOptions("help", "", map[string]string{"指令名稱": "ticket"})
	responder := fakediscord.NewResponder()
	if err := module.HelpHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	msg := responder.Follow[0]
	if len(msg.Embeds) != 1 || !strings.Contains(msg.Embeds[0].Title, "無效的指令") || strings.Contains(msg.Embeds[0].Title, "help command not found") {
		t.Fatalf("unsafe not-found response: %#v", msg.Embeds)
	}
}

func TestHelpHandlerSlashCategoryMatchesLegacyShape(t *testing.T) {
	module := featureutility.NewModule(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), nil, nil, nil)
	interaction := fakediscord.SlashInteractionWithOptions("help", "", map[string]string{"指令名稱": "實用工具 extra words"})
	responder := fakediscord.NewResponder()
	if err := module.HelpHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	msg := responder.Follow[0]
	if msg.Ephemeral || len(msg.Embeds) != 1 || msg.Embeds[0].Title != "__實用工具 指令!__" {
		t.Fatalf("slash category message = %#v", msg)
	}
	foundPing := false
	for _, field := range msg.Embeds[0].Fields {
		if strings.Contains(field.Name, "`ping`") {
			foundPing = true
			if !strings.HasPrefix(field.Name, "/<:icons_goodping:1084881470075703367>") || field.Value != "查看我的ping" {
				t.Fatalf("slash category ping field = %#v", field)
			}
		}
	}
	if !foundPing {
		t.Fatalf("slash category fields = %#v", msg.Embeds[0].Fields)
	}
}

func TestHelpHandlerPreservesLiteralSpaceQuerySplit(t *testing.T) {
	module := featureutility.NewModule(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), nil, nil, nil)
	for _, query := range []string{" ping", "\tping"} {
		responder := fakediscord.NewResponder()
		interaction := fakediscord.SlashInteractionWithOptions("help", "", map[string]string{"指令名稱": query})
		if err := module.HelpHandler()(context.Background(), interaction, responder); err != nil {
			t.Fatalf("handler for %q: %v", query, err)
		}
		if len(responder.Follow) != 1 || len(responder.Follow[0].Embeds) != 1 || !strings.Contains(responder.Follow[0].Embeds[0].Title, "無效的指令") {
			t.Fatalf("query %q response = %#v", query, responder.Follow)
		}
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
	interaction := fakeinteractions.ComponentWithValues("helphelphelphelpmenu", "實用工具")
	interaction.GuildLocale = "zh-TW"
	if err := router.Handle(context.Background(), interaction, responder); err != nil {
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
		if field.Name == "<:icons_goodping:1084881470075703367> </ping:964185876559196181>" && field.Value == "```fix\n查看我的ping```" {
			foundPing = true
		}
	}
	if !foundPing {
		t.Fatalf("legacy category fields = %#v", msg.Embeds[0].Fields)
	}
}

func TestHelpLegacyCategoryLocalizesCommandsAndExpandsSubcommands(t *testing.T) {
	module := featureutility.NewModule(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), nil, nil, nil)
	router := interactions.NewRouter()
	router.SetCustomIDParser(interactions.DefaultCustomIDParser{})
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	interaction := fakeinteractions.ComponentWithValues("helphelphelphelpmenu", "代幣系統")
	interaction.GuildLocale = "zh-TW"
	responder := fakediscord.NewResponder()
	if err := router.Handle(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handle: %v", err)
	}
	fields := responder.Edits[0].Embeds[0].Fields
	foundLocalized := false
	foundSubcommand := false
	for _, field := range fields {
		if field.Name == "<:coins:997374177944281190> </代幣相關設定:964185876559196181>" && strings.Contains(field.Value, "有關代幣的各項設定") {
			foundLocalized = true
		}
		if field.Name == "<:store:1001118704651743372> </代幣商店 商品增加:964185876559196181>" && strings.HasPrefix(field.Value, "[前往文檔](https://docsmhcat.yorukot.me/allcommands/") {
			foundSubcommand = true
		}
	}
	if !foundLocalized || !foundSubcommand {
		t.Fatalf("localized=%t subcommand=%t fields=%#v", foundLocalized, foundSubcommand, fields)
	}
}
