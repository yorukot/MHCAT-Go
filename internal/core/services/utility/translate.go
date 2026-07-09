package utility

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

var (
	ErrTranslateUnavailable         = errors.New("translate provider unavailable")
	ErrInvalidTranslateInput        = errors.New("invalid translate input")
	ErrUnsupportedTranslateLanguage = errors.New("unsupported translate language")
)

type TranslateService struct {
	Translator ports.Translator
}

var SupportedTranslateLanguages = map[string]string{
	"zh-TW": "🇹🇼中文(traditional Chinese)",
	"en":    "🇺🇸英文(English)",
	"ja":    "🇯🇵日文(Japanese)",
	"ko":    "🇰🇷韓語(Korean)",
	"de":    "🇩🇪德語(German)",
	"fr":    "🇫🇷法語(French)",
	"ru":    "🇷🇺俄語(Russian)",
	"es":    "🇪🇸西班牙語(Spanish)",
	"zh-CN": "🇨🇳簡體中文(Simplified Chinese)",
}

func (s TranslateService) Translate(ctx context.Context, text string, targetLanguage string) (ports.TranslationResult, error) {
	text = strings.TrimSpace(text)
	targetLanguage = strings.TrimSpace(targetLanguage)
	if text == "" || targetLanguage == "" {
		return ports.TranslationResult{}, ErrInvalidTranslateInput
	}
	if _, ok := SupportedTranslateLanguages[targetLanguage]; !ok {
		return ports.TranslationResult{}, fmt.Errorf("%w: %s", ErrUnsupportedTranslateLanguage, targetLanguage)
	}
	if s.Translator == nil {
		return ports.TranslationResult{}, ErrTranslateUnavailable
	}
	result, err := s.Translator.Translate(ctx, ports.TranslationRequest{
		Text:           text,
		TargetLanguage: targetLanguage,
	})
	if err != nil {
		return ports.TranslationResult{}, fmt.Errorf("%w: %v", ErrTranslateUnavailable, err)
	}
	if strings.TrimSpace(result.Text) == "" {
		return ports.TranslationResult{}, ErrTranslateUnavailable
	}
	return result, nil
}
