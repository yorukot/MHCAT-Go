package ports

import "context"

type TranslationRequest struct {
	Text           string
	TargetLanguage string
}

type TranslationResult struct {
	Text string
}

type Translator interface {
	Translate(ctx context.Context, request TranslationRequest) (TranslationResult, error)
}
