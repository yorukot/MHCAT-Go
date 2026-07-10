package notifications

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

func TestAutoNotificationDefinitionsMatchLegacyVisibleContract(t *testing.T) {
	setup := SetupDefinition()
	if setup.Type != commands.CommandTypeChatInput || setup.Name != "automatic-notification" || setup.Description != "Set where automatic notification should be send" || setup.DocsURL != "https://youtu.be/D43zPrZU5Fw" || setup.DefaultMemberPermissions != nil {
		t.Fatalf("setup definition = %#v", setup)
	}
	wantNameLocalizations := map[string]string{"en-US": "automatic-notification", "zh-TW": "自動通知", "zh-CN": "自动通知"}
	if !reflect.DeepEqual(setup.NameLocalizations, wantNameLocalizations) {
		t.Fatalf("name localizations = %#v", setup.NameLocalizations)
	}
	// The duplicate zh-TW key in the JavaScript literal overwrote the first value.
	wantDescriptionLocalizations := map[string]string{"en-US": "Set where automatic notification should be send", "zh-TW": "设置自动聊天频道要在哪发送"}
	if !reflect.DeepEqual(setup.DescriptionLocalizations, wantDescriptionLocalizations) {
		t.Fatalf("description localizations = %#v", setup.DescriptionLocalizations)
	}
	if len(setup.Options) != 1 {
		t.Fatalf("setup options = %#v", setup.Options)
	}
	channel := setup.Options[0]
	if channel.Type != commands.OptionTypeChannel || channel.Name != "channel" || channel.Description != "Enter channel to send!" || !channel.Required || !reflect.DeepEqual(channel.ChannelTypes, []int{0, 5}) {
		t.Fatalf("channel option = %#v", channel)
	}
	if !reflect.DeepEqual(channel.NameLocalizations, map[string]string{"en-US": "channel", "zh-TW": "頻道", "zh-CN": "頻道"}) || !reflect.DeepEqual(channel.DescriptionLocalizations, map[string]string{"en-US": "Enter channel to send!", "zh-TW": "輸入要發送的頻道!", "zh-CN": "输入要发送的频道"}) {
		t.Fatalf("channel localizations = %#v / %#v", channel.NameLocalizations, channel.DescriptionLocalizations)
	}

	list := ListDefinition()
	if list.Type != commands.CommandTypeChatInput || list.Name != "自動通知列表" || list.Description != "查看所有的自動通知列表" || list.DefaultMemberPermissions != nil || len(list.Options) != 0 {
		t.Fatalf("list definition = %#v", list)
	}
	deleteDefinition := DeleteDefinition()
	if deleteDefinition.Type != commands.CommandTypeChatInput || deleteDefinition.Name != "自動通知刪除" || deleteDefinition.Description != "刪除之前設定的自動通知" || deleteDefinition.DefaultMemberPermissions != nil || len(deleteDefinition.Options) != 1 {
		t.Fatalf("delete definition = %#v", deleteDefinition)
	}
	deleteID := deleteDefinition.Options[0]
	if deleteID.Type != commands.OptionTypeString || deleteID.Name != "id" || deleteID.Description != "輸入要刪除的自動通知id!" || !deleteID.Required {
		t.Fatalf("delete id option = %#v", deleteID)
	}
}

func TestAutoNotificationSetupModalMatchesLegacyPayload(t *testing.T) {
	want := responses.Modal{
		CustomID: "1700000000000",
		Title:    "自動發送通知系統!",
		Rows: []responses.ModalRow{
			{Inputs: []responses.TextInput{{CustomID: "cron_setcron", Label: "請輸入corn表達式(如想用簡化版，請直接輸入取消或cancel就可以簡易設置corn)", Style: responses.TextInputStyleShort, Required: true}}},
			{Inputs: []responses.TextInput{{CustomID: "cron_setmsg", Label: "請輸入文字(如不輸入這項請務必輸入下面三項)", Style: responses.TextInputStyleParagraph}}},
			{Inputs: []responses.TextInput{{CustomID: "cron_setcolor", Label: "請輸入你的嵌入訊息顏色(如不輸入嵌入訊息相關，請務必輸入文字)", Style: responses.TextInputStyleShort}}},
			{Inputs: []responses.TextInput{{CustomID: "cron_settitle", Label: "請輸入你的嵌入標題(如不輸入嵌入訊息相關，請務必輸入文字)", Style: responses.TextInputStyleShort}}},
			{Inputs: []responses.TextInput{{CustomID: "cron_setcontent", Label: "請輸入嵌入內文(如不輸入嵌入訊息相關，請務必輸入文字)", Style: responses.TextInputStyleParagraph}}},
		},
	}
	if got := autoNotificationSetupModal("1700000000000"); !reflect.DeepEqual(got, want) {
		t.Fatalf("setup modal = %#v, want %#v", got, want)
	}
}

func TestAutoNotificationMessagesMatchLegacyVisibleContract(t *testing.T) {
	errorMessage := autoNotificationErrorMessage("你需要有`訊息管理`才能使用此指令")
	if len(errorMessage.Embeds) != 1 || errorMessage.Embeds[0].Title != "<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`訊息管理`才能使用此指令" || errorMessage.Embeds[0].Description != "<a:arrow_pink:996242460294512690> [點我前往教學網址](https://youtu.be/D43zPrZU5Fw)" || errorMessage.Embeds[0].Color != 0xED4245 {
		t.Fatalf("error message = %#v", errorMessage)
	}

	complete := autoNotificationSetupCompleteMessage("1700000000000")
	wantComplete := ":white_check_mark:**以下是該自動通知id:**`1700000000000`\n使用`/自動通知刪除 id:1700000000000`進行刪除\n~~我只是個分隔線，下面是你的訊息預覽~~"
	if complete.Content != wantComplete {
		t.Fatalf("setup completion = %q", complete.Content)
	}

	deleted := autoNotificationDeleteMessage()
	if len(deleted.Embeds) != 1 || deleted.Embeds[0].Title != "<a:green_tick:994529015652163614>自動通知系統" || deleted.Embeds[0].Description != "<:trashbin:995991389043163257>成功刪除該自動通知" || deleted.Embeds[0].Color != 0x57F287 {
		t.Fatalf("delete message = %#v", deleted)
	}

	listed := autoNotificationListMessage("測試伺服器", []domain.AutoNotificationSchedule{
		{ID: "one", Cron: "  */30 * * * *  ", ChannelID: "channel-1"},
		{ID: "two", Cron: "0 9 * * 7", ChannelID: "channel-2"},
	}, 0x123456)
	wantList := "輸入`/自動通知刪除 id`可進行刪除之前設定的自動通知\n\n**❰1❱ id:`one` cron設定:`  */30 * * * *  ` 頻道:**<#channel-1> \n**❰2❱ id:`two` cron設定:`0 9 * * 7` 頻道:**<#channel-2>"
	if len(listed.Embeds) != 1 || listed.Embeds[0].Title != "<:list:992002476360343602> 以下是測試伺服器的所有自動通知id" || listed.Embeds[0].Description != wantList || listed.Embeds[0].Color != 0x123456 {
		t.Fatalf("list message = %#v", listed)
	}
}

func TestAutoNotificationWizardMatchesLegacyControls(t *testing.T) {
	module := Module{color: func() int { return 0x123456 }}
	expiresAt := time.Unix(1_700_000_300, 600*time.Millisecond.Nanoseconds())
	week := module.autoNotificationWeekMessage("week-id", expiresAt, "https://example.test/avatar.png")
	if !strings.Contains(week.Embeds[0].Description, "<t:1700000301:R>") || week.Components[0].Components[0].MinValues != 1 || week.Components[0].Components[0].MaxValues != 7 || len(week.Components[0].Components[0].Options) != 7 {
		t.Fatalf("week message = %#v", week)
	}
	hour := module.autoNotificationHourMessage("hour-id", expiresAt, "avatar")
	hourControl := hour.Components[0].Components[0]
	if hourControl.MinValues != 1 || hourControl.MaxValues != 24 || len(hourControl.Options) != 24 || hourControl.Options[0].Value != "1" || hourControl.Options[23].Value != "0" {
		t.Fatalf("hour control = %#v", hourControl)
	}
	minute := module.autoNotificationMinuteMessage("minute-id", expiresAt, "avatar")
	minuteControl := minute.Components[0].Components[0]
	if minuteControl.MinValues != 1 || minuteControl.MaxValues != 6 || len(minuteControl.Options) != 12 || minuteControl.Options[0].Value != "0" || minuteControl.Options[11].Value != "55" {
		t.Fatalf("minute control = %#v", minuteControl)
	}
}
