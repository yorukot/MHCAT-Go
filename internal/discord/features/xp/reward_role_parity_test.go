package xp

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestRewardRoleSuccessMessagesMatchLegacy(t *testing.T) {
	for _, tc := range []struct {
		name  string
		label string
	}{
		{name: "text", label: "聊天"},
		{name: "voice", label: "語音"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			wantSave := responses.Message{
				Embeds: []responses.Embed{{
					Title:       "<:Channel:994524759289233438>" + tc.label + "經驗系統",
					Description: "<a:green_tick:994529015652163614>成功`增加`/`修改`該設定",
					Color:       0x53FF53,
				}},
				AllowedMentions: &responses.AllowedMentions{},
			}
			if got := rewardRoleSaveMessage(tc.label); !reflect.DeepEqual(got, wantSave) {
				t.Fatalf("save message = %#v, want %#v", got, wantSave)
			}

			wantDelete := responses.Message{
				Embeds: []responses.Embed{{
					Title:       "<:trashbin:995991389043163257>" + tc.label + "經驗系統",
					Description: "<a:green_tick:994529015652163614>成功刪除該設定",
					Color:       0x53FF53,
				}},
				AllowedMentions: &responses.AllowedMentions{},
			}
			if got := rewardRoleDeleteMessage(tc.label); !reflect.DeepEqual(got, wantDelete) {
				t.Fatalf("delete message = %#v, want %#v", got, wantDelete)
			}
		})
	}
}

func TestRewardRoleErrorMessagesMatchLegacy(t *testing.T) {
	for _, tc := range []struct {
		name    string
		err     error
		content string
	}{
		{name: "hierarchy", err: ports.ErrDiscordRoleNotAssignable, content: "我沒有權限給大家這個身分組(請把我的身分組調高)!"},
		{name: "limit", err: ports.ErrXPRewardRoleLimitExceeded, content: "你的設定已經過多，請先刪除一些!"},
		{name: "missing text delete", err: ports.ErrTextXPRewardRoleMissing, content: "你沒有設定過這個選項!"},
		{name: "missing voice delete", err: ports.ErrVoiceXPRewardRoleMissing, content: "你沒有設定過這個選項!"},
		{name: "unknown", err: errors.New("database unavailable"), content: "很抱歉，出現了未知的錯誤，請重試!"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			want := responses.Message{
				Embeds: []responses.Embed{{
					Title: "<a:Discord_AnimatedNo:1015989839809757295> | " + tc.content,
					Color: 0xED4245,
				}},
				AllowedMentions: &responses.AllowedMentions{},
			}
			if got := rewardRoleErrorMessage(tc.err); !reflect.DeepEqual(got, want) {
				t.Fatalf("error message = %#v, want %#v", got, want)
			}
		})
	}
}

func TestRewardRoleQueryMessagesMatchLegacy(t *testing.T) {
	configs := []domain.XPRewardRoleConfig{
		{GuildID: "guild-1", Level: 3, RoleID: "role-1"},
		{GuildID: "guild-1", Level: 8, RoleID: "role-2", DeleteWhenNot: true},
	}
	for _, tc := range []struct {
		name       string
		label      string
		text       bool
		kind       string
		embedTitle string
	}{
		{name: "text", label: "聊天", text: true, kind: "text", embedTitle: "<:Channel:994524759289233438> 以下是聊天經驗身分組的所有設定!!"},
		{name: "voice", label: "語音", text: false, kind: "voice", embedTitle: "<:Channel:994524759289233438> 以下是語音經驗身分組的所有設定!!"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			want := responses.Message{
				Embeds: []responses.Embed{{
					Title: tc.embedTitle,
					Color: 0x123456,
					Fields: []responses.EmbedField{
						{Name: "第1個:", Value: "<:levelup:990254382845157406> **等級:**`3`\n<:roleplaying:985945121264635964> **身分組:**<@&role-1>\n<:trashbin:995991389043163257> **是否自動刪除身分組:**false", Inline: true},
						{Name: "第2個:", Value: "<:levelup:990254382845157406> **等級:**`8`\n<:roleplaying:985945121264635964> **身分組:**<@&role-2>\n<:trashbin:995991389043163257> **是否自動刪除身分組:**true", Inline: true},
					},
					Footer: &responses.EmbedFooter{Text: "總共: 2 筆資料\n第 1 / 1 頁(按按鈕會自動更新喔!"},
				}},
				Components: []responses.ComponentRow{{
					Components: []responses.Component{
						{Type: responses.ComponentTypeButton, CustomID: "-1" + tc.kind + "_leave_role", Emoji: "<:previous:986067803910066256>", Label: "上一頁", Style: responses.ButtonStyleSuccess, Disabled: true},
						{Type: responses.ComponentTypeButton, CustomID: "1" + tc.kind + "_leave_role", Emoji: "<:next:986067802056167495>", Label: "下一頁", Style: responses.ButtonStyleSuccess, Disabled: true},
					},
				}},
				AllowedMentions: &responses.AllowedMentions{},
			}
			if got := rewardRoleListMessage(tc.label, configs, 0, tc.text, 0x123456); !reflect.DeepEqual(got, want) {
				t.Fatalf("query message = %#v, want %#v", got, want)
			}
		})
	}
}

func TestRewardRoleHandlersMatchLegacyPermissionAndMissingDeleteErrors(t *testing.T) {
	module := NewRewardRoleModule(fakemongo.NewTextXPRewardRoleRepository(), fakemongo.NewVoiceXPRewardRoleRepository(), fakediscord.NewSideEffects(), nil)
	for _, tc := range []struct {
		name        string
		commandName string
		handler     interactions.Handler
	}{
		{name: "text", commandName: TextXPRewardRoleCommandName, handler: module.TextHandler()},
		{name: "voice", commandName: VoiceXPRewardRoleCommandName, handler: module.VoiceHandler()},
	} {
		t.Run(tc.name+" permission", func(t *testing.T) {
			interaction := fakediscord.SlashInteractionWithOptions(tc.commandName, "設定查詢", nil)
			responder := fakediscord.NewResponder()
			if err := tc.handler(context.Background(), interaction, responder); err != nil {
				t.Fatalf("handler: %v", err)
			}
			assertXPConfigPublicDefer(t, responder)
			assertXPConfigEdit(t, responder, "<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`訊息管理`才能使用此指令", "", 0xED4245)
		})

		t.Run(tc.name+" missing delete", func(t *testing.T) {
			interaction := fakediscord.SlashInteractionWithOptions(tc.commandName, "刪除", map[string]string{"等級": "1", "身分組": "role-1"})
			interaction.Actor.PermissionBits = permissionManageMessages
			responder := fakediscord.NewResponder()
			if err := tc.handler(context.Background(), interaction, responder); err != nil {
				t.Fatalf("handler: %v", err)
			}
			assertXPConfigPublicDefer(t, responder)
			assertXPConfigEdit(t, responder, "<a:Discord_AnimatedNo:1015989839809757295> | 你沒有設定過這個選項!", "", 0xED4245)
		})
	}
}

func TestRewardRoleRandomColorUsesDiscordColorRange(t *testing.T) {
	for i := 0; i < 32; i++ {
		if color := randomXPColor(); color < 0 || color > 0xFFFFFF {
			t.Fatalf("random color = %#x", color)
		}
	}
	if got := (RewardRoleModule{}).nextColor(); got != 0x5865F2 {
		t.Fatalf("fallback color = %#x", got)
	}
}
