package utility

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/faketranslate"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestTranslateDefinitionMatchesLegacyShape(t *testing.T) {
	definition := commands.TranslateDefinition()
	if definition.Name != "翻譯" || definition.Description != "翻譯成各種語言" {
		t.Fatalf("definition = %#v", definition)
	}
	if len(definition.Options) != 2 || definition.Options[0].Name != "要的翻譯" || definition.Options[1].Name != "目標語言" {
		t.Fatalf("options = %#v", definition.Options)
	}
	if len(definition.Options[1].Choices) != 9 {
		t.Fatalf("choices = %#v", definition.Options[1].Choices)
	}
}

func TestTranslateHandlerRendersLegacyLoadingAndResultEmbeds(t *testing.T) {
	provider := &faketranslate.Translator{Result: ports.TranslationResult{Text: "hello"}}
	usage := &fakeusage.Tracker{}
	module := NewModuleWithTranslator(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), nil, nil, provider, nil, usage)
	module.translateColor = func() int { return 0x123456 }
	interaction := fakediscord.SlashInteractionWithOptions("翻譯", "", map[string]string{
		"要的翻譯": "你好",
		"目標語言": "en",
	})
	interaction.Actor.UserTag = "Tester#0001"
	interaction.Actor.AvatarURL = "https://example.invalid/avatar.png"
	responder := fakediscord.NewResponder()
	if err := module.TranslateHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("translate handler: %v", err)
	}
	if len(responder.Defers) != 1 {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(responder.Edits) != 2 {
		t.Fatalf("expected loading and result edits, got %#v", responder.Edits)
	}
	if responder.Edits[0].Embeds[0].Title != "<a:load:986319593444352071> | 我正在玩命幫你翻譯!" {
		t.Fatalf("loading embed = %#v", responder.Edits[0].Embeds[0])
	}
	final := responder.Edits[1].Embeds[0]
	if final.Title != "<:translate:986870996147507231> 翻譯系統" || final.Color != 0x123456 || len(final.Fields) != 3 {
		t.Fatalf("final embed = %#v", final)
	}
	if !strings.Contains(final.Fields[0].Value, "你好") || !strings.Contains(final.Fields[2].Value, "hello") {
		t.Fatalf("fields = %#v", final.Fields)
	}
	if final.Footer == nil || final.Footer.Text != "Tester#0001的查詢" {
		t.Fatalf("footer = %#v", final.Footer)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != "翻譯" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestLegacyTranslateRandomColorUsesFullDiscordRange(t *testing.T) {
	for range 100 {
		color := legacyTranslateRandomColor()
		if color < 0 || color > 0xFFFFFF {
			t.Fatalf("color = %#x", color)
		}
	}
}

func TestTranslateHandlerReturnsSafeProviderError(t *testing.T) {
	provider := &faketranslate.Translator{Err: errors.New("provider internal token abc")}
	module := NewModuleWithTranslator(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), nil, nil, provider, nil, nil)
	responder := fakediscord.NewResponder()
	err := module.TranslateHandler()(context.Background(), fakediscord.SlashInteractionWithOptions("翻譯", "", map[string]string{
		"要的翻譯": "hello",
		"目標語言": "en",
	}), responder)
	if err != nil {
		t.Fatalf("translate handler: %v", err)
	}
	last := responder.Edits[len(responder.Edits)-1]
	if strings.Contains(last.Embeds[0].Title, "provider internal") || !strings.Contains(last.Embeds[0].Title, "翻譯失敗") {
		t.Fatalf("error embed leaked provider detail: %#v", last.Embeds[0])
	}
}

func TestTranslateHandlerCanRenderProviderTimeoutError(t *testing.T) {
	module := NewModuleWithTranslator(
		commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}),
		nil,
		nil,
		blockingTranslator{},
		nil,
		nil,
	)
	module.translateTimeout = time.Millisecond
	responder := fakediscord.NewResponder()
	err := module.TranslateHandler()(context.Background(), fakediscord.SlashInteractionWithOptions("翻譯", "", map[string]string{
		"要的翻譯": "hello",
		"目標語言": "ja",
	}), responder)
	if err != nil {
		t.Fatalf("translate handler: %v", err)
	}
	if len(responder.Edits) != 2 || !strings.Contains(responder.Edits[1].Embeds[0].Title, "翻譯失敗") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

type blockingTranslator struct{}

func (blockingTranslator) Translate(ctx context.Context, _ ports.TranslationRequest) (ports.TranslationResult, error) {
	<-ctx.Done()
	return ports.TranslationResult{}, ctx.Err()
}
