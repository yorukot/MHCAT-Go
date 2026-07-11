package onboarding

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestJoinRoleDefinitionsMatchLegacySurface(t *testing.T) {
	defs := Definitions()
	if len(defs) != 2 {
		t.Fatalf("defs = %#v", defs)
	}
	set := defs[0]
	if set.Name != JoinRoleSetCommandName || set.Description != "設定玩家加入時要給甚麼身份組" {
		t.Fatalf("set definition = %#v", set)
	}
	if set.DefaultMemberPermissions != nil {
		t.Fatalf("legacy set command is publicly discoverable: %#v", set.DefaultMemberPermissions)
	}
	if len(set.Options) != 2 || set.Options[0].Type != commands.OptionTypeRole || set.Options[0].Name != "身分組" {
		t.Fatalf("set options = %#v", set.Options)
	}
	if got := set.Options[1].Choices[0].Value; got != domain.JoinRoleGiveAllUsers {
		t.Fatalf("choice value = %#v", got)
	}
	del := defs[1]
	if del.Name != JoinRoleDeleteCommandName || del.Description != "刪除之前設定的加入身份組" {
		t.Fatalf("delete definition = %#v", del)
	}
	if del.DefaultMemberPermissions != nil {
		t.Fatalf("legacy delete command is publicly discoverable: %#v", del.DefaultMemberPermissions)
	}
}

func TestMessageDefinitionsMatchLegacySurface(t *testing.T) {
	defs := MessageDefinitions()
	if len(defs) != 2 {
		t.Fatalf("defs = %#v", defs)
	}
	join := defs[0]
	if join.Name != JoinMessageSetCommandName || join.Description != "設定玩家加入時發送甚麼訊息" {
		t.Fatalf("join definition = %#v", join)
	}
	if join.DefaultMemberPermissions != nil {
		t.Fatalf("legacy join command is publicly discoverable: %#v", join.DefaultMemberPermissions)
	}
	if len(join.Options) != 1 || join.Options[0].Type != commands.OptionTypeChannel || join.Options[0].Name != "頻道" || join.Options[0].Description != "輸入加入訊息要在那發送!" {
		t.Fatalf("join options = %#v", join.Options)
	}
	if len(join.Options[0].ChannelTypes) != 2 || join.Options[0].ChannelTypes[0] != 0 || join.Options[0].ChannelTypes[1] != 5 {
		t.Fatalf("join channel types = %#v", join.Options[0].ChannelTypes)
	}
	leave := defs[1]
	if leave.Name != LeaveMessageSetCommandName || leave.Description != "設定玩家退出時發送甚麼訊息" {
		t.Fatalf("leave definition = %#v", leave)
	}
	if leave.DefaultMemberPermissions != nil {
		t.Fatalf("legacy leave command is publicly discoverable: %#v", leave.DefaultMemberPermissions)
	}
	if len(leave.Options) != 1 || leave.Options[0].Description != "輸入加入訊息要在那發送!" {
		t.Fatalf("leave options = %#v", leave.Options)
	}
}

func TestVerificationDefinitionMatchesLegacySurface(t *testing.T) {
	defs := VerificationDefinitions()
	if len(defs) != 1 {
		t.Fatalf("defs = %#v", defs)
	}
	def := defs[0]
	if def.Name != VerificationSetCommandName || def.Description != "設置驗證完成後要給甚麼身份組" {
		t.Fatalf("definition = %#v", def)
	}
	if def.DefaultMemberPermissions != nil {
		t.Fatalf("legacy command is publicly discoverable: %#v", def.DefaultMemberPermissions)
	}
	if len(def.Options) != 2 {
		t.Fatalf("options = %#v", def.Options)
	}
	if def.Options[0].Type != commands.OptionTypeRole || def.Options[0].Name != "身分組" || def.Options[0].Description != "輸入身份組!" || !def.Options[0].Required {
		t.Fatalf("role option = %#v", def.Options[0])
	}
	if def.Options[1].Type != commands.OptionTypeString || def.Options[1].Name != "改名" || def.Options[1].Description != "輸入名稱，{name}代表原本的名稱ex:平名 | {name} 就會變成 平名 | 夜貓" || def.Options[1].Required {
		t.Fatalf("rename option = %#v", def.Options[1])
	}
	if def.Ownership == nil || !def.Ownership.Managed || def.Ownership.SinceWave != "verification-config" {
		t.Fatalf("ownership = %#v", def.Ownership)
	}
}

func TestAccountAgeDefinitionMatchesLegacySurface(t *testing.T) {
	defs := AccountAgeDefinitions()
	if len(defs) != 1 {
		t.Fatalf("defs = %#v", defs)
	}
	def := defs[0]
	if def.Name != AccountAgeCommandName || def.Description != "設定用戶要創建多久才能加入這個伺服器" {
		t.Fatalf("definition = %#v", def)
	}
	if def.DefaultMemberPermissions != nil {
		t.Fatalf("legacy command has no Discord-side default permission gate: %#v", def.DefaultMemberPermissions)
	}
	if len(def.Options) != 4 {
		t.Fatalf("options = %#v", def.Options)
	}
	if def.Options[0].Name != "小時數" || def.Options[0].Options[0].Name != "小時數" || !def.Options[0].Options[0].Required {
		t.Fatalf("hours subcommand = %#v", def.Options[0])
	}
	if def.Options[1].Name != "被踢出資訊頻道" || def.Options[1].Options[0].Name != "頻道" {
		t.Fatalf("channel subcommand = %#v", def.Options[1])
	}
	if len(def.Options[1].Options[0].ChannelTypes) != 2 || def.Options[1].Options[0].ChannelTypes[0] != 0 || def.Options[1].Options[0].ChannelTypes[1] != 5 {
		t.Fatalf("channel types = %#v", def.Options[1].Options[0].ChannelTypes)
	}
	if def.Options[2].Name != "創建時數刪除" || def.Options[3].Name != "被踢出資訊頻道刪除" {
		t.Fatalf("delete subcommands = %#v", def.Options)
	}
	if def.Ownership == nil || !def.Ownership.Managed || def.Ownership.SinceWave != "account-age-config" {
		t.Fatalf("ownership = %#v", def.Ownership)
	}
}
