package utility

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
	"unicode/utf16"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/faketranslate"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestTranslateDefinitionMatchesLegacyShape(t *testing.T) {
	want := commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        "翻譯",
		Description: "翻譯成各種語言",
		DocsURL:     "https://docsmhcat.yorukot.me/docs/translate",
		Ownership:   commands.ManagedOwnership("translate", commands.ScopeGuild),
		Options: []commands.Option{
			{Type: commands.OptionTypeString, Name: "要的翻譯", Description: "你要翻譯的句子或是單詞!", Required: true},
			{
				Type:        commands.OptionTypeString,
				Name:        "目標語言",
				Description: "你要翻譯成的語言!",
				Required:    true,
				Choices: []commands.Choice{
					{Name: "🇹🇼中文(traditional Chinese)", Value: "zh-TW"},
					{Name: "🇺🇸英文(English)", Value: "en"},
					{Name: "🇯🇵日文(Japanese)", Value: "ja"},
					{Name: "🇰🇷韓語(Korean)", Value: "ko"},
					{Name: "🇩🇪德語(German)", Value: "de"},
					{Name: "🇫🇷法語(French)", Value: "fr"},
					{Name: "🇷🇺俄語(Russian)", Value: "ru"},
					{Name: "🇪🇸西班牙語(Spanish)", Value: "es"},
					{Name: "🇨🇳簡體中文(Simplified Chinese)", Value: "zh-CN"},
				},
			},
		},
	}
	if got := commands.TranslateDefinition(); !reflect.DeepEqual(got, want) {
		t.Fatalf("definition = %#v, want %#v", got, want)
	}
}

func TestTranslateHandlerRendersLegacyLoadingAndResultEmbeds(t *testing.T) {
	provider := &faketranslate.Translator{Result: ports.TranslationResult{Text: "hello"}}
	usage := &fakeusage.Tracker{}
	module := NewModuleWithTranslator(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), nil, nil, provider, nil, usage)
	module.translateColor = func() int { return 0x123456 }
	interaction := fakediscord.SlashInteractionWithOptions("翻譯", "", map[string]string{
		"要的翻譯": " 你好 ",
		"目標語言": "en",
	})
	interaction.Actor.UserTag = "Tester#0001"
	interaction.Actor.AvatarURL = "https://example.invalid/avatar.png"
	responder := fakediscord.NewResponder()
	if err := module.TranslateHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("translate handler: %v", err)
	}
	if !reflect.DeepEqual(responder.Defers, []responses.DeferOptions{{}}) {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(responder.Follow) != 1 || len(responder.FollowEdits) != 1 || len(responder.Edits) != 0 {
		t.Fatalf("follow=%#v follow edits=%#v original edits=%#v", responder.Follow, responder.FollowEdits, responder.Edits)
	}
	wantLoading := responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:load:986319593444352071> | 我正在玩命幫你翻譯!",
			Color: 0x57F287,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
	if !reflect.DeepEqual(responder.Follow[0], wantLoading) {
		t.Fatalf("loading = %#v, want %#v", responder.Follow[0], wantLoading)
	}
	if responder.FollowEdits[0].MessageID != responder.FollowIDs[0] {
		t.Fatalf("follow-up edit = %#v ids=%#v", responder.FollowEdits, responder.FollowIDs)
	}
	wantFinal := responses.Message{
		Embeds: []responses.Embed{{
			Title: "<:translate:986870996147507231> 翻譯系統",
			Color: 0x123456,
			Fields: []responses.EmbedField{
				{Name: "**<:edittext:986873966884962304> 原文**:", Value: "` 你好 `", Inline: false},
				{Name: "**<:answer:986873630178832414> 目標語言:**", Value: "`en`", Inline: false},
				{Name: "**<:translate1:986873633483939901> 譯文:**", Value: "`hello`", Inline: false},
			},
			Footer: &responses.EmbedFooter{Text: "Tester#0001的查詢", IconURL: "https://example.invalid/avatar.png"},
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
	if got := responder.FollowEdits[0].Message; !reflect.DeepEqual(got, wantFinal) {
		t.Fatalf("final = %#v, want %#v", got, wantFinal)
	}
	if len(provider.Requests) != 1 || provider.Requests[0].Text != " 你好 " {
		t.Fatalf("provider requests = %#v", provider.Requests)
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

func TestTranslateCodeFieldUsesDiscordUTF16Limit(t *testing.T) {
	value := strings.Repeat("😀", 600)
	got := codeField(value)
	if units := len(utf16.Encode([]rune(got))); units > 1024 {
		t.Fatalf("UTF-16 units = %d", units)
	}
	if !strings.HasSuffix(got, "...`") || strings.ContainsRune(got, '\uFFFD') {
		t.Fatalf("field = %q", got)
	}
	if fitting := codeField(strings.Repeat("a", 1022)); fitting != "`"+strings.Repeat("a", 1022)+"`" {
		t.Fatalf("fitting field was truncated: %q", fitting)
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
	want := responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | 很抱歉，翻譯失敗，請稍後再試!",
			Color: 0xEA0000,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
	last := responder.FollowEdits[len(responder.FollowEdits)-1].Message
	if !reflect.DeepEqual(last, want) || strings.Contains(last.Embeds[0].Title, "provider internal") {
		t.Fatalf("error = %#v, want %#v", last, want)
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
	if len(responder.Follow) != 1 || len(responder.FollowEdits) != 1 || !strings.Contains(responder.FollowEdits[0].Message.Embeds[0].Title, "翻譯失敗") {
		t.Fatalf("follow=%#v follow edits=%#v", responder.Follow, responder.FollowEdits)
	}
}

type blockingTranslator struct{}

func (blockingTranslator) Translate(ctx context.Context, _ ports.TranslationRequest) (ports.TranslationResult, error) {
	<-ctx.Done()
	return ports.TranslationResult{}, ctx.Err()
}
