package ticket

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

const (
	testGuildID    = "111111111111111111"
	testCategoryID = "222222222222222222"
	testAdminRole  = "333333333333333333"
)

func TestSetupHandlerShowsLegacyModalWithoutSavingConfig(t *testing.T) {
	repo := fakemongo.NewTicketConfigRepository()
	module := NewModule(repo, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions("私人頻道設置", "", map[string]string{
		"類別":     testCategoryID,
		"管理員身分組": testAdminRole,
	})
	interaction.Actor.GuildID = testGuildID
	interaction.Actor.PermissionBits = permissionManageMessages

	if err := module.SetupHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("setup handler: %v", err)
	}
	if len(responder.Modals) != 1 {
		t.Fatalf("modals = %#v", responder.Modals)
	}
	modal := responder.Modals[0]
	if modal.Title != "私人頻道系統!" || !strings.HasPrefix(modal.CustomID, "mhcat:v1:ticket:setup:") {
		t.Fatalf("modal = %#v", modal)
	}
	gotInputs := flattenInputs(modal)
	wantInputs := []responses.TextInput{
		{CustomID: "ticketcolor", Label: "請輸入嵌入顏色", Style: responses.TextInputStyleShort, Required: true},
		{CustomID: "tickettitle", Label: "請輸入標題", Style: responses.TextInputStyleShort, Required: true},
		{CustomID: "ticketcontent", Label: "請輸入內文", Style: responses.TextInputStyleParagraph, Required: true},
	}
	if len(gotInputs) != len(wantInputs) {
		t.Fatalf("inputs = %#v", gotInputs)
	}
	for index, want := range wantInputs {
		got := gotInputs[index]
		if got.CustomID != want.CustomID || got.Label != want.Label || got.Style != want.Style || got.Required != want.Required {
			t.Fatalf("input %d = %#v, want %#v", index, got, want)
		}
	}
	if _, err := repo.GetTicketConfig(context.Background(), testGuildID); !errors.Is(err, ports.ErrTicketConfigNotFound) {
		t.Fatalf("config saved before modal submit: %v", err)
	}
}

func TestSetupHandlerRequiresManageMessagesWithLegacyError(t *testing.T) {
	module := NewModule(fakemongo.NewTicketConfigRepository(), nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions("私人頻道設置", "", map[string]string{
		"類別":     testCategoryID,
		"管理員身分組": testAdminRole,
	})
	interaction.Actor.GuildID = testGuildID

	if err := module.SetupHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("setup handler: %v", err)
	}
	if len(responder.Replies) != 1 || !responder.Replies[0].Ephemeral {
		t.Fatalf("permission reply = %#v", responder.Replies)
	}
	if got := responder.Replies[0].Embeds[0].Title; got != "<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`訊息管理`才能使用此指令" {
		t.Fatalf("permission title = %q", got)
	}
	if len(responder.Modals) != 0 {
		t.Fatalf("unexpected modals = %#v", responder.Modals)
	}
}

func TestSetupHandlerExistingConfigUsesLegacyDuplicateError(t *testing.T) {
	repo := seededTicketRepo(t)
	module := NewModule(repo, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions("私人頻道設置", "", map[string]string{
		"類別":     "444444444444444444",
		"管理員身分組": "555555555555555555",
	})
	interaction.Actor.GuildID = testGuildID
	interaction.Actor.PermissionBits = permissionManageMessages

	if err := module.SetupHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("setup handler: %v", err)
	}
	if len(responder.Replies) != 1 || !responder.Replies[0].Ephemeral {
		t.Fatalf("duplicate reply = %#v", responder.Replies)
	}
	embed := responder.Replies[0].Embeds[0]
	if embed.Title != "__**錯誤**__" || !strings.Contains(embed.Description, "`<>h 刪除私人頻道`") {
		t.Fatalf("duplicate embed = %#v", embed)
	}
	if len(responder.Modals) != 0 {
		t.Fatalf("unexpected modals = %#v", responder.Modals)
	}
	config, err := repo.GetTicketConfig(context.Background(), testGuildID)
	if err != nil {
		t.Fatalf("get config: %v", err)
	}
	if config.CategoryID != testCategoryID || config.AdminRoleID != testAdminRole {
		t.Fatalf("config was overwritten: %#v", config)
	}
}

func TestSetupModalSavesConfigAfterValidModalAndRepliesPanel(t *testing.T) {
	repo := fakemongo.NewTicketConfigRepository()
	sideEffects := fakediscord.NewSideEffects()
	usage := &fakeusage.Tracker{}
	module := NewModuleWithSideEffects(repo, usage, nil, sideEffects, "")
	responder := fakediscord.NewResponder()
	interaction := ticketModalInteraction(t, "#00ff00", "建立私人頻道", "按下按鈕創建客服頻道")

	if err := module.SetupModalHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("setup modal handler: %v", err)
	}
	config, err := repo.GetTicketConfig(context.Background(), testGuildID)
	if err != nil {
		t.Fatalf("get saved config: %v", err)
	}
	wantConfig := domain.TicketConfig{
		GuildID:        testGuildID,
		CategoryID:     testCategoryID,
		AdminRoleID:    testAdminRole,
		EveryoneRoleID: testGuildID,
	}
	if config != wantConfig {
		t.Fatalf("config = %#v, want %#v", config, wantConfig)
	}
	if len(responder.Defers) != 1 || len(responder.Edits) != 1 {
		t.Fatalf("defers=%#v edits=%#v replies=%#v", responder.Defers, responder.Edits, responder.Replies)
	}
	if len(responder.Edits[0].Embeds) != 1 || responder.Edits[0].Embeds[0].Title != "<a:green_tick:994529015652163614> | 成功創建私人頻道" {
		t.Fatalf("success edit = %#v", responder.Edits)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].ChannelID != "panel-channel" {
		t.Fatalf("sent messages = %#v", sideEffects.Sent)
	}
	sent := sideEffects.Sent[0].Message
	if len(sent.Embeds) != 1 || sent.Embeds[0].Title != "建立私人頻道" || sent.Embeds[0].Description != "按下按鈕創建客服頻道" || sent.Embeds[0].Color != 0x00ff00 {
		t.Fatalf("panel embed = %#v", sent.Embeds)
	}
	if len(sent.Components) != 1 || len(sent.Components[0].Components) != 1 {
		t.Fatalf("panel components = %#v", sent.Components)
	}
	button := sent.Components[0].Components[0]
	if button.CustomID != "tic" || button.Label != "🎫 點我創建客服頻道!" || button.Style != "primary" {
		t.Fatalf("panel button = %#v", button)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != "私人頻道設置" || usage.Events[0].Feature != "ticket" {
		t.Fatalf("usage events = %#v", usage.Events)
	}
}

func TestSetupModalRejectsInvalidColorWithoutSavingConfig(t *testing.T) {
	repo := fakemongo.NewTicketConfigRepository()
	sideEffects := fakediscord.NewSideEffects()
	module := NewModuleWithSideEffects(repo, nil, nil, sideEffects, "")
	responder := fakediscord.NewResponder()
	interaction := ticketModalInteraction(t, "not-a-color", "建立私人頻道", "按下按鈕創建客服頻道")

	if err := module.SetupModalHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("setup modal handler: %v", err)
	}
	if _, err := repo.GetTicketConfig(context.Background(), testGuildID); !errors.Is(err, ports.ErrTicketConfigNotFound) {
		t.Fatalf("config saved after invalid modal: %v", err)
	}
	if len(sideEffects.Sent) != 0 {
		t.Fatalf("sent messages = %#v", sideEffects.Sent)
	}
	if len(responder.Defers) != 1 || len(responder.Edits) != 1 {
		t.Fatalf("invalid color defers=%#v edits=%#v replies=%#v", responder.Defers, responder.Edits, responder.Replies)
	}
	if len(responder.Edits[0].Embeds) != 1 || responder.Edits[0].Embeds[0].Title != "你傳送的並不是顏色(色碼)" {
		t.Fatalf("invalid color embed = %#v", responder.Edits[0].Embeds)
	}
}

func TestLegacyPanelSubmitSendsPanelAndEditsSuccess(t *testing.T) {
	sideEffects := fakediscord.NewSideEffects()
	usage := &fakeusage.Tracker{}
	module := NewModuleWithSideEffects(nil, usage, nil, sideEffects, "")
	responder := fakediscord.NewResponder()
	interaction := legacyTicketModalInteraction("#00ff00", "建立私人頻道", "按下按鈕創建客服頻道")

	if err := module.LegacyPanelSubmitHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("legacy panel submit: %v", err)
	}
	if len(responder.Defers) != 1 || len(responder.Edits) != 1 {
		t.Fatalf("defers=%#v edits=%#v", responder.Defers, responder.Edits)
	}
	if len(responder.Edits[0].Embeds) != 1 || responder.Edits[0].Embeds[0].Title != "<a:green_tick:994529015652163614> | 成功創建私人頻道" {
		t.Fatalf("success edit = %#v", responder.Edits)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].ChannelID != "panel-channel" {
		t.Fatalf("sent messages = %#v", sideEffects.Sent)
	}
	sent := sideEffects.Sent[0].Message
	if len(sent.Embeds) != 1 || sent.Embeds[0].Title != "建立私人頻道" || sent.Embeds[0].Description != "按下按鈕創建客服頻道" || sent.Embeds[0].Color != 0x00ff00 {
		t.Fatalf("panel embed = %#v", sent.Embeds)
	}
	if len(sent.Components) != 1 || sent.Components[0].Components[0].CustomID != "tic" {
		t.Fatalf("panel components = %#v", sent.Components)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != "私人頻道設置" {
		t.Fatalf("usage events = %#v", usage.Events)
	}
}

func TestLegacyPanelSubmitRejectsInvalidColor(t *testing.T) {
	sideEffects := fakediscord.NewSideEffects()
	module := NewModuleWithSideEffects(nil, nil, nil, sideEffects, "")
	responder := fakediscord.NewResponder()
	interaction := legacyTicketModalInteraction("bad", "建立私人頻道", "按下按鈕創建客服頻道")

	if err := module.LegacyPanelSubmitHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("legacy panel submit: %v", err)
	}
	if len(sideEffects.Sent) != 0 {
		t.Fatalf("sent messages = %#v", sideEffects.Sent)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "你傳送的並不是顏色(色碼)" {
		t.Fatalf("invalid color edit = %#v", responder.Edits)
	}
}

func TestParseLegacyColorMatchesHTMLColorNames(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{name: "basic name", in: "red", want: 0xFF0000},
		{name: "legacy validate-color name", in: "aqua", want: 0x00FFFF},
		{name: "modern css name", in: "rebeccapurple", want: 0x663399},
		{name: "case insensitive", in: "DarkSlateGray", want: 0x2F4F4F},
		{name: "grey alias", in: "lightgrey", want: 0xD3D3D3},
		{name: "three digit hex", in: "#0f0", want: 0x00FF00},
		{name: "six digit hex", in: "#00ff00", want: 0x00FF00},
		{name: "uppercase hex", in: "#ABCDEF", want: 0xABCDEF},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseLegacyColor(tt.in)
			if !ok {
				t.Fatalf("parseLegacyColor(%q) rejected", tt.in)
			}
			if got != tt.want {
				t.Fatalf("parseLegacyColor(%q) = %#06x, want %#06x", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseLegacyColorRejectsUnsupportedForms(t *testing.T) {
	for _, value := range []string{"", "not-a-color"} {
		t.Run(value, func(t *testing.T) {
			if got, ok := parseLegacyColor(value); ok {
				t.Fatalf("parseLegacyColor(%q) = %#06x, want reject", value, got)
			}
		})
	}
}

func TestDeleteHandlerDeletesConfigWithLegacyMessage(t *testing.T) {
	repo := fakemongo.NewTicketConfigRepository()
	if err := repo.SaveTicketConfig(context.Background(), domain.TicketConfig{
		GuildID:        testGuildID,
		CategoryID:     testCategoryID,
		AdminRoleID:    testAdminRole,
		EveryoneRoleID: testGuildID,
	}); err != nil {
		t.Fatalf("seed config: %v", err)
	}
	usage := &fakeusage.Tracker{}
	module := NewModule(repo, usage)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction("私人頻道刪除")
	interaction.Actor.GuildID = testGuildID
	interaction.Actor.PermissionBits = permissionManageMessages

	if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("delete handler: %v", err)
	}
	if _, err := repo.GetTicketConfig(context.Background(), testGuildID); !errors.Is(err, ports.ErrTicketConfigNotFound) {
		t.Fatalf("config still exists after delete: %v", err)
	}
	if len(responder.Defers) != 1 || len(responder.Edits) != 1 {
		t.Fatalf("defers=%#v edits=%#v", responder.Defers, responder.Edits)
	}
	description := responder.Edits[0].Embeds[0].Description
	if !strings.Contains(description, "成功刪除私人頻道的設置") {
		t.Fatalf("delete description = %q", description)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != "私人頻道刪除" {
		t.Fatalf("usage events = %#v", usage.Events)
	}
}

func TestDeleteHandlerMissingConfigUsesLegacyFailureMessage(t *testing.T) {
	module := NewModule(fakemongo.NewTicketConfigRepository(), nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction("私人頻道刪除")
	interaction.Actor.GuildID = testGuildID
	interaction.Actor.PermissionBits = permissionManageMessages

	if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("delete handler: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	description := responder.Edits[0].Embeds[0].Description
	if !strings.Contains(description, "你還沒有創建私人頻道的設定") {
		t.Fatalf("delete missing description = %q", description)
	}
}

func TestDeleteHandlerRequiresManageMessagesBeforeDeletingConfig(t *testing.T) {
	repo := seededTicketRepo(t)
	usage := &fakeusage.Tracker{}
	module := NewModule(repo, usage)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction("私人頻道刪除")
	interaction.Actor.GuildID = testGuildID

	if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("delete handler: %v", err)
	}
	config, err := repo.GetTicketConfig(context.Background(), testGuildID)
	if err != nil {
		t.Fatalf("config was deleted without permission: %v", err)
	}
	if config.CategoryID != testCategoryID || config.AdminRoleID != testAdminRole {
		t.Fatalf("config changed without permission: %#v", config)
	}
	if len(responder.Defers) != 1 || responder.Defers[0].Ephemeral {
		t.Fatalf("permission defers = %#v", responder.Defers)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Ephemeral {
		t.Fatalf("permission edits = %#v", responder.Edits)
	}
	wantTitle := "<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`訊息管理`才能使用此指令"
	if got := responder.Edits[0].Embeds[0].Title; got != wantTitle {
		t.Fatalf("permission title = %q, want %q", got, wantTitle)
	}
	if len(usage.Events) != 0 {
		t.Fatalf("route-level usage events = %#v", usage.Events)
	}
}

func TestOpenHandlerCreatesTicketChannelAndSendsLegacyWelcome(t *testing.T) {
	repo := seededTicketRepo(t)
	sideEffects := fakediscord.NewSideEffects()
	usage := &fakeusage.Tracker{}
	module := NewModuleWithSideEffects(repo, usage, sideEffects, sideEffects, "")
	responder := fakediscord.NewResponder()
	interaction := ticketButtonInteraction("tic")
	interaction.ApplicationID = "444444444444444444"

	if err := module.OpenHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("open handler: %v", err)
	}
	if len(sideEffects.Created) != 1 {
		t.Fatalf("created channels = %#v", sideEffects.Created)
	}
	created := sideEffects.Created[0]
	if created.GuildID != testGuildID || created.ParentID != testCategoryID || created.Name != interaction.Actor.UserID || created.Type != discordChannelTypeGuildText {
		t.Fatalf("created channel request = %#v", created)
	}
	if len(created.PermissionOverwrites) != 4 {
		t.Fatalf("permission overwrites = %#v", created.PermissionOverwrites)
	}
	wantAllow := int64(permissionViewChannel | permissionSendMessages | permissionReadMessageHistory)
	if created.PermissionOverwrites[0].ID != testAdminRole || created.PermissionOverwrites[0].Type != permissionOverwriteRole || created.PermissionOverwrites[0].Allow != wantAllow {
		t.Fatalf("admin overwrite = %#v", created.PermissionOverwrites[0])
	}
	if created.PermissionOverwrites[1].ID != testGuildID || created.PermissionOverwrites[1].Deny != int64(permissionViewChannel) {
		t.Fatalf("everyone overwrite = %#v", created.PermissionOverwrites[1])
	}
	if created.PermissionOverwrites[3].ID != interaction.ApplicationID || created.PermissionOverwrites[3].Type != permissionOverwriteMember || created.PermissionOverwrites[3].Allow != wantAllow {
		t.Fatalf("bot overwrite = %#v", created.PermissionOverwrites[3])
	}
	if len(sideEffects.Sent) != 1 {
		t.Fatalf("sent messages = %#v", sideEffects.Sent)
	}
	sent := sideEffects.Sent[0].Message
	if sent.Content != "||@everyone||" || sent.AllowedMentions.ParseEveryone {
		t.Fatalf("welcome content/mentions = %#v", sent)
	}
	if len(sent.Embeds) != 1 || sent.Embeds[0].Title != "__**私人頻道**__" || sent.Embeds[0].Description != "你開啟了一個私人頻道，請等待客服人員的回復!" {
		t.Fatalf("welcome embeds = %#v", sent.Embeds)
	}
	if len(sent.Components) != 1 || len(sent.Components[0].Components) != 1 || sent.Components[0].Components[0].CustomID != "del" {
		t.Fatalf("welcome components = %#v", sent.Components)
	}
	if len(responder.Replies) != 1 || !responder.Replies[0].Ephemeral {
		t.Fatalf("open reply = %#v", responder.Replies)
	}
	if len(responder.Replies[0].Embeds) != 1 || responder.Replies[0].Embeds[0].Title != "__**頻道**__" || responder.Replies[0].Embeds[0].Description != ":white_check_mark: 你成功開啟了頻道!" {
		t.Fatalf("open success embed = %#v", responder.Replies[0].Embeds)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != "私人頻道開啟" {
		t.Fatalf("usage events = %#v", usage.Events)
	}
}

func TestTicketBotUserIDUsesConfiguredFallbackOutsideRuntimeInteractions(t *testing.T) {
	if got := ticketBotUserID("", " fallback-bot "); got != "fallback-bot" {
		t.Fatalf("ticket bot user id = %q, want fallback-bot", got)
	}
}

func TestOpenHandlerDuplicateChannelUsesLegacyWarning(t *testing.T) {
	repo := seededTicketRepo(t)
	sideEffects := fakediscord.NewSideEffects()
	sideEffects.Channels = append(sideEffects.Channels, ports.ChannelRef{
		GuildID:   testGuildID,
		ChannelID: "existing-channel",
		Name:      "user-1",
		Type:      discordChannelTypeGuildText,
	})
	module := NewModuleWithSideEffects(repo, nil, sideEffects, sideEffects, "")
	responder := fakediscord.NewResponder()

	if err := module.OpenHandler()(context.Background(), ticketButtonInteraction("tic"), responder); err != nil {
		t.Fatalf("open handler: %v", err)
	}
	if len(sideEffects.Created) != 0 || len(sideEffects.Sent) != 0 {
		t.Fatalf("unexpected side effects: created=%#v sent=%#v", sideEffects.Created, sideEffects.Sent)
	}
	if len(responder.Replies) != 1 || !responder.Replies[0].Ephemeral || responder.Replies[0].Embeds[0].Description != ":warning: 你已經有一個客服頻道了!" {
		t.Fatalf("duplicate reply = %#v", responder.Replies)
	}
}

func TestOpenHandlerDuplicateCheckMatchesAnyLegacyChannelType(t *testing.T) {
	repo := seededTicketRepo(t)
	sideEffects := fakediscord.NewSideEffects()
	sideEffects.Channels = append(sideEffects.Channels, ports.ChannelRef{
		GuildID:   testGuildID,
		ChannelID: "same-name-voice-channel",
		Name:      "user-1",
		Type:      2,
	})
	module := NewModuleWithSideEffects(repo, nil, sideEffects, sideEffects, "")
	responder := fakediscord.NewResponder()

	if err := module.OpenHandler()(context.Background(), ticketButtonInteraction("tic"), responder); err != nil {
		t.Fatalf("open handler: %v", err)
	}
	if len(sideEffects.Created) != 0 || len(sideEffects.Sent) != 0 {
		t.Fatalf("unexpected side effects: created=%#v sent=%#v", sideEffects.Created, sideEffects.Sent)
	}
	if len(responder.Replies) != 1 || responder.Replies[0].Embeds[0].Description != ":warning: 你已經有一個客服頻道了!" {
		t.Fatalf("duplicate reply = %#v", responder.Replies)
	}
}

func TestOpenHandlerMissingConfigDeletesPanelAndRepliesLegacyMessage(t *testing.T) {
	sideEffects := fakediscord.NewSideEffects()
	module := NewModuleWithSideEffects(fakemongo.NewTicketConfigRepository(), nil, sideEffects, sideEffects, "")
	responder := fakediscord.NewResponder()
	interaction := ticketButtonInteraction("tic")

	if err := module.OpenHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("open handler: %v", err)
	}
	if len(sideEffects.DeletedMessage) != 1 || sideEffects.DeletedMessage[0].MessageID != "panel-message" {
		t.Fatalf("deleted panel messages = %#v", sideEffects.DeletedMessage)
	}
	if len(sideEffects.Created) != 0 || len(sideEffects.Sent) != 0 {
		t.Fatalf("unexpected side effects: created=%#v sent=%#v", sideEffects.Created, sideEffects.Sent)
	}
	if len(responder.Replies) != 1 || responder.Replies[0].Content != ":x: 這個創建私人頻道的設置已經被刪除了喔，請麻煩管理員重新創建!" {
		t.Fatalf("missing config reply = %#v", responder.Replies)
	}
}

func TestCloseHandlerDeletesOwnerTicketChannel(t *testing.T) {
	sideEffects := fakediscord.NewSideEffects()
	module := NewModuleWithSideEffects(nil, nil, sideEffects, sideEffects, "")
	interaction := ticketButtonInteraction("del")
	interaction.ChannelID = "ticket-channel"
	interaction.ChannelName = interaction.Actor.UserID

	if err := module.CloseHandler()(context.Background(), interaction, fakediscord.NewResponder()); err != nil {
		t.Fatalf("close handler: %v", err)
	}
	if len(sideEffects.Deleted) != 1 || sideEffects.Deleted[0] != "ticket-channel" {
		t.Fatalf("deleted channels = %#v", sideEffects.Deleted)
	}
}

func TestCloseHandlerDeletesWhenActorCanManageMessages(t *testing.T) {
	sideEffects := fakediscord.NewSideEffects()
	module := NewModuleWithSideEffects(nil, nil, sideEffects, sideEffects, "")
	interaction := ticketButtonInteraction("del")
	interaction.ChannelID = "ticket-channel"
	interaction.ChannelName = "someone-else"
	interaction.Actor.PermissionBits = permissionManageMessages

	if err := module.CloseHandler()(context.Background(), interaction, fakediscord.NewResponder()); err != nil {
		t.Fatalf("close handler: %v", err)
	}
	if len(sideEffects.Deleted) != 1 || sideEffects.Deleted[0] != "ticket-channel" {
		t.Fatalf("deleted channels = %#v", sideEffects.Deleted)
	}
}

func TestCloseHandlerDeniesWrongChannelWithoutManageMessages(t *testing.T) {
	sideEffects := fakediscord.NewSideEffects()
	module := NewModuleWithSideEffects(nil, nil, sideEffects, sideEffects, "")
	responder := fakediscord.NewResponder()
	interaction := ticketButtonInteraction("del")
	interaction.ChannelID = "ticket-channel"
	interaction.ChannelName = "someone-else"

	if err := module.CloseHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("close handler: %v", err)
	}
	if len(sideEffects.Deleted) != 0 {
		t.Fatalf("deleted channels = %#v", sideEffects.Deleted)
	}
	if len(responder.Replies) != 1 || responder.Replies[0].Embeds[0].Description != "你開啟了一個私人頻道，請靜候客服人員的回復!" {
		t.Fatalf("denied reply = %#v", responder.Replies)
	}
}

func ticketModalInteraction(t *testing.T, color string, title string, content string) interactions.Interaction {
	t.Helper()
	payload, err := customid.KeyValuePayload(map[string]string{"c": testCategoryID, "r": testAdminRole})
	if err != nil {
		t.Fatalf("payload: %v", err)
	}
	modalID, err := customid.Encode(customid.InteractionKindModal, "ticket", "setup", payload)
	if err != nil {
		t.Fatalf("encode modal id: %v", err)
	}
	return interactions.Interaction{
		Type:      interactions.TypeModal,
		CustomID:  modalID,
		ChannelID: "panel-channel",
		Actor:     interactions.Actor{GuildID: testGuildID, UserID: "user-1"},
		ModalFields: []customid.ModalField{
			{CustomID: "ticketcolor", Value: color},
			{CustomID: "tickettitle", Value: title},
			{CustomID: "ticketcontent", Value: content},
		},
	}
}

func legacyTicketModalInteraction(color string, title string, content string) interactions.Interaction {
	return interactions.Interaction{
		Type:      interactions.TypeModal,
		CustomID:  "nal",
		ChannelID: "panel-channel",
		Actor:     interactions.Actor{GuildID: testGuildID, UserID: "user-1"},
		ModalFields: []customid.ModalField{
			{CustomID: "ticketcolor", Value: color},
			{CustomID: "tickettitle", Value: title},
			{CustomID: "ticketcontent", Value: content},
		},
	}
}

func ticketButtonInteraction(customID string) interactions.Interaction {
	return interactions.Interaction{
		Type:      interactions.TypeComponent,
		CustomID:  customID,
		ChannelID: "panel-channel",
		MessageID: "panel-message",
		Actor:     interactions.Actor{GuildID: testGuildID, UserID: "user-1"},
	}
}

func seededTicketRepo(t *testing.T) *fakemongo.TicketConfigRepository {
	t.Helper()
	repo := fakemongo.NewTicketConfigRepository()
	if err := repo.SaveTicketConfig(context.Background(), domain.TicketConfig{
		GuildID:        testGuildID,
		CategoryID:     testCategoryID,
		AdminRoleID:    testAdminRole,
		EveryoneRoleID: testGuildID,
	}); err != nil {
		t.Fatalf("seed config: %v", err)
	}
	return repo
}

func flattenInputs(modal responses.Modal) []responses.TextInput {
	var inputs []responses.TextInput
	for _, row := range modal.Rows {
		inputs = append(inputs, row.Inputs...)
	}
	return inputs
}
