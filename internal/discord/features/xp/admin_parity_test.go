package xp

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestXPAdminSuccessMessageMatchesLegacy(t *testing.T) {
	want := responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<:xp:990254386792005663> 經驗系統",
			Description: "<a:green_tick:994529015652163614>成功為:<@user-2>\n增加:`-75`",
			Color:       0x53FF53,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
	if got := xpAdminSuccessMessage(" user-2 ", -75); !reflect.DeepEqual(got, want) {
		t.Fatalf("success message = %#v, want %#v", got, want)
	}
}

func TestXPAdminHandlerErrorsMatchLegacy(t *testing.T) {
	t.Run("permission", func(t *testing.T) {
		module := NewAdminModule(fakemongo.NewXPAdminRepository(), nil)
		interaction := fakediscord.SlashInteractionWithOptions(XPAdminCommandName, "聊天經驗改變", map[string]string{
			"使用者": "user-2",
			"經驗值": "1",
		})
		responder := fakediscord.NewResponder()
		if err := module.AdminHandler()(context.Background(), interaction, responder); err != nil {
			t.Fatalf("handler: %v", err)
		}
		assertXPConfigPublicDefer(t, responder)
		assertXPConfigEdit(t, responder, "<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`踢出用戶`才能使用此指令", "", 0xED4245)
	})

	t.Run("persistence", func(t *testing.T) {
		repo := fakemongo.NewXPAdminRepository()
		repo.Err = errors.New("database unavailable")
		module := NewAdminModule(repo, nil)
		interaction := fakediscord.SlashInteractionWithOptions(XPAdminCommandName, "語音經驗改變", map[string]string{
			"使用者": "user-2",
			"經驗值": "1",
		})
		interaction.Actor.PermissionBits = permissionKickMembers
		responder := fakediscord.NewResponder()
		if err := module.AdminHandler()(context.Background(), interaction, responder); err != nil {
			t.Fatalf("handler: %v", err)
		}
		assertXPConfigPublicDefer(t, responder)
		assertXPConfigEdit(t, responder, "<a:Discord_AnimatedNo:1015989839809757295> | 很抱歉，出現了未知的錯誤，請重試!", "", 0xED4245)
	})
}
