package onboarding

import (
	"reflect"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !reflect.DeepEqual(tc.got, tc.want) {
				t.Fatalf("message = %#v, want %#v", tc.got, tc.want)
			}
		})
	}
}
