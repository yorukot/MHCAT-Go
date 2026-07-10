package autochat

import (
	"reflect"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

func TestAutoChatConfigMessagesMatchLegacyPayloads(t *testing.T) {
	wantSet := responses.Message{
		Embeds: []responses.Embed{{
			Title:       "自動聊天系統",
			Description: "您的自動聊天頻道成功創建\n您目前的自動聊天頻道為 <#channel-1>",
			Color:       0x57F287,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
	if got := autoChatSetSuccessMessage("channel-1"); !reflect.DeepEqual(got, wantSet) {
		t.Fatalf("set success = %#v, want %#v", got, wantSet)
	}

	wantDelete := responses.Message{
		Embeds: []responses.Embed{{
			Title:       "自動聊天系統",
			Description: "您的自動聊天頻道成功刪除",
			Color:       0x57F287,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
	if got := autoChatDeleteSuccessMessage(); !reflect.DeepEqual(got, wantDelete) {
		t.Fatalf("delete success = %#v, want %#v", got, wantDelete)
	}

	wantPermissionError := responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`訊息管理`才能使用此指令",
			Color: 0xED4245,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
	if got := autoChatErrorMessage("你需要有`訊息管理`才能使用此指令"); !reflect.DeepEqual(got, wantPermissionError) {
		t.Fatalf("permission error = %#v, want %#v", got, wantPermissionError)
	}

	wantUnknown := autoChatErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
	if got := autoChatUnknownError(nil); !reflect.DeepEqual(got, wantUnknown) {
		t.Fatalf("unknown error = %#v, want %#v", got, wantUnknown)
	}
}

func TestPaidAutoChatVisibleContract(t *testing.T) {
	if LegacyAutoChatPaidResponseDelay != 10*time.Second {
		t.Fatalf("paid response delay = %s", LegacyAutoChatPaidResponseDelay)
	}
	if legacyAutoChatUnsafeInputDelay != 4*time.Second || legacyAutoChatBusyWarningDelay != 2*time.Second {
		t.Fatalf("warning delays = unsafe %s busy %s", legacyAutoChatUnsafeInputDelay, legacyAutoChatBusyWarningDelay)
	}
	if legacyAutoChatUnsafeInputMessage != "為防止伺服器招到tag攻擊，請勿在與機器人聊天時含有@" {
		t.Fatalf("unsafe input message = %q", legacyAutoChatUnsafeInputMessage)
	}
	if legacyAutoChatBusyMessage != "一次只能傳送一個消息，請等待機器人回復完成後在繼續!" {
		t.Fatalf("busy message = %q", legacyAutoChatBusyMessage)
	}
	if legacyAutoChatUnsafeOutputMessage != "由於chatGPT回傳回來的訊息含有@，為防止遭到tag攻擊，已自動迴避該則消息!" {
		t.Fatalf("unsafe output message = %q", legacyAutoChatUnsafeOutputMessage)
	}
}
