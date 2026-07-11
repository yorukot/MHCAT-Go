package utility_test

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/utility"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/faketranslate"
)

func TestTranslateServiceTranslatesSupportedLanguage(t *testing.T) {
	provider := &faketranslate.Translator{Result: ports.TranslationResult{Text: "hello"}}
	service := utility.TranslateService{Translator: provider}
	result, err := service.Translate(context.Background(), " 你好 ", "en")
	if err != nil {
		t.Fatalf("translate: %v", err)
	}
	if result.Text != "hello" || provider.Requests[0].Text != " 你好 " || provider.Requests[0].TargetLanguage != "en" {
		t.Fatalf("result=%#v requests=%#v", result, provider.Requests)
	}
}

func TestTranslateServiceRejectsInvalidInput(t *testing.T) {
	service := utility.TranslateService{Translator: &faketranslate.Translator{}}
	if _, err := service.Translate(context.Background(), "", "en"); !errors.Is(err, utility.ErrInvalidTranslateInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
	if _, err := service.Translate(context.Background(), "hello", "xx"); !errors.Is(err, utility.ErrUnsupportedTranslateLanguage) {
		t.Fatalf("expected unsupported language, got %v", err)
	}
}

func TestTranslateServiceMapsProviderError(t *testing.T) {
	service := utility.TranslateService{Translator: &faketranslate.Translator{Err: errors.New("network secret")}}
	if _, err := service.Translate(context.Background(), "hello", "ja"); !errors.Is(err, utility.ErrTranslateUnavailable) {
		t.Fatalf("expected provider unavailable, got %v", err)
	}
}
