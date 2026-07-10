package onboarding

import (
	"reflect"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

func TestJoinRoleMessagesPreserveLegacyPayloads(t *testing.T) {
	tests := []struct {
		name string
		got  responses.Message
		want responses.Message
	}{
		{
			name: "set success",
			got:  joinRoleSetSuccessMessage("role-1"),
			want: responses.Message{
				Embeds: []responses.Embed{{
					Title:       "🪂 加入身分組系統",
					Description: "<a:green_tick:994529015652163614> **成功創建加入給身分組!**\n**身分組:** <@role-1>!",
					Color:       joinRoleSuccessColor,
				}},
				AllowedMentions: &responses.AllowedMentions{},
			},
		},
		{
			name: "delete success",
			got:  joinRoleDeleteSuccessMessage("role-1"),
			want: responses.Message{
				Embeds: []responses.Embed{{
					Title:       "🪂 加入身分組系統",
					Description: "<:trashbin:986308183674990592>**成功刪除:**\n身分組: <@role-1>!",
					Color:       joinRoleSuccessColor,
				}},
				AllowedMentions: &responses.AllowedMentions{},
			},
		},
		{
			name: "permission error",
			got:  joinRoleErrorMessage("你需要有`訊息管理`才能使用此指令"),
			want: responses.Message{
				Embeds: []responses.Embed{{
					Title: "<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`訊息管理`才能使用此指令",
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

func TestJoinMessageDashboardPreservesLegacyPayload(t *testing.T) {
	want := responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:announcement:1005035747197337650> | 該指令已經移往控制面板，請前往控制面板進行設定",
			Color: legacyDashboardColor,
		}},
		Components: []responses.ComponentRow{{Components: []responses.Component{{
			Type:  responses.ComponentTypeButton,
			Style: responses.ButtonStyleLink,
			URL:   "https://mhcat.yorukot.meguilds/guild-1/welcome",
			Label: "點我前往儀錶板設定!",
			Emoji: "<a:arrow:986268851786375218>",
		}}}},
		AllowedMentions: &responses.AllowedMentions{},
	}
	if got := joinMessageDashboardMessage("guild-1"); !reflect.DeepEqual(got, want) {
		t.Fatalf("message = %#v, want %#v", got, want)
	}
}

func TestLeaveMessageModalAndPreviewPreserveLegacyPayloads(t *testing.T) {
	config := domain.LeaveMessageConfig{
		GuildID:        "guild-1",
		ChannelID:      "channel-1",
		Color:          "Green",
		Title:          "Bye",
		MessageContent: "Goodbye {MEMBERNAME}",
	}
	wantModal := responses.Modal{
		CustomID: "nal",
		Title:    "退出訊息設置!",
		Rows: []responses.ModalRow{
			{Inputs: []responses.TextInput{{
				CustomID: "leave_msgcolor",
				Label:    "請輸入你的加入訊息要甚麼顏色(要隨機顏色可輸入:Random)",
				Style:    responses.TextInputStyleShort,
				Required: true,
				Value:    "Green",
			}}},
			{Inputs: []responses.TextInput{{
				CustomID: "leave_msgtitle",
				Label:    "請輸入訊息標題",
				Style:    responses.TextInputStyleShort,
				Required: true,
				Value:    "Bye",
			}}},
			{Inputs: []responses.TextInput{{
				CustomID: "leave_msgcontent",
				Label:    "請輸入訊息內文(如要顯示用戶名可輸入: {MEMBERNAME} )",
				Style:    responses.TextInputStyleParagraph,
				Required: true,
				Value:    "Goodbye {MEMBERNAME}",
			}}},
		},
	}
	if got := leaveMessageModal(config); !reflect.DeepEqual(got, wantModal) {
		t.Fatalf("modal = %#v, want %#v", got, wantModal)
	}

	now := time.Date(2026, 7, 10, 1, 2, 3, 0, time.UTC)
	wantPreview := responses.Message{
		Content: "下面為預覽，想修改嗎?再次輸入指令即可修改((MEMBERNAME)在到時候會變正常喔)",
		Embeds: []responses.Embed{{
			Title:       "Bye",
			Description: "Goodbye {MEMBERNAME}",
			Color:       0x57F287,
			Timestamp:   now,
			Thumbnail:   &responses.EmbedImage{URL: "https://example.test/avatar.png"},
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
	if got := leaveMessagePreviewMessage(config, "https://example.test/avatar.png", now); !reflect.DeepEqual(got, wantPreview) {
		t.Fatalf("preview = %#v, want %#v", got, wantPreview)
	}
}
