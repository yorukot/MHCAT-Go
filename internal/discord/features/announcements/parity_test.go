package announcements

import (
	"errors"
	"reflect"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

func TestAnnouncementDefinitionsMatchLegacyVisibleContract(t *testing.T) {
	channel := func() commands.Option {
		return commands.Option{
			Type:         commands.OptionTypeChannel,
			Name:         "頻道",
			Description:  "輸入公告發送的頻道!",
			Required:     true,
			ChannelTypes: []int{0, 5},
		}
	}
	wantConfig := commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        "公告頻道設置",
		Description: "設定公告在哪發送",
		DocsURL:     "https://docsmhcat.yorukot.meocs/ann_set",
		Ownership:   commands.ManagedOwnership("announcement-config", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:        commands.OptionTypeSubCommand,
				Name:        "一次性公告頻道",
				Description: "設定一次性公告頻道要在哪發送",
				Options:     []commands.Option{channel()},
			},
			{
				Type:        commands.OptionTypeSubCommand,
				Name:        "綁定公告頻道",
				Description: "設定綁定型公告要在哪發送以及發送時的格式",
				Options: []commands.Option{
					channel(),
					{Type: commands.OptionTypeString, Name: "標註", Description: "輸入要標註哪個身分組!", Required: true},
					{Type: commands.OptionTypeString, Name: "顏色", Description: "輸入這個綁定公告頻道的設定!(隨機顏色請輸入Random)", Required: true},
					{Type: commands.OptionTypeString, Name: "標題", Description: "輸入公告發送的頻道!", Required: true},
				},
			},
			{
				Type:        commands.OptionTypeSubCommand,
				Name:        "綁定公告頻道刪除",
				Description: "刪除之前的設定",
				Options:     []commands.Option{channel()},
			},
		},
	}
	if got := ConfigDefinition(); !reflect.DeepEqual(got, wantConfig) {
		t.Fatalf("config definition = %#v, want %#v", got, wantConfig)
	}

	wantSend := commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        "公告發送",
		Description: "發送公告訊息",
		DocsURL:     "https://docsmhcat.yorukot.me/docs/ann",
		Ownership:   commands.ManagedOwnership("announcement-send", commands.ScopeGuild),
	}
	if got := SendDefinition(); !reflect.DeepEqual(got, wantSend) {
		t.Fatalf("send definition = %#v, want %#v", got, wantSend)
	}
}

func TestAnnouncementConfigMessagesMatchAuditedLegacyContract(t *testing.T) {
	emptyMentions := &responses.AllowedMentions{}
	tests := []struct {
		name string
		got  responses.Message
		want responses.Message
	}{
		{
			name: "once create",
			got:  onceSuccessMessage("channel-1", true),
			want: responses.Message{
				Embeds: []responses.Embed{{
					Title:       "<:megaphone:985943890148327454> 公告系統",
					Description: "<:Channel:994524759289233438> **您的公告頻道成功__創建__!!**\n**您目前的公告頻道為**:<#channel-1>",
					Color:       0x53FF53,
				}},
				AllowedMentions: emptyMentions,
			},
		},
		{
			name: "once update",
			got:  onceSuccessMessage("channel-1", false),
			want: responses.Message{
				Embeds: []responses.Embed{{
					Title:       "<:megaphone:985943890148327454> 公告系統",
					Description: "<:Channel:994524759289233438> **您的公告頻道成功__更新__!!**\n**您目前的公告頻道為**:<#channel-1>",
					Color:       0x53FF53,
				}},
				AllowedMentions: emptyMentions,
			},
		},
		{
			name: "bound create",
			got:  boundSuccessMessage("channel-1", true),
			want: responses.Message{
				Embeds: []responses.Embed{{
					Title:       "<:megaphone:985943890148327454> 綁定型公告系統",
					Description: "<:Channel:994524759289233438> **您的綁定型公告頻道成功__創建__!!**\n**新增綁定型公告頻道為**:<#channel-1>",
					Color:       0x53FF53,
				}},
				AllowedMentions: emptyMentions,
			},
		},
		{
			name: "bound update",
			got:  boundSuccessMessage("channel-1", false),
			want: responses.Message{
				Embeds: []responses.Embed{{
					Title:       "<:megaphone:985943890148327454> 綁定型公告系統",
					Description: "<:Channel:994524759289233438> **您的綁定型公告頻道成功__更新__!!**\n**新增綁定型公告頻道為**:<#channel-1>",
					Color:       0x53FF53,
				}},
				AllowedMentions: emptyMentions,
			},
		},
		{
			name: "bound delete",
			got:  boundDeleteSuccessMessage("channel-1"),
			want: responses.Message{
				Embeds: []responses.Embed{{
					Title:       "<:megaphone:985943890148327454> 綁定型公告系統",
					Description: "<:trashbin:995991389043163257> **您的綁定型公告頻道成功__刪除__!!**\n**刪除的綁定型公告頻道為**:<#channel-1>",
					Color:       0x53FF53,
				}},
				AllowedMentions: emptyMentions,
			},
		},
		{
			name: "config error",
			got:  announcementErrorMessage("你需要有`undefined`才能使用此指令"),
			want: responses.Message{
				Embeds:          []responses.Embed{{Title: "<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`undefined`才能使用此指令", Color: 0xED4245}},
				AllowedMentions: emptyMentions,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !reflect.DeepEqual(test.got, test.want) {
				t.Fatalf("message = %#v, want %#v", test.got, test.want)
			}
		})
	}
}

func TestAnnouncementModalPreviewAndConfirmationMatchAuditedContract(t *testing.T) {
	wantModal := responses.Modal{
		CustomID: "modal-id",
		Title:    "公告系統",
		Rows: []responses.ModalRow{
			{Inputs: []responses.TextInput{{CustomID: "anntag", Label: "請輸入你要tag誰", Style: responses.TextInputStyleShort, Required: true}}},
			{Inputs: []responses.TextInput{{CustomID: "anncolor", Label: "請輸入你的公告要甚麼顏色", Style: responses.TextInputStyleShort, Required: true}}},
			{Inputs: []responses.TextInput{{CustomID: "anntitle", Label: "請輸入你的公告標題", Style: responses.TextInputStyleShort, Required: true}}},
			{Inputs: []responses.TextInput{{CustomID: "anncontent", Label: "請輸入公告內文", Style: responses.TextInputStyleParagraph, Required: true}}},
		},
	}
	if got := announcementSendModal("modal-id"); !reflect.DeepEqual(got, wantModal) {
		t.Fatalf("modal = %#v, want %#v", got, wantModal)
	}

	draft := AnnouncementDraft{
		GuildID:   "guild-1",
		UserID:    "user-1",
		UserTag:   "Yoru#1234",
		AvatarURL: "https://example.invalid/avatar.png",
		Tag:       "@everyone",
		Color:     0x53FF53,
		Title:     " 公告 ",
		Content:   " 內容 ",
	}
	wantPreview := responses.Message{
		Content: "@everyone",
		Embeds: []responses.Embed{{
			Title:       " 公告 ",
			Description: " 內容 ",
			Color:       0x53FF53,
			Footer:      &responses.EmbedFooter{Text: "來自Yoru#1234的公告", IconURL: "https://example.invalid/avatar.png"},
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
	if got := previewMessage(draft); !reflect.DeepEqual(got, wantPreview) {
		t.Fatalf("preview = %#v, want %#v", got, wantPreview)
	}

	stateID := "abcdefghijklmnop"
	wantConfirmation := responses.Message{
		Embeds: []responses.Embed{{Title: "是否將此訊息送往公告?(請於六秒內點擊:P)", Color: 0x00FF19}},
		Components: []responses.ComponentRow{{Components: []responses.Component{
			{Type: responses.ComponentTypeButton, CustomID: "mhcat:v1:announcement:confirm:state=" + stateID, Emoji: "✅", Label: "是", Style: responses.ButtonStylePrimary},
			{Type: responses.ComponentTypeButton, CustomID: "mhcat:v1:announcement:cancel:state=" + stateID, Emoji: "❎", Label: "否", Style: responses.ButtonStyleDanger},
		}}},
		AllowedMentions: &responses.AllowedMentions{},
	}
	if got := confirmationMessage(stateID); !reflect.DeepEqual(got, wantConfirmation) {
		t.Fatalf("confirmation = %#v, want %#v", got, wantConfirmation)
	}

	wantOutbound := ports.OutboundMessage{
		Content: "@everyone",
		Embeds: []ports.OutboundEmbed{{
			Title:         " 公告 ",
			Description:   " 內容 ",
			Color:         0x53FF53,
			FooterText:    "來自Yoru#1234的公告",
			FooterIconURL: "https://example.invalid/avatar.png",
		}},
		AllowedMentions: ports.AllowedMentions{},
	}
	if got := outboundAnnouncementMessage(draft); !reflect.DeepEqual(got, wantOutbound) {
		t.Fatalf("outbound = %#v, want %#v", got, wantOutbound)
	}
}

func TestAnnouncementFixedSendResponsesMatchLegacyContract(t *testing.T) {
	emptyMentions := &responses.AllowedMentions{}
	tests := []struct {
		name string
		got  responses.Message
		want responses.Message
	}{
		{name: "success", got: sendSuccessMessage(), want: responses.Message{Content: "<a:green_tick:994529015652163614> | 成功發送!", AllowedMentions: emptyMentions}},
		{name: "missing channel", got: missingAnnouncementChannelMessage(), want: responses.Message{Content: "很抱歉!\n你還沒有對您的公告頻道進行選擇!\n命令:`<> 公告頻道設置 [公告頻道id]`\n有問題歡迎打`<>幫助`", AllowedMentions: emptyMentions}},
		{name: "modal color error", got: modalErrorMessage("你傳送的並不是顏色(色碼)"), want: responses.Message{Embeds: []responses.Embed{{Title: "你傳送的並不是顏色(色碼)", Color: 0xED4245}}, AllowedMentions: emptyMentions}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !reflect.DeepEqual(test.got, test.want) {
				t.Fatalf("message = %#v, want %#v", test.got, test.want)
			}
		})
	}
}

func TestAnnouncementLegacyAndVersionedRouteContract(t *testing.T) {
	legacy, err := customid.ParseModal("nal", []customid.ModalField{{CustomID: "anntag", Value: "@here"}})
	if err != nil {
		t.Fatalf("parse legacy modal: %v", err)
	}
	if !legacy.Legacy || legacy.Feature != announcementFeature || legacy.Action != sendModalAction {
		t.Fatalf("legacy modal = %#v", legacy)
	}

	for _, action := range []string{confirmAction, cancelAction} {
		encoded := announcementStateComponentID(action, "abcdefghijklmnop")
		parsed, err := customid.ParseComponent(encoded)
		if err != nil {
			t.Fatalf("parse %s: %v", action, err)
		}
		if parsed.Legacy || parsed.Feature != announcementFeature || parsed.Action != action || parsed.Payload.StateID != "abcdefghijklmnop" || len(encoded) > customid.MaxCustomIDLength {
			t.Fatalf("parsed %s = %#v id=%q", action, parsed, encoded)
		}
	}

	for _, raw := range []string{"announcement_yes", "announcement_no"} {
		if _, err := customid.ParseComponent(raw); !errors.Is(err, customid.ErrAmbiguousID) {
			t.Fatalf("legacy collision %q = %v", raw, err)
		}
	}
}

func TestAnnouncementRandomColorUsesFullDiscordRange(t *testing.T) {
	for range 256 {
		color := randomLegacyColor()
		if color < 0x000000 || color > 0xFFFFFF {
			t.Fatalf("random color out of range: %#x", color)
		}
	}
}
