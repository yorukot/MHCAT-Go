package logging

import (
	"reflect"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

func TestLoggingDefinitionMatchesLegacyVisibleContract(t *testing.T) {
	definition := LoggingConfigDefinition()
	if definition.Type != commands.CommandTypeChatInput || definition.Name != "set-log-channel" || definition.Description != "Set where log messages should send" || definition.DefaultMemberPermissions != nil {
		t.Fatalf("definition = %#v", definition)
	}
	if !reflect.DeepEqual(definition.NameLocalizations, map[string]string{
		"zh-TW": "設置日誌",
		"zh-CN": "设置日志",
		"en-US": "set-log-channel",
		"en-GB": "set-log-channel",
	}) {
		t.Fatalf("name localizations = %#v", definition.NameLocalizations)
	}
	if !reflect.DeepEqual(definition.DescriptionLocalizations, map[string]string{
		"en-US": "Set where log messages should send",
		"en-GB": "Set where log messages should send",
		"zh-TW": "設置日誌訊息要在哪發送",
		"zh-CN": "设置日志讯息要在哪发送",
	}) {
		t.Fatalf("description localizations = %#v", definition.DescriptionLocalizations)
	}
	if len(definition.Options) != 1 {
		t.Fatalf("options = %#v", definition.Options)
	}
	option := definition.Options[0]
	if option.Type != commands.OptionTypeChannel || option.Name != "channel" || option.Description != "Enter log channel!" || !option.Required || !reflect.DeepEqual(option.ChannelTypes, []int{0, 5}) {
		t.Fatalf("channel option = %#v", option)
	}
	if !reflect.DeepEqual(option.NameLocalizations, map[string]string{"en-US": "channel", "en-GB": "channel", "zh-TW": "頻道", "zh-CN": "頻道"}) ||
		!reflect.DeepEqual(option.DescriptionLocalizations, map[string]string{"en-US": "Enter log channel!", "en-GB": "Enter log channel!", "zh-TW": "輸入日誌頻道!", "zh-CN": "输入日志频道"}) {
		t.Fatalf("channel localizations = %#v / %#v", option.NameLocalizations, option.DescriptionLocalizations)
	}
}

func TestLoggingConfigMessagesMatchLegacyVisibleContract(t *testing.T) {
	wantError := responses.Message{
		Embeds:          []responses.Embed{{Title: "<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`訊息管理`才能使用此指令", Color: 0xED4245}},
		AllowedMentions: &responses.AllowedMentions{},
	}
	if got := loggingErrorMessage("你需要有`訊息管理`才能使用此指令"); !reflect.DeepEqual(got, wantError) {
		t.Fatalf("error message = %#v, want %#v", got, wantError)
	}

	expiresAt := time.UnixMilli(1_700_000_600_000)
	message := loggingPromptMessage("channel-1", "user-1", expiresAt, "https://cdn.discordapp.com/avatars/bot/hash.png", []string{"訊息刪除", "頻道更新"})
	if len(message.Embeds) != 1 || message.Embeds[0].Title != "<:logfile:985948561625710663> 日誌系統" || message.Embeds[0].Description != "**請選擇您需要的日誌(未來會更新更多喔)** \n目前的選擇:`訊息刪除`,`頻道更新`" || message.Embeds[0].Color != 0xFFDC35 {
		t.Fatalf("prompt embed = %#v", message.Embeds)
	}
	if footer := message.Embeds[0].Footer; footer == nil || footer.Text != loggingFooterText || footer.IconURL != "https://cdn.discordapp.com/avatars/bot/hash.webp" {
		t.Fatalf("prompt footer = %#v", footer)
	}
	if len(message.Components) != 1 || len(message.Components[0].Components) != 1 {
		t.Fatalf("components = %#v", message.Components)
	}
	menu := message.Components[0].Components[0]
	wantOptions := []responses.SelectOption{
		{Label: "訊息更新", Description: "當訊息編輯時發送日誌", Value: "訊息更新"},
		{Label: "訊息刪除", Description: "當訊息刪除時發送日誌", Value: "訊息刪除"},
		{Label: "頻道更新", Description: "當頻道更新時發送日誌", Value: "頻道更新"},
		{Label: "用戶語音狀態更新", Description: "當用戶離開或加入或是靜音之類的時發送這個通知", Value: "用戶語音更新"},
	}
	if menu.Type != responses.ComponentTypeSelect || menu.Placeholder != "請選擇您需要的日誌" || menu.MinValues != 1 || menu.MaxValues != 4 || !reflect.DeepEqual(menu.Options, wantOptions) {
		t.Fatalf("select menu = %#v", menu)
	}
	channelID, ownerID, deadline, ok := loggingConfigPayload(menu.CustomID)
	if !ok || channelID != "channel-1" || ownerID != "user-1" || !deadline.Equal(expiresAt) {
		t.Fatalf("collector payload = channel:%q owner:%q deadline:%v ok:%v", channelID, ownerID, deadline, ok)
	}
}

func TestLoggingPermissionContractMatchesLegacyOrderAndFormatting(t *testing.T) {
	wantNames := []string{
		"Create Instant Invite", "Kick Members", "Ban Members", "Administrator", "Manage Channels", "Manage Guild", "Add Reactions", "View AuditLog", "Priority Speaker", "Stream", "View Channel", "Send Messages", "Send TTS Messages", "Manage Messages", "Embed Links", "Attach Files", "Read Message History", "Mention Everyone", "Use External Emojis", "View Guild Insights", "Connect", "Speak", "Mute Members", "Deafen Members", "Move Members", "Use VAD", "Change Nickname", "Manage Nicknames", "Manage Roles", "Manage Webhooks", "Manage Emojis And Stickers", "Use Application Commands", "Request To Speak", "Manage Events", "Manage Threads", "Create Public Threads", "Create Private Threads", "Use External Stickers", "Send Messages In Threads", "Use Embedded Activities", "Moderate Members",
	}
	gotNames := make([]string, 0, len(loggingLegacyPermissionOrder))
	for _, permission := range loggingLegacyPermissionOrder {
		gotNames = append(gotNames, permission.Name)
	}
	if !reflect.DeepEqual(gotNames, wantNames) {
		t.Fatalf("permission names = %#v", gotNames)
	}

	diffs := loggingPermissionDiffs(events.ChannelUpdate{
		OldPermissionOverwrites: []events.PermissionOverwrite{{ID: "role-1", Type: 0, Allow: loggingPermissionSendMessages, Deny: loggingPermissionConnect}},
		NewPermissionOverwrites: []events.PermissionOverwrite{{ID: "role-1", Type: 0, Allow: loggingPermissionSendMessages | loggingPermissionManageMessages}},
	})
	if len(diffs) != 1 {
		t.Fatalf("diffs = %#v", diffs)
	}
	wantField := "<:icons_text1:1000814305068986590><@&role-1>\n" +
		"<:YellowSmallDot:1023970607429328946> Connect\n" +
		"<:check:1085240252978966548> Manage Messages\n"
	if got := loggingPermissionFieldValue(diffs[0]); got != wantField {
		t.Fatalf("permission field = %q, want %q", got, wantField)
	}
}
