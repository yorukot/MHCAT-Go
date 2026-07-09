package safety

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestToggleHandlerRequiresManageMessages(t *testing.T) {
	repo := fakemongo.NewAntiScamConfigRepository()
	usage := &fakeusage.Tracker{}
	module := NewModule(repo, usage)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction(AntiScamCommandName)

	if err := module.ToggleHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "你需要有`訊息管理`才能使用此指令") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	if responder.Edits[0].AllowedMentions == nil {
		t.Fatal("expected allowed mentions to be disabled explicitly")
	}
	if len(repo.Saved) != 0 {
		t.Fatalf("saved configs = %#v", repo.Saved)
	}
	if len(usage.Events) != 0 {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestToggleHandlerCreatesOpenConfigAndRendersLegacySuccess(t *testing.T) {
	repo := fakemongo.NewAntiScamConfigRepository()
	usage := &fakeusage.Tracker{}
	module := NewModule(repo, usage)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction(AntiScamCommandName)
	interaction.Actor.PermissionBits = permissionManageMessages

	if err := module.ToggleHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	saved, ok := repo.Last()
	if !ok || saved.GuildID != "guild-1" || !saved.Open {
		t.Fatalf("saved = %#v ok=%v", saved, ok)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != "<:fraudalert:1000408260777611355> 自動偵測詐騙連結" {
		t.Fatalf("title = %q", embed.Title)
	}
	if embed.Description != "您的防詐騙啟用狀態已改為:\ntrue" {
		t.Fatalf("description = %q", embed.Description)
	}
	if embed.Color != antiScamSuccessColor {
		t.Fatalf("color = %#x", embed.Color)
	}
	if responder.Edits[0].AllowedMentions == nil {
		t.Fatal("expected allowed mentions to be disabled explicitly")
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != AntiScamCommandName || usage.Events[0].Feature != "anti-scam-config" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestToggleHandlerFlipsExistingConfigFalse(t *testing.T) {
	repo := fakemongo.NewAntiScamConfigRepository()
	repo.Configs["guild-1"] = domain.AntiScamConfig{GuildID: "guild-1", Open: true}
	module := NewModule(repo, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction(AntiScamCommandName)
	interaction.Actor.PermissionBits = permissionManageMessages

	if err := module.ToggleHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	saved, ok := repo.Last()
	if !ok || saved.Open {
		t.Fatalf("saved = %#v ok=%v", saved, ok)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Description != "您的防詐騙啟用狀態已改為:\nfalse" {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestToggleHandlerUsesSafeUnknownError(t *testing.T) {
	repo := fakemongo.NewAntiScamConfigRepository()
	repo.Err = errors.New("mongodb://secret")
	module := NewModule(repo, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction(AntiScamCommandName)
	interaction.Actor.PermissionBits = permissionManageMessages

	if err := module.ToggleHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	title := responder.Edits[0].Embeds[0].Title
	if !strings.Contains(title, "很抱歉，出現了未知的錯誤，請重試!") || strings.Contains(title, "mongodb://secret") {
		t.Fatalf("unsafe error title = %q", title)
	}
}

func TestReportHandlerSendsWebhookAndRendersLegacySuccess(t *testing.T) {
	catalog := fakemongo.NewScamURLCatalogRepository()
	sender := &fakeReportSender{}
	usage := &fakeusage.Tracker{}
	module := NewReportModule(catalog, sender, usage)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(ScamReportCommandName, "", map[string]string{"網址": "https://bad.example"})

	if err := module.ReportHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(sender.Sent) != 1 || sender.Sent[0].URL != "https://bad.example" || sender.Sent[0].ReporterUserID != "user-1" {
		t.Fatalf("sent = %#v", sender.Sent)
	}
	if len(catalog.Checked) != 1 || catalog.Checked[0] != "https://bad.example" {
		t.Fatalf("checked = %#v", catalog.Checked)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != "<:fraudalert:1000408260777611355> 自動偵測詐騙連結" {
		t.Fatalf("title = %q", embed.Title)
	}
	if embed.Description != "成功回報https://bad.example" {
		t.Fatalf("description = %q", embed.Description)
	}
	if embed.Color != antiScamSuccessColor {
		t.Fatalf("color = %#x", embed.Color)
	}
	if responder.Edits[0].AllowedMentions == nil {
		t.Fatal("expected allowed mentions to be disabled explicitly")
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != ScamReportCommandName || usage.Events[0].Feature != "anti-scam-report" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestReportHandlerRejectsInvalidURL(t *testing.T) {
	catalog := fakemongo.NewScamURLCatalogRepository()
	sender := &fakeReportSender{}
	module := NewReportModule(catalog, sender, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(ScamReportCommandName, "", map[string]string{"網址": "not-a-url"})

	if err := module.ReportHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(sender.Sent) != 0 {
		t.Fatalf("sent = %#v", sender.Sent)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "你輸入的不是一個網址!") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestReportHandlerRejectsKnownURL(t *testing.T) {
	catalog := fakemongo.NewScamURLCatalogRepository()
	catalog.Known = []string{"https://bad.example/path"}
	sender := &fakeReportSender{}
	module := NewReportModule(catalog, sender, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(ScamReportCommandName, "", map[string]string{"網址": "https://bad.example"})

	if err := module.ReportHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(sender.Sent) != 0 {
		t.Fatalf("sent = %#v", sender.Sent)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "該網站已被回報過") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestReportHandlerUsesSafeUnknownError(t *testing.T) {
	catalog := fakemongo.NewScamURLCatalogRepository()
	webhookURL := "https://discord.com/api/" + "webhooks/secret"
	sender := &fakeReportSender{Err: errors.New(webhookURL)}
	module := NewReportModule(catalog, sender, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(ScamReportCommandName, "", map[string]string{"網址": "https://bad.example"})

	if err := module.ReportHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	title := responder.Edits[0].Embeds[0].Title
	if !strings.Contains(title, "很抱歉，出現了未知的錯誤，請重試!") || strings.Contains(title, webhookURL) {
		t.Fatalf("unsafe error title = %q", title)
	}
}

type fakeReportSender struct {
	Sent []domain.ScamURLReport
	Err  error
}

func (s *fakeReportSender) SendScamURLReport(ctx context.Context, report domain.ScamURLReport) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.Err != nil {
		return s.Err
	}
	s.Sent = append(s.Sent, report)
	return nil
}

var _ ports.ScamReportSender = (*fakeReportSender)(nil)
