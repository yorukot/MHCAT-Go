package xp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image/png"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakebotinfo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

const (
	legacyXPConfigErrorColor   = 0xED4245
	legacyXPConfigSuccessColor = 0x57F287
)

func TestSetHandlerRendersLegacySuccessAndPreview(t *testing.T) {
	repo := fakemongo.NewTextXPConfigRepository()
	repo.Configs["guild-1"] = domain.TextXPConfig{GuildID: "guild-1", ChannelID: "old-channel", Color: "red", Message: "old"}
	sideEffects := fakediscord.NewSideEffects()
	usage := &fakeusage.Tracker{}
	module := NewModule(repo, sideEffects, usage)
	interaction := fakediscord.SlashInteractionWithOptions(TextXPSetCommandName, "", map[string]string{
		"頻道": "channel-1",
		"訊息": "  {user} 升到了 {level}  ",
		"顏色": "rgba(0, 0, 0, .45)",
	})
	interaction.ChannelID = "invoke-channel"
	interaction.Actor.PermissionBits = permissionManageMessages
	responder := fakediscord.NewResponder()

	if err := module.SetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	assertXPConfigPublicDefer(t, responder)
	assertXPConfigEdit(t, responder, "聊天經驗系統", "您的聊天經驗升等頻道成功創建\n您目前的升等通知頻道為 <#channel-1>", legacyXPConfigSuccessColor)
	wantSaved := domain.TextXPConfig{GuildID: "guild-1", ChannelID: "channel-1", Color: "rgba(0, 0, 0, .45)", Message: "  {user} 升到了 {level}  "}
	if saved := repo.Configs["guild-1"]; len(repo.Configs) != 1 || !reflect.DeepEqual(saved, wantSaved) {
		t.Fatalf("saved config = %#v, want %#v", saved, wantSaved)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].ChannelID != "invoke-channel" {
		t.Fatalf("preview sends = %#v", sideEffects.Sent)
	}
	wantPreview := ports.OutboundMessage{
		Content:         "以下為你的訊息預覽:\n<:line:992363971803881493>我<:line:992363971803881493>只<:line:992363971803881493>是<:line:992363971803881493>分<:line:992363971803881493>隔<:line:992363971803881493>線<:line:992363971803881493>\n\n  {user} 升到了 {level}  ",
		AllowedMentions: ports.AllowedMentions{},
	}
	if !reflect.DeepEqual(sideEffects.Sent[0].Message, wantPreview) {
		t.Fatalf("preview = %#v, want %#v", sideEffects.Sent[0].Message, wantPreview)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != TextXPSetCommandName {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestRankSlashRepliesLoadingThenRendersPNGAndLegacyButtons(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	viewerID := "123456789012345678"
	_ = repo.SaveTextXPProfile(context.Background(), domain.XPProfile{GuildID: "guild-1", UserID: viewerID, Level: 1, XP: 1000})
	_ = repo.SaveTextXPProfile(context.Background(), domain.XPProfile{GuildID: "guild-1", UserID: "222222222222222222", Level: 3, XP: 0})
	info := &fakebotinfo.DiscordInfoProvider{
		Users: map[string]ports.DiscordUserInfo{
			viewerID:             {Username: "Viewer"},
			"222222222222222222": {Username: "Leader"},
		},
		Guild: ports.DiscordGuildInfo{Name: "Guild", CreatedAt: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)},
	}
	usage := &fakeusage.Tracker{}
	module := NewRankModule(repo, info, usage)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction(TextXPRankCommandName)
	interaction.Actor.UserID = viewerID
	interaction.Actor.AvatarURL = "https://example.test/avatar.png"

	if err := module.TextHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Replies) != 1 || len(responder.Replies[0].Embeds) != 1 || responder.Replies[0].Embeds[0].Author.Name != rankLoadingAuthor {
		t.Fatalf("loading reply = %#v", responder.Replies)
	}
	if len(responder.Edits) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	assertXPRankImage(t, responder.Edits[0].Files)
	if len(responder.Edits[0].Components) != 2 {
		t.Fatalf("components = %#v", responder.Edits[0].Components)
	}
	buttons := responder.Edits[0].Components[0].Components
	if buttons[0].CustomID != "["+viewerID+"]{-10}text_rank" || !buttons[0].Disabled || buttons[2].CustomID != "text_rank" || buttons[2].Label != "1/1" {
		t.Fatalf("pagination buttons = %#v", buttons)
	}
	target := responder.Edits[0].Components[1].Components[2]
	if target.CustomID != "["+viewerID+"]text_rank {0}" || target.Emoji != legacyRankTargetViewerEmoji || target.Disabled {
		t.Fatalf("target button = %#v", target)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != TextXPRankCommandName || usage.Events[0].Feature != "xp-rank" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestRankComponentUpdatesRequestedVoicePage(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	viewerID := "123456789012345678"
	users := map[string]ports.DiscordUserInfo{viewerID: {Username: "Viewer"}}
	for i := 0; i < 12; i++ {
		userID := fmt.Sprintf("%018d", 222222222222222222+i)
		if i == 0 {
			userID = viewerID
		}
		_ = repo.SaveVoiceXPProfile(context.Background(), domain.XPProfile{GuildID: "guild-1", UserID: userID, Level: 1, XP: int64(100 - i)})
		users[userID] = ports.DiscordUserInfo{Username: "User"}
	}
	module := NewRankModule(repo, &fakebotinfo.DiscordInfoProvider{Users: users, Guild: ports.DiscordGuildInfo{Name: "Guild"}}, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.ComponentInteractionFromID("[" + viewerID + "]{1}voice_rank")

	if err := module.VoicePageHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Updates) != 1 {
		t.Fatalf("updates = %#v", responder.Updates)
	}
	assertXPRankImage(t, responder.Updates[0].Files)
	if label := responder.Updates[0].Components[0].Components[2].Label; label != "2/2" {
		t.Fatalf("page label = %q", label)
	}
	if target := responder.Updates[0].Components[1].Components[2].CustomID; target != "["+viewerID+"]voice_rank {1}" {
		t.Fatalf("target custom id = %q", target)
	}
}

func TestRankComponentMissingUserUsesLegacyEphemeralError(t *testing.T) {
	module := NewRankModule(fakemongo.NewXPAdminRepository(), &fakebotinfo.DiscordInfoProvider{}, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.ComponentInteractionFromID("[123456789012345678]{1}text_rank")

	if err := module.TextPageHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Replies) != 1 || !responder.Replies[0].Ephemeral {
		t.Fatalf("replies = %#v", responder.Replies)
	}
	if title := responder.Replies[0].Embeds[0].Title; !strings.Contains(title, "找不到資料") {
		t.Fatalf("title = %q", title)
	}
}

func TestRankModuleRegistersLegacyComponentRoutes(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	viewerID := "123456789012345678"
	_ = repo.SaveTextXPProfile(context.Background(), domain.XPProfile{GuildID: "guild-1", UserID: viewerID, XP: 1})
	info := &fakebotinfo.DiscordInfoProvider{Users: map[string]ports.DiscordUserInfo{viewerID: {Username: "Viewer"}}}
	module := NewRankModule(repo, info, nil)
	router := interactions.NewRouter()
	router.SetCustomIDParser(interactions.DefaultCustomIDParser{})
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register: %v", err)
	}
	responder := fakediscord.NewResponder()
	interaction := fakediscord.ComponentInteractionFromID("[" + viewerID + "]{0}text_rank")
	if err := router.Handle(context.Background(), interaction, responder); err != nil {
		t.Fatalf("route: %v", err)
	}
	if len(responder.Updates) != 1 {
		t.Fatalf("updates = %#v", responder.Updates)
	}
}

func assertXPRankImage(t *testing.T, files []responses.File) {
	t.Helper()
	if len(files) != 1 || files[0].Name != rankFileName || files[0].ContentType != rankFileContentType {
		t.Fatalf("files = %#v", files)
	}
	img, err := png.Decode(bytes.NewReader(files[0].Data))
	if err != nil {
		t.Fatalf("decode png: %v", err)
	}
	if bounds := img.Bounds(); bounds.Dx() != 1000 || bounds.Dy() != 500 {
		t.Fatalf("bounds = %v", bounds)
	}
}

func TestSetHandlerRejectsMissingPermissionAndInvalidColor(t *testing.T) {
	module := NewModule(fakemongo.NewTextXPConfigRepository(), nil, nil)
	interaction := fakediscord.SlashInteractionWithOptions(TextXPSetCommandName, "", map[string]string{"頻道": "channel-1"})
	responder := fakediscord.NewResponder()
	if err := module.SetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	assertXPConfigPublicDefer(t, responder)
	assertXPConfigEdit(t, responder, "<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`訊息管理`才能使用此指令", "", legacyXPConfigErrorColor)

	interaction.Actor.PermissionBits = permissionManageMessages
	for _, color := range []string{"ffffff", " #fff"} {
		interaction.Options["顏色"] = color
		responder = fakediscord.NewResponder()
		if err := module.SetHandler()(context.Background(), interaction, responder); err != nil {
			t.Fatalf("handler for %q: %v", color, err)
		}
		assertLegacyXPColorError(t, responder)
	}
}

func TestDeleteHandlerSuccessAndMissing(t *testing.T) {
	repo := fakemongo.NewTextXPConfigRepository()
	repo.Configs["guild-1"] = domain.TextXPConfig{GuildID: "guild-1", ChannelID: "channel-1"}
	usage := &fakeusage.Tracker{}
	module := NewModule(repo, nil, usage)
	interaction := fakediscord.SlashInteraction(TextXPDeleteCommandName)
	responder := fakediscord.NewResponder()
	if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("permission handler: %v", err)
	}
	if _, ok := repo.Configs["guild-1"]; !ok {
		t.Fatal("permission denial deleted the config")
	}
	assertXPConfigPublicDefer(t, responder)
	assertXPConfigEdit(t, responder, "<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`訊息管理`才能使用此指令", "", legacyXPConfigErrorColor)
	if len(usage.Events) != 0 {
		t.Fatalf("permission denial usage = %#v", usage.Events)
	}

	interaction.Actor.PermissionBits = permissionManageMessages
	responder = fakediscord.NewResponder()
	if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if _, ok := repo.Configs["guild-1"]; ok {
		t.Fatal("config was not deleted")
	}
	assertXPConfigPublicDefer(t, responder)
	assertXPConfigEdit(t, responder, "聊天經驗系統", "成功刪除!", legacyXPConfigSuccessColor)
	if len(usage.Events) != 1 || usage.Events[0].CommandName != TextXPDeleteCommandName {
		t.Fatalf("usage = %#v", usage.Events)
	}

	responder = fakediscord.NewResponder()
	if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler missing: %v", err)
	}
	assertXPConfigPublicDefer(t, responder)
	assertXPConfigEdit(t, responder, "<a:Discord_AnimatedNo:1015989839809757295> | 你本來就沒有對聊天經驗設定喔!", "", legacyXPConfigErrorColor)
}

func TestVoiceSetHandlerRendersLegacySuccessAndIgnoresBackground(t *testing.T) {
	repo := fakemongo.NewVoiceXPConfigRepository()
	repo.Configs["guild-1"] = domain.VoiceXPConfig{GuildID: "guild-1", ChannelID: "old-channel", Color: "red", Message: "old"}
	sideEffects := fakediscord.NewSideEffects()
	usage := &fakeusage.Tracker{}
	module := NewVoiceModule(repo, sideEffects, usage)
	interaction := fakediscord.SlashInteractionWithOptions(VoiceXPSetCommandName, "", map[string]string{
		"頻道": "voice-channel-1",
		"訊息": "  {user} 升到了 {level}  ",
		"顏色": "hwb(180deg 0% 0% / 100%)",
		"背景": "https://example.invalid/background.png",
	})
	interaction.ChannelID = "invoke-channel"
	interaction.Actor.PermissionBits = permissionManageMessages
	responder := fakediscord.NewResponder()

	if err := module.SetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	assertXPConfigPublicDefer(t, responder)
	assertXPConfigEdit(t, responder, "語音經驗系統", "您的語音經驗升等頻道成功創建\n您目前的升等通知頻道為 <#voice-channel-1>", legacyXPConfigSuccessColor)
	wantSaved := domain.VoiceXPConfig{GuildID: "guild-1", ChannelID: "voice-channel-1", Color: "hwb(180deg 0% 0% / 100%)", Message: "  {user} 升到了 {level}  "}
	if saved := repo.Configs["guild-1"]; len(repo.Configs) != 1 || !reflect.DeepEqual(saved, wantSaved) {
		t.Fatalf("saved config = %#v, want %#v", saved, wantSaved)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].ChannelID != "invoke-channel" {
		t.Fatalf("preview sends = %#v", sideEffects.Sent)
	}
	wantPreview := ports.OutboundMessage{
		Content:         "以下為你的訊息預覽:\n<:line:992363971803881493>我<:line:992363971803881493>只<:line:992363971803881493>是<:line:992363971803881493>分<:line:992363971803881493>隔<:line:992363971803881493>線<:line:992363971803881493>\n\n  {user} 升到了 {level}  ",
		AllowedMentions: ports.AllowedMentions{},
	}
	if !reflect.DeepEqual(sideEffects.Sent[0].Message, wantPreview) {
		t.Fatalf("preview = %#v, want %#v", sideEffects.Sent[0].Message, wantPreview)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != VoiceXPSetCommandName || usage.Events[0].Feature != "voice-xp-config" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestVoiceSetHandlerRejectsMissingPermissionAndInvalidColor(t *testing.T) {
	module := NewVoiceModule(fakemongo.NewVoiceXPConfigRepository(), nil, nil)
	interaction := fakediscord.SlashInteractionWithOptions(VoiceXPSetCommandName, "", map[string]string{"頻道": "channel-1"})
	responder := fakediscord.NewResponder()
	if err := module.SetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	assertXPConfigPublicDefer(t, responder)
	assertXPConfigEdit(t, responder, "<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`訊息管理`才能使用此指令", "", legacyXPConfigErrorColor)

	interaction.Actor.PermissionBits = permissionManageMessages
	for _, color := range []string{"ffffff", " #fff"} {
		interaction.Options["顏色"] = color
		responder = fakediscord.NewResponder()
		if err := module.SetHandler()(context.Background(), interaction, responder); err != nil {
			t.Fatalf("handler for %q: %v", color, err)
		}
		assertLegacyXPColorError(t, responder)
	}
}

func assertLegacyXPColorError(t *testing.T, responder *fakediscord.Responder) {
	t.Helper()
	assertXPConfigPublicDefer(t, responder)
	assertXPConfigEdit(t, responder, "<a:Discord_AnimatedNo:1015989839809757295> | 你傳送的並不是顏色(色碼)", "", legacyXPConfigErrorColor)
}

func TestVoiceDeleteHandlerSuccessAndMissing(t *testing.T) {
	repo := fakemongo.NewVoiceXPConfigRepository()
	repo.Configs["guild-1"] = domain.VoiceXPConfig{GuildID: "guild-1", ChannelID: "channel-1"}
	usage := &fakeusage.Tracker{}
	module := NewVoiceModule(repo, nil, usage)
	interaction := fakediscord.SlashInteraction(VoiceXPDeleteCommandName)
	responder := fakediscord.NewResponder()
	if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("permission handler: %v", err)
	}
	if _, ok := repo.Configs["guild-1"]; !ok {
		t.Fatal("permission denial deleted the voice config")
	}
	assertXPConfigPublicDefer(t, responder)
	assertXPConfigEdit(t, responder, "<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`訊息管理`才能使用此指令", "", legacyXPConfigErrorColor)
	if len(usage.Events) != 0 {
		t.Fatalf("permission denial usage = %#v", usage.Events)
	}

	interaction.Actor.PermissionBits = permissionManageMessages
	responder = fakediscord.NewResponder()
	if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if _, ok := repo.Configs["guild-1"]; ok {
		t.Fatal("voice config was not deleted")
	}
	assertXPConfigPublicDefer(t, responder)
	assertXPConfigEdit(t, responder, "語音經驗系統", "成功刪除!", legacyXPConfigSuccessColor)
	if len(usage.Events) != 1 || usage.Events[0].CommandName != VoiceXPDeleteCommandName {
		t.Fatalf("usage = %#v", usage.Events)
	}

	responder = fakediscord.NewResponder()
	if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler missing: %v", err)
	}
	assertXPConfigPublicDefer(t, responder)
	assertXPConfigEdit(t, responder, "<a:Discord_AnimatedNo:1015989839809757295> | 你本來就沒有對語音經驗設定喔!", "", legacyXPConfigErrorColor)
}

func TestXPSetHandlersSkipPreviewWithoutCustomMessage(t *testing.T) {
	t.Run("text", func(t *testing.T) {
		repo := fakemongo.NewTextXPConfigRepository()
		messages := fakediscord.NewSideEffects()
		interaction := fakediscord.SlashInteractionWithOptions(TextXPSetCommandName, "", map[string]string{"頻道": "channel-1"})
		interaction.ChannelID = "invoke-channel"
		interaction.Actor.PermissionBits = permissionManageMessages
		responder := fakediscord.NewResponder()

		if err := NewModule(repo, messages, nil).SetHandler()(context.Background(), interaction, responder); err != nil {
			t.Fatalf("handler: %v", err)
		}
		if len(messages.Sent) != 0 {
			t.Fatalf("preview sends = %#v", messages.Sent)
		}
		if saved := repo.Configs["guild-1"]; saved.Color != "" || saved.Message != "" {
			t.Fatalf("saved config = %#v", saved)
		}
	})

	t.Run("voice", func(t *testing.T) {
		repo := fakemongo.NewVoiceXPConfigRepository()
		messages := fakediscord.NewSideEffects()
		interaction := fakediscord.SlashInteractionWithOptions(VoiceXPSetCommandName, "", map[string]string{
			"頻道": "channel-1",
			"背景": "https://example.invalid/ignored.png",
		})
		interaction.ChannelID = "invoke-channel"
		interaction.Actor.PermissionBits = permissionManageMessages
		responder := fakediscord.NewResponder()

		if err := NewVoiceModule(repo, messages, nil).SetHandler()(context.Background(), interaction, responder); err != nil {
			t.Fatalf("handler: %v", err)
		}
		if len(messages.Sent) != 0 {
			t.Fatalf("preview sends = %#v", messages.Sent)
		}
		if saved := repo.Configs["guild-1"]; saved.Color != "" || saved.Message != "" {
			t.Fatalf("saved config = %#v", saved)
		}
	})
}

func TestXPConfigRepositoryFailuresUseSafeLegacyErrorPayload(t *testing.T) {
	repositoryErr := errors.New("mongo unavailable")
	textRepo := fakemongo.NewTextXPConfigRepository()
	textRepo.Err = repositoryErr
	voiceRepo := fakemongo.NewVoiceXPConfigRepository()
	voiceRepo.Err = repositoryErr

	for _, tc := range []struct {
		name        string
		handler     interactions.Handler
		interaction interactions.Interaction
	}{
		{name: "text set", handler: NewModule(textRepo, nil, nil).SetHandler(), interaction: fakediscord.SlashInteractionWithOptions(TextXPSetCommandName, "", map[string]string{"頻道": "channel-1"})},
		{name: "text delete", handler: NewModule(textRepo, nil, nil).DeleteHandler(), interaction: fakediscord.SlashInteraction(TextXPDeleteCommandName)},
		{name: "voice set", handler: NewVoiceModule(voiceRepo, nil, nil).SetHandler(), interaction: fakediscord.SlashInteractionWithOptions(VoiceXPSetCommandName, "", map[string]string{"頻道": "channel-1"})},
		{name: "voice delete", handler: NewVoiceModule(voiceRepo, nil, nil).DeleteHandler(), interaction: fakediscord.SlashInteraction(VoiceXPDeleteCommandName)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.interaction.Actor.PermissionBits = permissionManageMessages
			responder := fakediscord.NewResponder()
			if err := tc.handler(context.Background(), tc.interaction, responder); err != nil {
				t.Fatalf("handler: %v", err)
			}
			assertXPConfigPublicDefer(t, responder)
			assertXPConfigEdit(t, responder, "<a:Discord_AnimatedNo:1015989839809757295> | 很抱歉，出現了未知的錯誤，請重試!", "", legacyXPConfigErrorColor)
		})
	}
}

func assertXPConfigPublicDefer(t *testing.T, responder *fakediscord.Responder) {
	t.Helper()
	want := []responses.DeferOptions{{}}
	if !reflect.DeepEqual(responder.Defers, want) {
		t.Fatalf("defers = %#v, want %#v", responder.Defers, want)
	}
}

func assertXPConfigEdit(t *testing.T, responder *fakediscord.Responder, title string, description string, color int) {
	t.Helper()
	want := responses.Message{
		Embeds: []responses.Embed{{
			Title:       title,
			Description: description,
			Color:       color,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
	if len(responder.Edits) != 1 || !reflect.DeepEqual(responder.Edits[0], want) {
		t.Fatalf("edits = %#v, want %#v", responder.Edits, want)
	}
}

func TestDisabledProfileHandlersReturnLegacyRemovalMessage(t *testing.T) {
	usage := &fakeusage.Tracker{}
	module := NewDisabledProfileModule(usage)

	for index, tc := range []struct {
		name    string
		handler func() interactions.Handler
	}{
		{name: TextXPProfileCommandName, handler: module.TextHandler},
		{name: VoiceXPProfileCommandName, handler: module.VoiceHandler},
	} {
		t.Run(tc.name, func(t *testing.T) {
			interaction := fakediscord.SlashInteractionWithOptions(tc.name, "", map[string]string{"玩家": "ignored-user"})
			responder := fakediscord.NewResponder()
			if err := tc.handler()(context.Background(), interaction, responder); err != nil {
				t.Fatalf("handler: %v", err)
			}
			if len(responder.Defers) != 1 {
				t.Fatalf("defers = %#v", responder.Defers)
			}
			if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
				t.Fatalf("edits = %#v", responder.Edits)
			}
			embed := responder.Edits[0].Embeds[0]
			if embed.Title != "<a:Discord_AnimatedNo:1015989839809757295> | "+disabledProfileMessage || embed.Color != textXPErrorColor {
				t.Fatalf("embed = %#v", embed)
			}
			if responder.Edits[0].AllowedMentions == nil {
				t.Fatalf("allowed mentions should be explicitly empty: %#v", responder.Edits[0])
			}
			message := responder.Edits[0]
			if message.Content != "" || embed.Description != "" || len(message.Components) != 0 || len(message.Files) != 0 || message.Ephemeral {
				t.Fatalf("unexpected payload fields = %#v", message)
			}
			if len(usage.Events) != index+1 || usage.Events[index].UserID == "ignored-user" {
				t.Fatalf("player option affected disabled path: %#v", usage.Events)
			}
		})
	}

	if len(usage.Events) != 2 {
		t.Fatalf("usage = %#v", usage.Events)
	}
	if usage.Events[0].CommandName != TextXPProfileCommandName || usage.Events[0].Feature != "xp-profile-disabled" {
		t.Fatalf("text usage = %#v", usage.Events[0])
	}
	if usage.Events[1].CommandName != VoiceXPProfileCommandName || usage.Events[1].Feature != "xp-profile-disabled" {
		t.Fatalf("voice usage = %#v", usage.Events[1])
	}
}

func TestAdminHandlerAddsTextXPAndRendersLegacySuccess(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	usage := &fakeusage.Tracker{}
	module := NewAdminModule(repo, usage)
	interaction := fakediscord.SlashInteractionWithOptions(XPAdminCommandName, "聊天經驗改變", map[string]string{
		"使用者": "user-2",
		"經驗值": "150",
	})
	interaction.Actor.PermissionBits = permissionKickMembers
	responder := fakediscord.NewResponder()

	if err := module.AdminHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	profile := repo.TextProfiles["guild-1/user-2"]
	if profile.Level != 1 || profile.XP != 50 {
		t.Fatalf("profile = %#v", profile)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != "<:xp:990254386792005663> 經驗系統" || embed.Description != doneEmoji+"成功為:<@user-2>\n增加:`150`" {
		t.Fatalf("embed = %#v", embed)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != XPAdminCommandName || usage.Events[0].Feature != "xp-admin" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestAdminHandlerAddsVoiceXPWithTypedOptions(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	repo.VoiceProfiles["guild-1/user-2"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-2", Level: 2, XP: 50}
	module := NewAdminModule(repo, nil)
	interaction := fakediscord.SlashInteractionWithOptions(XPAdminCommandName, "語音經驗改變", map[string]string{})
	interaction.Actor.PermissionBits = permissionKickMembers
	interaction.CommandOptions = map[string]interactions.CommandOptionValue{
		"使用者": {Type: interactions.CommandOptionUser, String: "user-2"},
		"經驗值": {Type: interactions.CommandOptionInteger, Int: 500, String: "500"},
	}
	responder := fakediscord.NewResponder()

	if err := module.AdminHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	profile := repo.VoiceProfiles["guild-1/user-2"]
	if profile.Level != 3 || profile.XP != 250 {
		t.Fatalf("profile = %#v", profile)
	}
}

func TestAdminHandlerRequiresKickMembers(t *testing.T) {
	module := NewAdminModule(fakemongo.NewXPAdminRepository(), nil)
	interaction := fakediscord.SlashInteractionWithOptions(XPAdminCommandName, "聊天經驗改變", map[string]string{
		"使用者": "user-2",
		"經驗值": "1",
	})
	responder := fakediscord.NewResponder()

	if err := module.AdminHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "踢出用戶") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestResetHandlerRequiresGuildOwner(t *testing.T) {
	module := NewResetModule(
		fakemongo.NewXPAdminRepository(),
		&fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{OwnerID: "owner-1"}},
		fakediscord.NewSideEffects(),
		nil,
		&xpResetTestClock{now: time.Unix(1000, 0)},
	)
	interaction := fakediscord.SlashInteractionWithOptions(XPResetCommandName, "重製個人聊天經驗", map[string]string{"使用者": "user-2"})
	responder := fakediscord.NewResponder()

	if err := module.ResetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "你必須擁有`服主`才能使用") {
		t.Fatalf("owner response = %#v", responder.Edits)
	}
}

func TestResetHandlerDeletesIndividualProfiles(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	repo.TextProfiles["guild-1/user-2"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-2", XP: 20, Level: 1}
	repo.VoiceProfiles["guild-1/user-3"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-3", XP: 30, Level: 2}
	usage := &fakeusage.Tracker{}
	module := NewResetModule(
		repo,
		&fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{OwnerID: "user-1"}},
		fakediscord.NewSideEffects(),
		usage,
		&xpResetTestClock{now: time.Unix(1000, 0)},
	)

	textInteraction := fakediscord.SlashInteractionWithOptions(XPResetCommandName, "重製個人聊天經驗", map[string]string{"使用者": "user-2"})
	responder := fakediscord.NewResponder()
	if err := module.ResetHandler()(context.Background(), textInteraction, responder); err != nil {
		t.Fatalf("text handler: %v", err)
	}
	if _, ok := repo.TextProfiles["guild-1/user-2"]; ok {
		t.Fatal("text profile was not deleted")
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Content != doneEmoji+" | 成功清除<@user-2>的聊天經驗" {
		t.Fatalf("text response = %#v", responder.Edits)
	}

	voiceInteraction := fakediscord.SlashInteractionWithOptions(XPResetCommandName, "重製個人語音經驗", nil)
	voiceInteraction.CommandOptions = map[string]interactions.CommandOptionValue{
		"使用者": {Type: interactions.CommandOptionUser, String: "user-3"},
	}
	responder = fakediscord.NewResponder()
	if err := module.ResetHandler()(context.Background(), voiceInteraction, responder); err != nil {
		t.Fatalf("voice handler: %v", err)
	}
	if _, ok := repo.VoiceProfiles["guild-1/user-3"]; ok {
		t.Fatal("voice profile was not deleted")
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Content != doneEmoji+" | 成功清除<@user-3>的語音經驗" {
		t.Fatalf("voice response = %#v", responder.Edits)
	}
	if len(usage.Events) != 2 || usage.Events[0].CommandName != XPResetCommandName || usage.Events[0].Feature != "xp-reset" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestResetHandlerReportsMissingIndividualProfile(t *testing.T) {
	module := NewResetModule(
		fakemongo.NewXPAdminRepository(),
		&fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{OwnerID: "user-1"}},
		fakediscord.NewSideEffects(),
		nil,
		&xpResetTestClock{now: time.Unix(1000, 0)},
	)
	interaction := fakediscord.SlashInteractionWithOptions(XPResetCommandName, "重製個人聊天經驗", map[string]string{"使用者": "user-2"})
	responder := fakediscord.NewResponder()

	if err := module.ResetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "這位使用者還沒有任何的經驗值喔!") {
		t.Fatalf("missing response = %#v", responder.Edits)
	}
}

func TestResetHandlerArmsGuildResetWarning(t *testing.T) {
	module := NewResetModule(
		fakemongo.NewXPAdminRepository(),
		&fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{OwnerID: "user-1"}},
		fakediscord.NewSideEffects(),
		nil,
		&xpResetTestClock{now: time.Unix(1000, 0)},
	)
	interaction := fakediscord.SlashInteractionWithOptions(XPResetCommandName, "聊天經驗重製", nil)
	interaction.ChannelID = "channel-1"
	responder := fakediscord.NewResponder()

	if err := module.ResetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Content != xpResetWarningContent {
		t.Fatalf("warning response = %#v", responder.Edits)
	}
}

func TestResetConfirmationCancelsOnWrongContent(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	repo.TextProfiles["guild-1/user-2"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-2"}
	sideEffects := fakediscord.NewSideEffects()
	clock := &xpResetTestClock{now: time.Unix(1000, 0)}
	module := NewResetModule(
		repo,
		&fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{OwnerID: "user-1"}},
		sideEffects,
		nil,
		clock,
	)
	interaction := fakediscord.SlashInteractionWithOptions(XPResetCommandName, "聊天經驗重製", nil)
	interaction.ChannelID = "channel-1"
	if err := module.ResetHandler()(context.Background(), interaction, fakediscord.NewResponder()); err != nil {
		t.Fatalf("arm handler: %v", err)
	}

	err := module.ConfirmationHandler()(context.Background(), events.Event{
		Type:      events.TypeMessageCreate,
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		UserID:    "user-1",
		Content:   "確認",
		CreatedAt: clock.now.Add(time.Second),
	})
	if err != nil {
		t.Fatalf("confirm handler: %v", err)
	}
	if len(sideEffects.Sent) != 1 || !strings.Contains(sideEffects.Sent[0].Message.Embeds[0].Title, "你輸入了錯誤的確認!因此視為取消還原") {
		t.Fatalf("wrong confirmation response = %#v", sideEffects.Sent)
	}
	if _, ok := repo.TextProfiles["guild-1/user-2"]; !ok {
		t.Fatal("profile should not be deleted after wrong confirmation")
	}

	err = module.ConfirmationHandler()(context.Background(), events.Event{
		Type:      events.TypeMessageCreate,
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		UserID:    "user-1",
		Content:   xpResetConfirmContent,
		CreatedAt: clock.now.Add(2 * time.Second),
	})
	if err != nil {
		t.Fatalf("second confirm handler: %v", err)
	}
	if len(sideEffects.Sent) != 1 {
		t.Fatalf("confirmation should be one-shot, sends = %#v", sideEffects.Sent)
	}
}

func TestResetConfirmationDeletesGuildProfiles(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	repo.VoiceProfiles["guild-1/user-2"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-2"}
	repo.VoiceProfiles["guild-2/user-3"] = domain.XPProfile{GuildID: "guild-2", UserID: "user-3"}
	sideEffects := fakediscord.NewSideEffects()
	clock := &xpResetTestClock{now: time.Unix(1000, 0)}
	module := NewResetModule(
		repo,
		&fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{OwnerID: "user-1"}},
		sideEffects,
		nil,
		clock,
	)
	interaction := fakediscord.SlashInteractionWithOptions(XPResetCommandName, "語音經驗重製", nil)
	interaction.ChannelID = "channel-1"
	if err := module.ResetHandler()(context.Background(), interaction, fakediscord.NewResponder()); err != nil {
		t.Fatalf("arm handler: %v", err)
	}

	err := module.ConfirmationHandler()(context.Background(), events.Event{
		Type:      events.TypeMessageCreate,
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		UserID:    "user-1",
		Content:   xpResetConfirmContent,
		CreatedAt: clock.now.Add(time.Second),
	})
	if err != nil {
		t.Fatalf("confirm handler: %v", err)
	}
	if _, ok := repo.VoiceProfiles["guild-1/user-2"]; ok {
		t.Fatal("guild voice profile was not deleted")
	}
	if _, ok := repo.VoiceProfiles["guild-2/user-3"]; !ok {
		t.Fatal("other guild voice profile was deleted")
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].Message.Embeds[0].Title != deleteEmoji+"成功刪除伺服器內所有語音經驗" {
		t.Fatalf("success response = %#v", sideEffects.Sent)
	}
}

func TestResetConfirmationReportsNoGuildData(t *testing.T) {
	sideEffects := fakediscord.NewSideEffects()
	clock := &xpResetTestClock{now: time.Unix(1000, 0)}
	module := NewResetModule(
		fakemongo.NewXPAdminRepository(),
		&fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{OwnerID: "user-1"}},
		sideEffects,
		nil,
		clock,
	)
	interaction := fakediscord.SlashInteractionWithOptions(XPResetCommandName, "聊天經驗重製", nil)
	interaction.ChannelID = "channel-1"
	if err := module.ResetHandler()(context.Background(), interaction, fakediscord.NewResponder()); err != nil {
		t.Fatalf("arm handler: %v", err)
	}

	err := module.ConfirmationHandler()(context.Background(), events.Event{
		Type:      events.TypeMessageCreate,
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		UserID:    "user-1",
		Content:   xpResetConfirmContent,
		CreatedAt: clock.now.Add(time.Second),
	})
	if err != nil {
		t.Fatalf("confirm handler: %v", err)
	}
	if len(sideEffects.Sent) != 1 || !strings.Contains(sideEffects.Sent[0].Message.Embeds[0].Title, "伺服器沒有任何聊天經驗的資料!") {
		t.Fatalf("missing guild response = %#v", sideEffects.Sent)
	}
}

func TestRewardRoleAddDeleteAndQueryHandlers(t *testing.T) {
	textRepo := fakemongo.NewTextXPRewardRoleRepository()
	voiceRepo := fakemongo.NewVoiceXPRewardRoleRepository()
	roles := fakediscord.NewSideEffects()
	roles.AssignableRoles["guild-1/role-1"] = true
	usage := &fakeusage.Tracker{}
	module := NewRewardRoleModule(textRepo, voiceRepo, roles, usage)

	add := fakediscord.SlashInteractionWithOptions(TextXPRewardRoleCommandName, "增加", map[string]string{
		"等級":     "12",
		"身分組":    "role-1",
		"是否自動刪除": "true",
	})
	add.Actor.PermissionBits = permissionManageMessages
	responder := fakediscord.NewResponder()
	if err := module.TextHandler()(context.Background(), add, responder); err != nil {
		t.Fatalf("add handler: %v", err)
	}
	if len(textRepo.Configs) != 1 || textRepo.Configs[0].Level != 12 || textRepo.Configs[0].RoleID != "role-1" || !textRepo.Configs[0].DeleteWhenNot {
		t.Fatalf("saved reward roles = %#v", textRepo.Configs)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != channelEmoji+"聊天經驗系統" || !strings.Contains(responder.Edits[0].Embeds[0].Description, "成功`增加`/`修改`該設定") {
		t.Fatalf("add response = %#v", responder.Edits)
	}

	query := fakediscord.SlashInteractionWithOptions(TextXPRewardRoleCommandName, "設定查詢", nil)
	query.Actor.PermissionBits = permissionManageMessages
	responder = fakediscord.NewResponder()
	if err := module.TextHandler()(context.Background(), query, responder); err != nil {
		t.Fatalf("query handler: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds[0].Fields) != 1 {
		t.Fatalf("query response = %#v", responder.Edits)
	}
	field := responder.Edits[0].Embeds[0].Fields[0]
	if !strings.Contains(field.Value, "**等級:**`12`") || !strings.Contains(field.Value, "**身分組:**<@&role-1>") {
		t.Fatalf("query field = %#v", field)
	}
	if responder.Edits[0].AllowedMentions == nil {
		t.Fatalf("allowed mentions should be explicit: %#v", responder.Edits[0])
	}

	del := fakediscord.SlashInteractionWithOptions(TextXPRewardRoleCommandName, "刪除", map[string]string{"等級": "12", "身分組": "role-1"})
	del.Actor.PermissionBits = permissionManageMessages
	responder = fakediscord.NewResponder()
	if err := module.TextHandler()(context.Background(), del, responder); err != nil {
		t.Fatalf("delete handler: %v", err)
	}
	if len(textRepo.Configs) != 0 {
		t.Fatalf("reward role was not deleted: %#v", textRepo.Configs)
	}
	if len(usage.Events) != 3 || usage.Events[0].Feature != "text-xp-role-config" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

type xpResetTestClock struct {
	now time.Time
}

func (c *xpResetTestClock) Now() time.Time {
	return c.now
}

func TestRewardRoleRejectsPermissionRoleAndMissingDelete(t *testing.T) {
	textRepo := fakemongo.NewTextXPRewardRoleRepository()
	module := NewRewardRoleModule(textRepo, fakemongo.NewVoiceXPRewardRoleRepository(), fakediscord.NewSideEffects(), nil)

	interaction := fakediscord.SlashInteractionWithOptions(TextXPRewardRoleCommandName, "增加", map[string]string{"等級": "1", "身分組": "role-1"})
	responder := fakediscord.NewResponder()
	if err := module.TextHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("permission handler: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Title, "你需要有`訊息管理`才能使用此指令") {
		t.Fatalf("permission response = %#v", responder.Edits)
	}

	interaction.Actor.PermissionBits = permissionManageMessages
	responder = fakediscord.NewResponder()
	if err := module.TextHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("role handler: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Title, "我沒有權限給大家這個身分組") {
		t.Fatalf("role response = %#v", responder.Edits)
	}

	deleteInteraction := fakediscord.SlashInteractionWithOptions(TextXPRewardRoleCommandName, "刪除", map[string]string{"等級": "1", "身分組": "role-1"})
	deleteInteraction.Actor.PermissionBits = permissionManageMessages
	responder = fakediscord.NewResponder()
	if err := module.TextHandler()(context.Background(), deleteInteraction, responder); err != nil {
		t.Fatalf("missing delete handler: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Title, "你沒有設定過這個選項!") {
		t.Fatalf("missing delete response = %#v", responder.Edits)
	}
}

func TestRewardRoleVoicePaginationUpdatesMessage(t *testing.T) {
	voiceRepo := fakemongo.NewVoiceXPRewardRoleRepository()
	for i := 0; i < 13; i++ {
		voiceRepo.Configs = append(voiceRepo.Configs, domain.XPRewardRoleConfig{GuildID: "guild-1", Level: int64(i + 1), RoleID: "role"})
	}
	module := NewRewardRoleModule(fakemongo.NewTextXPRewardRoleRepository(), voiceRepo, nil, nil)
	module.color = func() int { return 0x123456 }
	interaction := fakediscord.ComponentInteractionFromID("1voice_leave_role")
	responder := fakediscord.NewResponder()
	if err := module.VoicePageHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("page handler: %v", err)
	}
	if len(responder.Updates) != 1 || len(responder.Updates[0].Embeds[0].Fields) != 1 {
		t.Fatalf("updates = %#v", responder.Updates)
	}
	if responder.Updates[0].Embeds[0].Color != 0x123456 {
		t.Fatalf("color = %#v", responder.Updates[0].Embeds[0])
	}
	if responder.Updates[0].Components[0].Components[0].CustomID != "0voice_leave_role" || !responder.Updates[0].Components[0].Components[1].Disabled {
		t.Fatalf("components = %#v", responder.Updates[0].Components)
	}
}

func TestRewardRoleListPreservesLegacySixthFieldIndexing(t *testing.T) {
	configs := make([]domain.XPRewardRoleConfig, 13)
	for index := range configs {
		configs[index] = domain.XPRewardRoleConfig{GuildID: "guild-1", Level: int64(index + 1), RoleID: fmt.Sprintf("role-%d", index+1), DeleteWhenNot: index == 5}
	}

	fields := rewardRoleListMessage("聊天", configs[:12], 0, true, 0x123456).Embeds[0].Fields
	if len(fields) != 11 {
		t.Fatalf("12-row fields = %#v", fields)
	}
	for _, field := range fields {
		if field.Name == "第6個:" {
			t.Fatalf("legacy field six should be absent without a next page: %#v", fields)
		}
	}

	fields = rewardRoleListMessage("聊天", configs, 0, true, 0x123456).Embeds[0].Fields
	if len(fields) != 12 || fields[5].Name != "第6個:" {
		t.Fatalf("13-row fields = %#v", fields)
	}
	if !strings.Contains(fields[5].Value, "**等級:**`13`") || !strings.Contains(fields[5].Value, "**身分組:**<@&role-6>") || !strings.HasSuffix(fields[5].Value, "true") {
		t.Fatalf("legacy mixed sixth field = %#v", fields[5])
	}
}

func TestRewardRoleQueryDisplaysThenDeletesMissingCachedRoles(t *testing.T) {
	for _, tc := range []struct {
		name string
		text bool
	}{
		{name: "text", text: true},
		{name: "voice", text: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			textRepo := fakemongo.NewTextXPRewardRoleRepository()
			voiceRepo := fakemongo.NewVoiceXPRewardRoleRepository()
			configs := []domain.XPRewardRoleConfig{
				{GuildID: "guild-1", Level: 1, RoleID: "missing-role"},
				{GuildID: "guild-1", Level: 2, RoleID: "live-role"},
			}
			if tc.text {
				textRepo.Configs = append(textRepo.Configs, configs...)
			} else {
				voiceRepo.Configs = append(voiceRepo.Configs, configs...)
			}
			roles := fakediscord.NewSideEffects()
			roles.MissingRoles["guild-1/missing-role"] = true
			module := NewRewardRoleModule(textRepo, voiceRepo, roles, nil)
			module.color = func() int { return 0x123456 }
			commandName := TextXPRewardRoleCommandName
			handler := module.TextHandler()
			if !tc.text {
				commandName = VoiceXPRewardRoleCommandName
				handler = module.VoiceHandler()
			}
			interaction := fakediscord.SlashInteractionWithOptions(commandName, "設定查詢", nil)
			interaction.Actor.PermissionBits = permissionManageMessages

			responder := fakediscord.NewResponder()
			if err := handler(context.Background(), interaction, responder); err != nil {
				t.Fatalf("first query: %v", err)
			}
			if fields := responder.Edits[0].Embeds[0].Fields; len(fields) != 2 || !strings.Contains(fields[0].Value, "missing-role") {
				t.Fatalf("first query fields = %#v", fields)
			}
			remaining := textRepo.Configs
			if !tc.text {
				remaining = voiceRepo.Configs
			}
			if len(remaining) != 1 || remaining[0].RoleID != "live-role" {
				t.Fatalf("remaining configs = %#v", remaining)
			}

			responder = fakediscord.NewResponder()
			if err := handler(context.Background(), interaction, responder); err != nil {
				t.Fatalf("second query: %v", err)
			}
			if fields := responder.Edits[0].Embeds[0].Fields; len(fields) != 1 || strings.Contains(fields[0].Value, "missing-role") {
				t.Fatalf("second query fields = %#v", fields)
			}
		})
	}
}
