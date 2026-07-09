package faketranslate

import (
	"context"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type Translator struct {
	Result   ports.TranslationResult
	Err      error
	Requests []ports.TranslationRequest
}

func (t *Translator) Translate(ctx context.Context, request ports.TranslationRequest) (ports.TranslationResult, error) {
	if err := ctx.Err(); err != nil {
		return ports.TranslationResult{}, err
	}
	t.Requests = append(t.Requests, request)
	if t.Err != nil {
		return ports.TranslationResult{}, t.Err
	}
	return t.Result, nil
}
