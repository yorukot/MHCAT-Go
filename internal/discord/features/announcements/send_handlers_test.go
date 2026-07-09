package announcements

import (
	"context"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestSendHandlerShowsLegacyModal(t *testing.T) {
	module := NewSendModule(fakemongo.NewAnnouncementConfigRepository(), fakediscord.NewSideEffects(), nil)
	interaction := sendInteraction()
	responder := fakediscord.NewResponder()

	if err := module.SendHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Modals) != 1 {
		t.Fatalf("modals = %#v", responder.Modals)
	}
	modal := responder.Modals[0]
	if modal.Title != "公告系統" || !strings.HasPrefix(modal.CustomID, "mhcat:v1:announcement:submit:") {
		t.Fatalf("modal = %#v", modal)
	}
	wantLabels := []string{"請輸入你要tag誰", "請輸入你的公告要甚麼顏色", "請輸入你的公告標題", "請輸入公告內文"}
	for index, want := range wantLabels {
		if got := modal.Rows[index].Inputs[0].Label; got != want {
			t.Fatalf("label %d = %q want %q", index, got, want)
		}
	}
}

func TestSendHandlerRequiresManageMessages(t *testing.T) {
	module := NewSendModule(fakemongo.NewAnnouncementConfigRepository(), fakediscord.NewSideEffects(), nil)
	interaction := fakediscord.SlashInteraction(SendCommandName)
	responder := fakediscord.NewResponder()

	if err := module.SendHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Replies) != 1 || !strings.Contains(responder.Replies[0].Embeds[0].Title, "你需要有`訊息管理`才能使用此指令") {
		t.Fatalf("permission response = %#v", responder.Replies)
	}
}

func TestSendModalPreviewsAndConfirmsAnnouncement(t *testing.T) {
	repo := fakemongo.NewAnnouncementConfigRepository()
	repo.AnnouncementChannels["guild-1"] = "announcement-channel"
	sideEffects := fakediscord.NewSideEffects()
	module := NewSendModule(repo, sideEffects, nil)
	interaction := modalInteraction(map[string]string{
		fieldTag:     "@everyone",
		fieldColor:   "#53FF53",
		fieldTitle:   "重要公告",
		fieldContent: "今天維護",
	})
	responder := fakediscord.NewResponder()

	if err := module.SendModalHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("modal handler: %v", err)
	}
	if len(responder.Defers) != 1 || len(responder.Edits) != 1 || len(responder.Follow) != 1 {
		t.Fatalf("defers=%#v edits=%#v follow=%#v", responder.Defers, responder.Edits, responder.Follow)
	}
	preview := responder.Edits[0]
	if preview.Content != "@everyone" || preview.AllowedMentions == nil || len(preview.Embeds) != 1 {
		t.Fatalf("preview = %#v", preview)
	}
	if preview.Embeds[0].Title != "重要公告" || preview.Embeds[0].Description != "今天維護" || preview.Embeds[0].Footer == nil || preview.Embeds[0].Footer.Text != "來自tester#0001的公告" {
		t.Fatalf("preview embed = %#v", preview.Embeds[0])
	}
	confirm := responder.Follow[0]
	if len(confirm.Components) != 1 || len(confirm.Components[0].Components) != 2 {
		t.Fatalf("confirm components = %#v", confirm.Components)
	}
	if confirm.Embeds[0].Title != "是否將此訊息送往公告?(請於六秒內點擊:P)" {
		t.Fatalf("confirm embed = %#v", confirm.Embeds)
	}
	confirmID := confirm.Components[0].Components[0].CustomID
	if !strings.HasPrefix(confirmID, "mhcat:v1:announcement:confirm:state=") {
		t.Fatalf("confirm id = %q", confirmID)
	}
	confirmInteraction := fakediscord.ComponentInteractionFromID(confirmID)
	confirmInteraction.Actor.UserID = "user-1"
	confirmInteraction.Actor.GuildID = "guild-1"
	confirmResponder := fakediscord.NewResponder()
	if err := module.ConfirmHandler()(context.Background(), confirmInteraction, confirmResponder); err != nil {
		t.Fatalf("confirm handler: %v", err)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].ChannelID != "announcement-channel" {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
	sent := sideEffects.Sent[0].Message
	if sent.Content != "@everyone" || len(sent.AllowedMentions.UserIDs) != 0 || sent.AllowedMentions.ParseEveryone {
		t.Fatalf("sent allowed mentions/content = %#v", sent)
	}
	if len(confirmResponder.Edits) != 1 || confirmResponder.Edits[0].Content != "<a:green_tick:994529015652163614> | 成功發送!" {
		t.Fatalf("confirm response = %#v", confirmResponder.Edits)
	}
}

func TestSendModalRejectsInvalidColor(t *testing.T) {
	module := NewSendModule(fakemongo.NewAnnouncementConfigRepository(), fakediscord.NewSideEffects(), nil)
	interaction := modalInteraction(map[string]string{
		fieldTag:     "@here",
		fieldColor:   "Random",
		fieldTitle:   "公告",
		fieldContent: "內容",
	})
	responder := fakediscord.NewResponder()

	if err := module.SendModalHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || responder.Edits[0].Embeds[0].Title != "你傳送的並不是顏色(色碼)" {
		t.Fatalf("color response = %#v", responder.Edits)
	}
}

func TestConfirmMissingAnnouncementChannel(t *testing.T) {
	module := NewSendModule(fakemongo.NewAnnouncementConfigRepository(), fakediscord.NewSideEffects(), nil)
	stateID, err := module.draftStore().Put(AnnouncementDraft{
		GuildID: "guild-1",
		UserID:  "user-1",
		Tag:     "@here",
		Color:   0x53FF53,
		Title:   "公告",
		Content: "內容",
	})
	if err != nil {
		t.Fatalf("put draft: %v", err)
	}
	interaction := fakediscord.ComponentInteractionFromID(announcementStateComponentID(confirmAction, stateID))
	interaction.Actor.UserID = "user-1"
	interaction.Actor.GuildID = "guild-1"
	responder := fakediscord.NewResponder()

	if err := module.ConfirmHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Content, "你還沒有對您的公告頻道進行選擇") {
		t.Fatalf("missing response = %#v", responder.Edits)
	}
}

func TestCancelDropsDraft(t *testing.T) {
	module := NewSendModule(fakemongo.NewAnnouncementConfigRepository(), fakediscord.NewSideEffects(), nil)
	stateID, err := module.draftStore().Put(AnnouncementDraft{GuildID: "guild-1", UserID: "user-1", Color: 0x53FF53, Title: "公告", Content: "內容", Tag: "@here"})
	if err != nil {
		t.Fatalf("put draft: %v", err)
	}
	interaction := fakediscord.ComponentInteractionFromID(announcementStateComponentID(cancelAction, stateID))
	responder := fakediscord.NewResponder()
	if err := module.CancelHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if len(responder.Replies) != 1 || responder.Replies[0].Content != "已取消" {
		t.Fatalf("cancel response = %#v", responder.Replies)
	}
	if _, err := module.draftStore().Take(stateID); err != ErrAnnouncementDraftNotFound {
		t.Fatalf("draft should be gone, got %v", err)
	}
}

func TestLegacyAnnouncementModalStillRoutesToSubmit(t *testing.T) {
	id, err := customid.ParseModal(sendModalLegacyID, []customid.ModalField{{CustomID: fieldTag, Value: "@here"}})
	if err != nil {
		t.Fatalf("parse legacy modal: %v", err)
	}
	if id.Feature != announcementFeature || id.Action != sendModalAction || !id.Legacy {
		t.Fatalf("legacy id = %#v", id)
	}
}

func sendInteraction() interactions.Interaction {
	interaction := fakediscord.SlashInteraction(SendCommandName)
	interaction.Actor.PermissionBits = permissionManageMessages
	interaction.Actor.UserTag = "tester#0001"
	interaction.Actor.AvatarURL = "https://example.invalid/avatar.png"
	return interaction
}

func modalInteraction(fields map[string]string) interactions.Interaction {
	interaction := fakediscord.ModalInteraction(interactions.ModalKey{Version: customid.VersionV1, Feature: announcementFeature, Action: sendModalAction})
	interaction.CustomID = "mhcat:v1:announcement:submit:"
	interaction.Actor.UserID = "user-1"
	interaction.Actor.GuildID = "guild-1"
	interaction.Actor.UserTag = "tester#0001"
	interaction.Actor.AvatarURL = "https://example.invalid/avatar.png"
	for _, key := range []string{fieldTag, fieldColor, fieldTitle, fieldContent} {
		interaction.ModalFields = append(interaction.ModalFields, customid.ModalField{CustomID: key, Value: fields[key]})
	}
	return interaction
}
