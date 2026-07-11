package autochat

import (
	"reflect"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestDefinitionsMatchLegacyCommandShape(t *testing.T) {
	want := []commands.Definition{
		{
			Type:        commands.CommandTypeChatInput,
			Name:        "自動聊天頻道",
			Description: "設定自動聊天頻道要在哪裡發送",
			DocsURL:     "https://docsmhcat.yorukot.me/docs/chat_xp_set",
			Ownership:   commands.ManagedOwnership("autochat-config", commands.ScopeGuild),
			Options: []commands.Option{{
				Type:         commands.OptionTypeChannel,
				Name:         "頻道",
				Description:  "輸入頻道!",
				Required:     true,
				ChannelTypes: []int{0, 5},
			}},
		},
		{
			Type:        commands.CommandTypeChatInput,
			Name:        "自動聊天頻道刪除",
			Description: "刪除自動聊天頻道要在哪裡發送",
			DocsURL:     "https://docsmhcat.yorukot.me/docs/chat_xp_set",
			Ownership:   commands.ManagedOwnership("autochat-config", commands.ScopeGuild),
		},
	}
	if got := Definitions(); !reflect.DeepEqual(got, want) {
		t.Fatalf("definitions = %#v, want %#v", got, want)
	}
}
