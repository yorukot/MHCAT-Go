package onboarding

import (
	"reflect"
	"testing"

	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/onboarding"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

func TestAccountAgeDefinitionMatchesLegacyPayload(t *testing.T) {
	want := commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        AccountAgeCommandName,
		Description: "設定用戶要創建多久才能加入這個伺服器",
		Ownership:   commands.ManagedOwnership("account-age-config", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type: commands.OptionTypeSubCommand, Name: "小時數", Description: "設定用戶需要滿幾小時才能夠進入伺服器",
				Options: []commands.Option{{Type: commands.OptionTypeInteger, Name: "小時數", Description: "輸入當未滿幾個小時時要自動踢出!", Required: true}},
			},
			{
				Type: commands.OptionTypeSubCommand, Name: "被踢出資訊頻道", Description: "當有人因為未滿創建時數被踢出時要在哪裡發送",
				Options: []commands.Option{{Type: commands.OptionTypeChannel, Name: "頻道", Description: "設定因未達創建時數而被踢出的使用者資訊!", Required: true, ChannelTypes: []int{0, 5}}},
			},
			{Type: commands.OptionTypeSubCommand, Name: "創建時數刪除", Description: "刪除之前設定的小時數以及被踢出後再發的頻道"},
			{Type: commands.OptionTypeSubCommand, Name: "被踢出資訊頻道刪除", Description: "刪除之前的設定發送頻道"},
		},
	}
	if got := AccountAgeDefinition(); !reflect.DeepEqual(got, want) {
		t.Fatalf("definition = %#v, want %#v", got, want)
	}
}

func TestAccountAgeCommandMessagesPreserveLegacyPayloads(t *testing.T) {
	tests := []struct {
		name string
		got  responses.Message
		want responses.Message
	}{
		{
			name: "hours success",
			got:  accountAgeHoursSuccessMessage(6),
			want: responses.Message{
				Embeds: []responses.Embed{{
					Title:       "<a:green_tick:994529015652163614>群組防護系統",
					Description: "已為您設定必須創建帳號0.3天才能加入伺服器",
					Color:       accountAgeSuccessColor,
				}},
				AllowedMentions: &responses.AllowedMentions{},
			},
		},
		{
			name: "channel success",
			got:  accountAgeChannelSuccessMessage("channel-1"),
			want: responses.Message{
				Embeds: []responses.Embed{{
					Title:       "<a:green_tick:994529015652163614>群組防護系統",
					Description: "已為您設定當未達創建時數時會在:\n<#channel-1>發送使用者資運",
					Color:       accountAgeSuccessColor,
				}},
				AllowedMentions: &responses.AllowedMentions{},
			},
		},
		{
			name: "delete success",
			got:  accountAgeDeleteSuccessMessage("已刪除帳號需創建時數所有設定"),
			want: responses.Message{
				Embeds: []responses.Embed{{
					Title:       "<:trashbin:995991389043163257>群組防護系統",
					Description: "已刪除帳號需創建時數所有設定",
					Color:       accountAgeSuccessColor,
				}},
				AllowedMentions: &responses.AllowedMentions{},
			},
		},
		{
			name: "channel delete success",
			got:  accountAgeDeleteSuccessMessage("已刪除被踢出資訊頻道"),
			want: responses.Message{
				Embeds:          []responses.Embed{{Title: "<:trashbin:995991389043163257>群組防護系統", Description: "已刪除被踢出資訊頻道", Color: accountAgeSuccessColor}},
				AllowedMentions: &responses.AllowedMentions{},
			},
		},
		{
			name: "permission error",
			got:  accountAgeErrorMessage("你需要有`踢出用戶`才能使用此指令"),
			want: responses.Message{
				Embeds: []responses.Embed{{
					Title: "<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`踢出用戶`才能使用此指令",
					Color: joinRoleErrorColor,
				}},
				AllowedMentions: &responses.AllowedMentions{},
			},
		},
		{
			name: "positive hours error",
			got:  accountAgeErrorMessage("不可為負數或0!!!"),
			want: responses.Message{
				Embeds:          []responses.Embed{{Title: "<a:Discord_AnimatedNo:1015989839809757295> | 不可為負數或0!!!", Color: joinRoleErrorColor}},
				AllowedMentions: &responses.AllowedMentions{},
			},
		},
		{
			name: "missing requirement error",
			got:  accountAgeErrorMessage("你必須先設定`/帳號需創建時數 小時數`"),
			want: responses.Message{
				Embeds:          []responses.Embed{{Title: "<a:Discord_AnimatedNo:1015989839809757295> | 你必須先設定`/帳號需創建時數 小時數`", Color: joinRoleErrorColor}},
				AllowedMentions: &responses.AllowedMentions{},
			},
		},
		{
			name: "missing config error",
			got:  accountAgeErrorMessage("你還沒設定過喔!"),
			want: responses.Message{
				Embeds:          []responses.Embed{{Title: "<a:Discord_AnimatedNo:1015989839809757295> | 你還沒設定過喔!", Color: joinRoleErrorColor}},
				AllowedMentions: &responses.AllowedMentions{},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !reflect.DeepEqual(tc.got, tc.want) {
				t.Fatalf("message = %#v, want %#v", tc.got, tc.want)
			}
		})
	}
}

func TestAccountAgeJavaScriptRoundingBoundaries(t *testing.T) {
	for _, tc := range []struct {
		hours int64
		want  string
	}{
		{hours: 1, want: "0"},
		{hours: 2, want: "0.1"},
		{hours: 6, want: "0.3"},
		{hours: 12, want: "0.5"},
		{hours: 24, want: "1"},
		{hours: 26, want: "1.1"},
	} {
		if got := coreservice.AccountAgeRoundedDays(tc.hours); got != tc.want {
			t.Fatalf("hours %d = %q, want %q", tc.hours, got, tc.want)
		}
	}
}
