package moderation

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestWarningHistoryDefinitionMatchesLegacyShape(t *testing.T) {
	definition := WarningHistoryDefinition()
	if definition.Name != "警告紀錄" || definition.Description != "收尋一位使用者的警告" {
		t.Fatalf("definition = %#v", definition)
	}
	if len(definition.Options) != 1 {
		t.Fatalf("options = %#v", definition.Options)
	}
	option := definition.Options[0]
	if option.Type != commands.OptionTypeUser || option.Name != "使用者" || option.Description != "要收尋的使用者!" || !option.Required {
		t.Fatalf("option = %#v", option)
	}
	if !commands.IsManagedForScope(definition, commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}) {
		t.Fatal("warning history command should be managed for guild staging")
	}
	if err := commands.ValidateRegistry(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, Definitions())); err != nil {
		t.Fatalf("validate registry: %v", err)
	}
}

func TestWarningSettingsDefinitionMatchesLegacyShape(t *testing.T) {
	definition := WarningSettingsDefinition()
	if definition.Name != "警告設定" || definition.Description != "警告的各種設定" {
		t.Fatalf("definition = %#v", definition)
	}
	if len(definition.Options) != 2 {
		t.Fatalf("options = %#v", definition.Options)
	}
	action := definition.Options[0]
	if action.Type != commands.OptionTypeString || action.Name != warningSettingsOptionAction || action.Description != "警告他的原因" || !action.Required {
		t.Fatalf("action option = %#v", action)
	}
	if len(action.Choices) != 2 || action.Choices[0].Name != "停權" || action.Choices[0].Value != "停權" || action.Choices[1].Name != "踢出" || action.Choices[1].Value != "踢出" {
		t.Fatalf("choices = %#v", action.Choices)
	}
	threshold := definition.Options[1]
	if threshold.Type != commands.OptionTypeInteger || threshold.Name != warningSettingsOptionThreshold || threshold.Description != "被警告幾次後要執行這個動作!" || !threshold.Required {
		t.Fatalf("threshold option = %#v", threshold)
	}
	if !commands.IsManagedForScope(definition, commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}) {
		t.Fatal("warning settings command should be managed for guild staging")
	}
	if err := commands.ValidateRegistry(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, SettingsDefinitions())); err != nil {
		t.Fatalf("validate registry: %v", err)
	}
}

func TestWarningIssueDefinitionMatchesLegacyShape(t *testing.T) {
	definition := WarningIssueDefinition()
	if definition.Name != "警告" || definition.Description != "警告一個使用者" {
		t.Fatalf("definition = %#v", definition)
	}
	if len(definition.Options) != 2 {
		t.Fatalf("options = %#v", definition.Options)
	}
	if definition.Options[0].Type != commands.OptionTypeUser || definition.Options[0].Name != warningOptionUser || definition.Options[0].Description != "要警告的使用者!" || !definition.Options[0].Required {
		t.Fatalf("user option = %#v", definition.Options[0])
	}
	if definition.Options[1].Type != commands.OptionTypeString || definition.Options[1].Name != warningIssueOptionReason || definition.Options[1].Description != "警告他的原因" || !definition.Options[1].Required {
		t.Fatalf("reason option = %#v", definition.Options[1])
	}
	if !commands.IsManagedForScope(definition, commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}) {
		t.Fatal("warning issue command should be managed for guild staging")
	}
	if err := commands.ValidateRegistry(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, IssueDefinitions())); err != nil {
		t.Fatalf("validate registry: %v", err)
	}
}

func TestWarningRemovalDefinitionsMatchLegacyShape(t *testing.T) {
	remove := WarningRemoveDefinition()
	if remove.Name != "警告清除" || remove.Description != "清除一個使用者的某個警告" {
		t.Fatalf("remove definition = %#v", remove)
	}
	if len(remove.Options) != 2 {
		t.Fatalf("remove options = %#v", remove.Options)
	}
	if remove.Options[0].Type != commands.OptionTypeUser || remove.Options[0].Name != warningOptionUser || remove.Options[0].Description != "要清除資料的使用者!" || !remove.Options[0].Required {
		t.Fatalf("remove user option = %#v", remove.Options[0])
	}
	if remove.Options[1].Type != commands.OptionTypeInteger || remove.Options[1].Name != warningRemoveOptionIndex || remove.Options[1].Description != "要清除第幾個警告!" || !remove.Options[1].Required {
		t.Fatalf("remove index option = %#v", remove.Options[1])
	}
	removeAll := WarningRemoveAllDefinition()
	if removeAll.Name != "警告全部清除" || removeAll.Description != "清除一個使用者的全部警告" {
		t.Fatalf("remove all definition = %#v", removeAll)
	}
	if len(removeAll.Options) != 1 || removeAll.Options[0].Type != commands.OptionTypeUser || removeAll.Options[0].Name != warningOptionUser || removeAll.Options[0].Description != "要清除資料的使用者!" || !removeAll.Options[0].Required {
		t.Fatalf("remove all options = %#v", removeAll.Options)
	}
	if !commands.IsManagedForScope(remove, commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}) || !commands.IsManagedForScope(removeAll, commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}) {
		t.Fatal("warning removal commands should be managed for guild staging")
	}
	if err := commands.ValidateRegistry(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, RemovalDefinitions())); err != nil {
		t.Fatalf("validate registry: %v", err)
	}
}

func TestCleanupDefinitionMatchesLegacyShape(t *testing.T) {
	definition := CleanupDefinition()
	if definition.Name != "刪除訊息" || definition.Description != "刪除大量訊息" {
		t.Fatalf("definition = %#v", definition)
	}
	if len(definition.Options) != 2 {
		t.Fatalf("options = %#v", definition.Options)
	}
	count := definition.Options[0]
	if count.Type != commands.OptionTypeInteger || count.Name != cleanupOptionCount || count.Description != "設定要刪除幾個訊息(最高1000超過200需要管理者權限)(只能刪除14天內的消息)" || !count.Required {
		t.Fatalf("count option = %#v", count)
	}
	user := definition.Options[1]
	if user.Type != commands.OptionTypeUser || user.Name != cleanupOptionUser || user.Description != "選擇是否要刪除某個特定的使用者的訊息(如填選這項，第一項代表的將是檢測訊息數量)" || user.Required {
		t.Fatalf("user option = %#v", user)
	}
	if !commands.IsManagedForScope(definition, commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}) {
		t.Fatal("cleanup command should be managed for guild staging")
	}
	if err := commands.ValidateRegistry(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, CleanupDefinitions())); err != nil {
		t.Fatalf("validate registry: %v", err)
	}
}

func TestDeleteDataDefinitionMatchesLegacyShape(t *testing.T) {
	definition := DeleteDataDefinition()
	if definition.Name != "刪除資料" || definition.Description != "刪除之前設置過的資料" {
		t.Fatalf("definition = %#v", definition)
	}
	if len(definition.Options) != 0 {
		t.Fatalf("options = %#v", definition.Options)
	}
	if definition.DefaultMemberPermissions != nil {
		t.Fatalf("legacy command is publicly discoverable: %#v", definition.DefaultMemberPermissions)
	}
	if !commands.IsManagedForScope(definition, commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}) {
		t.Fatal("delete-data command should be managed for guild staging")
	}
	if err := commands.ValidateRegistry(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, DeleteDataDefinitions())); err != nil {
		t.Fatalf("validate registry: %v", err)
	}
}
