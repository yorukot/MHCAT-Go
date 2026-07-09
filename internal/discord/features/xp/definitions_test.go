package xp

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestDefinitionsMatchLegacy(t *testing.T) {
	set := TextXPSetDefinition()
	if set.Name != TextXPSetCommandName || set.Description != "設定聊天經驗通知要在哪發送" {
		t.Fatalf("set definition = %#v", set)
	}
	if len(set.Options) != 3 || set.Options[0].Name != "頻道" || !set.Options[0].Required || len(set.Options[0].ChannelTypes) != 2 {
		t.Fatalf("set options = %#v", set.Options)
	}
	del := TextXPDeleteDefinition()
	if del.Name != TextXPDeleteCommandName || del.Description != "刪除聊天經驗發送訊息設置" {
		t.Fatalf("delete definition = %#v", del)
	}
	for _, definition := range Definitions() {
		if definition.Ownership == nil || definition.Ownership.Owner != commands.OwnerMHCATRefactor || definition.Ownership.SinceWave != "text-xp-config" {
			t.Fatalf("ownership = %#v", definition.Ownership)
		}
	}
	if err := commands.ValidateRegistry(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, Definitions())); err != nil {
		t.Fatalf("validate: %v", err)
	}
}

func TestVoiceDefinitionsMatchLegacy(t *testing.T) {
	set := VoiceXPSetDefinition()
	if set.Name != VoiceXPSetCommandName || set.Description != "設定語音經驗通知要在哪發送" {
		t.Fatalf("voice set definition = %#v", set)
	}
	if len(set.Options) != 4 || set.Options[0].Name != "頻道" || !set.Options[0].Required || len(set.Options[0].ChannelTypes) != 2 {
		t.Fatalf("voice set options = %#v", set.Options)
	}
	if set.Options[3].Name != "背景" || set.Options[3].Description != "輸入玩家查詢的背景(默認為discord色)支援png和jpg(可使用discord的複製連結)最佳大小為931*231" {
		t.Fatalf("voice background option = %#v", set.Options[3])
	}
	del := VoiceXPDeleteDefinition()
	if del.Name != VoiceXPDeleteCommandName || del.Description != "刪除語音發送訊息設置" {
		t.Fatalf("voice delete definition = %#v", del)
	}
	for _, definition := range VoiceDefinitions() {
		if definition.Ownership == nil || definition.Ownership.Owner != commands.OwnerMHCATRefactor || definition.Ownership.SinceWave != "voice-xp-config" {
			t.Fatalf("voice ownership = %#v", definition.Ownership)
		}
	}
	if err := commands.ValidateRegistry(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, VoiceDefinitions())); err != nil {
		t.Fatalf("validate voice: %v", err)
	}
}

func TestDisabledProfileDefinitionsMatchLegacy(t *testing.T) {
	text := TextXPProfileDefinition()
	if text.Name != TextXPProfileCommandName || text.Description != "查詢聊天經驗" {
		t.Fatalf("text profile definition = %#v", text)
	}
	voice := VoiceXPProfileDefinition()
	if voice.Name != VoiceXPProfileCommandName || voice.Description != "查詢語音經驗" {
		t.Fatalf("voice profile definition = %#v", voice)
	}
	for _, definition := range DisabledProfileDefinitions() {
		if definition.DefaultMemberPermissions != nil {
			t.Fatalf("profile command should not set Discord-side permissions: %#v", definition.DefaultMemberPermissions)
		}
		if definition.Ownership == nil || definition.Ownership.Owner != commands.OwnerMHCATRefactor || definition.Ownership.SinceWave != "xp-profile-disabled" {
			t.Fatalf("profile ownership = %#v", definition.Ownership)
		}
		if len(definition.Options) != 1 {
			t.Fatalf("profile options = %#v", definition.Options)
		}
		option := definition.Options[0]
		if option.Type != commands.OptionTypeUser || option.Name != "玩家" || option.Description != "輸入玩家!" || option.Required {
			t.Fatalf("profile user option = %#v", option)
		}
	}
	if err := commands.ValidateRegistry(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, DisabledProfileDefinitions())); err != nil {
		t.Fatalf("validate disabled profiles: %v", err)
	}
}

func TestRewardRoleDefinitionsMatchLegacy(t *testing.T) {
	definitions := RewardRoleDefinitions()
	if len(definitions) != 2 {
		t.Fatalf("definitions = %#v", definitions)
	}
	text := TextXPRewardRoleDefinition()
	if text.Name != TextXPRewardRoleCommandName || text.Description != "設定聊天經驗通知要在哪發送" {
		t.Fatalf("text reward definition = %#v", text)
	}
	voice := VoiceXPRewardRoleDefinition()
	if voice.Name != VoiceXPRewardRoleCommandName || voice.Description != "設定語音經驗通知要在哪發送(兼增加、刪除、設定查詢)" {
		t.Fatalf("voice reward definition = %#v", voice)
	}
	for _, definition := range definitions {
		if definition.DefaultMemberPermissions == nil || *definition.DefaultMemberPermissions != manageMessagesPermission {
			t.Fatalf("permissions = %#v", definition.DefaultMemberPermissions)
		}
		if len(definition.Options) != 3 {
			t.Fatalf("reward options = %#v", definition.Options)
		}
		if definition.Options[0].Name != "增加" || definition.Options[0].Type != commands.OptionTypeSubCommand || len(definition.Options[0].Options) != 3 {
			t.Fatalf("add subcommand = %#v", definition.Options[0])
		}
		if definition.Options[1].Name != "刪除" || definition.Options[1].Type != commands.OptionTypeSubCommand || len(definition.Options[1].Options) != 2 {
			t.Fatalf("delete subcommand = %#v", definition.Options[1])
		}
		if definition.Options[2].Name != "設定查詢" || definition.Options[2].Type != commands.OptionTypeSubCommand {
			t.Fatalf("query subcommand = %#v", definition.Options[2])
		}
	}
	if err := commands.ValidateRegistry(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, definitions)); err != nil {
		t.Fatalf("validate reward roles: %v", err)
	}
}

func TestTextAndVoiceDefinitionsAreSeparateGates(t *testing.T) {
	for _, definition := range TextDefinitions() {
		if definition.Name == VoiceXPSetCommandName || definition.Name == VoiceXPDeleteCommandName {
			t.Fatalf("text definitions leaked voice command: %#v", definition)
		}
	}
	for _, definition := range VoiceDefinitions() {
		if definition.Name == TextXPSetCommandName || definition.Name == TextXPDeleteCommandName {
			t.Fatalf("voice definitions leaked text command: %#v", definition)
		}
	}
	for _, definition := range DisabledProfileDefinitions() {
		switch definition.Name {
		case TextXPSetCommandName, TextXPDeleteCommandName, VoiceXPSetCommandName, VoiceXPDeleteCommandName:
			t.Fatalf("disabled profile definitions leaked config command: %#v", definition)
		}
	}
	for _, definition := range RewardRoleDefinitions() {
		switch definition.Name {
		case TextXPSetCommandName, TextXPDeleteCommandName, VoiceXPSetCommandName, VoiceXPDeleteCommandName, TextXPProfileCommandName, VoiceXPProfileCommandName:
			t.Fatalf("reward role definitions leaked another XP command: %#v", definition)
		}
	}
}
