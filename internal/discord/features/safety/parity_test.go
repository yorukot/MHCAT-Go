package safety

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestAntiScamDefinitionsMatchLegacyVisibleContract(t *testing.T) {
	toggle := AntiScamDefinition()
	if toggle.Type != commands.CommandTypeChatInput || toggle.Name != "防詐騙網址" || toggle.Description != "設定是否開啟防詐騙網址功能(輸入這個指令就會更改)" || toggle.DefaultMemberPermissions != nil || len(toggle.Options) != 0 {
		t.Fatalf("toggle definition = %#v", toggle)
	}
	report := ScamReportDefinition()
	if report.Type != commands.CommandTypeChatInput || report.Name != "詐騙網址回報" || report.Description != "回報詐騙網站" || report.DefaultMemberPermissions != nil || len(report.Options) != 1 {
		t.Fatalf("report definition = %#v", report)
	}
	option := report.Options[0]
	if option.Type != commands.OptionTypeString || option.Name != "網址" || option.Description != "回報網址" || !option.Required {
		t.Fatalf("report option = %#v", option)
	}
}

func TestAntiScamMessagesMatchLegacyVisibleContract(t *testing.T) {
	for _, test := range []struct {
		open bool
		want string
	}{
		{open: true, want: "您的防詐騙啟用狀態已改為:\ntrue"},
		{open: false, want: "您的防詐騙啟用狀態已改為:\nfalse"},
	} {
		message := antiScamSuccessMessage(test.open)
		want := responses.Message{
			Embeds: []responses.Embed{{
				Title:       "<:fraudalert:1000408260777611355> 自動偵測詐騙連結",
				Description: test.want,
				Color:       0x57F287,
			}},
			AllowedMentions: &responses.AllowedMentions{},
		}
		if !reflect.DeepEqual(message, want) {
			t.Fatalf("toggle message = %#v, want %#v", message, want)
		}
	}

	report := scamReportSuccessMessage("ftp://bad.example/path")
	if len(report.Embeds) != 1 || report.Embeds[0].Title != "<:fraudalert:1000408260777611355> 自動偵測詐騙連結" || report.Embeds[0].Description != "成功回報ftp://bad.example/path" || report.Embeds[0].Color != 0x57F287 {
		t.Fatalf("report message = %#v", report)
	}
	errorMessage := antiScamErrorMessage("你輸入的不是一個網址!")
	if len(errorMessage.Embeds) != 1 || errorMessage.Embeds[0].Title != "<a:Discord_AnimatedNo:1015989839809757295> | 你輸入的不是一個網址!" || errorMessage.Embeds[0].Description != "" || errorMessage.Embeds[0].Color != 0xED4245 {
		t.Fatalf("error message = %#v", errorMessage)
	}
	if antiScamDeleteWarning != "<:trashbin:995991389043163257> | 此消息包含詐騙或釣魚連結，以自動刪除!\n" {
		t.Fatalf("delete warning = %q", antiScamDeleteWarning)
	}
}

func TestScamReportOptionPreservesLegacyRawValue(t *testing.T) {
	interaction := interactions.Interaction{Options: map[string]string{"網址": " ftp://bad.example "}}
	if got := firstOption(interaction, "網址"); got != " ftp://bad.example " {
		t.Fatalf("option = %q", got)
	}
}

func TestAntiScamMessageRuntimePreservesLegacyBotScanning(t *testing.T) {
	configs := fakemongo.NewAntiScamConfigRepository()
	configs.Configs["guild-1"] = domain.AntiScamConfig{GuildID: "guild-1", Open: true}
	catalog := fakemongo.NewScamURLCatalogRepository()
	catalog.Known = []string{"https://bad.example"}
	sideEffects := fakediscord.NewSideEffects()
	module := NewMessageDeleteModule(configs, catalog, sideEffects)
	event := antiScamMessageEvent("https://bad.example")
	event.IsBot = true

	if err := module.MessageCreateHandler()(context.Background(), event); err != nil {
		t.Fatalf("message create: %v", err)
	}
	if len(sideEffects.DeletedMessage) != 1 || len(sideEffects.Sent) != 1 || sideEffects.Sent[0].Message.Content != antiScamDeleteWarning {
		t.Fatalf("side effects deleted=%#v sent=%#v", sideEffects.DeletedMessage, sideEffects.Sent)
	}
}

func TestAntiScamURLValidatorMatchesPinnedLegacyCases(t *testing.T) {
	for _, test := range []struct {
		raw  string
		want bool
	}{
		{raw: "https://example.com", want: true},
		{raw: "ftp://example.com/path", want: true},
		{raw: "custom_1://example.com", want: true},
		{raw: "//example.com", want: true},
		{raw: "//localhost", want: true},
		{raw: "http://localhost:3000/path", want: true},
		{raw: "https://例子.公司", want: true},
		{raw: "//a.😀", want: true},
		{raw: "https://example.com\u0085", want: true},
		{raw: ""},
		{raw: "example.com"},
		{raw: "mailto:test@example.com"},
		{raw: " https://example.com "},
		{raw: "https://a.x"},
		{raw: "//a.é"},
		{raw: "https://example.com\u00a0"},
		{raw: "https://example.com\ufeff"},
	} {
		if got := domain.LooksLikeURL(test.raw); got != test.want {
			t.Fatalf("LooksLikeURL(%q) = %v, want %v", test.raw, got, test.want)
		}
	}
}

func TestAntiScamMessageSideEffectOrdering(t *testing.T) {
	deleteErr := errors.New("delete failed")
	sendErr := errors.New("send failed")
	for _, test := range []struct {
		name      string
		deleteErr error
		sendErr   error
		wantErr   error
		wantOps   []string
	}{
		{name: "success", wantOps: []string{"delete", "send"}},
		{name: "delete failure stops warning", deleteErr: deleteErr, wantErr: deleteErr, wantOps: []string{"delete"}},
		{name: "warning failure keeps deletion", sendErr: sendErr, wantErr: sendErr, wantOps: []string{"delete", "send"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			configs := fakemongo.NewAntiScamConfigRepository()
			configs.Configs["guild-1"] = domain.AntiScamConfig{GuildID: "guild-1", Open: true}
			catalog := fakemongo.NewScamURLCatalogRepository()
			catalog.Known = []string{"https://bad.example"}
			messages := &orderedMessagePort{deleteErr: test.deleteErr, sendErr: test.sendErr}
			module := NewMessageDeleteModule(configs, catalog, messages)

			err := module.MessageCreateHandler()(context.Background(), antiScamMessageEvent("visit https://bad.example"))
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("error = %v, want %v", err, test.wantErr)
			}
			if !reflect.DeepEqual(messages.ops, test.wantOps) {
				t.Fatalf("operations = %#v, want %#v", messages.ops, test.wantOps)
			}
			if len(messages.sent) > 0 {
				message := messages.sent[0]
				if message.Content != antiScamDeleteWarning || !reflect.DeepEqual(message.AllowedMentions, ports.AllowedMentions{}) {
					t.Fatalf("warning = %#v", message)
				}
			}
		})
	}
}

type orderedMessagePort struct {
	ops       []string
	sent      []ports.OutboundMessage
	deleteErr error
	sendErr   error
}

func (p *orderedMessagePort) SendMessage(_ context.Context, _ string, message ports.OutboundMessage) (ports.MessageRef, error) {
	p.ops = append(p.ops, "send")
	p.sent = append(p.sent, message)
	return ports.MessageRef{}, p.sendErr
}

func (p *orderedMessagePort) EditMessage(context.Context, ports.MessageRef, ports.OutboundMessage) error {
	return errors.New("unexpected edit")
}

func (p *orderedMessagePort) DeleteMessage(_ context.Context, _ ports.MessageRef) error {
	p.ops = append(p.ops, "delete")
	return p.deleteErr
}
