package external

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

const defaultTranslateBaseURL = "https://translate.googleapis.com/translate_a/single"

type GoogleTranslateClient struct {
	Client  *http.Client
	BaseURL string
}

func NewGoogleTranslateClient() GoogleTranslateClient {
	return GoogleTranslateClient{
		Client:  &http.Client{Timeout: 10 * time.Second},
		BaseURL: defaultTranslateBaseURL,
	}
}

func (c GoogleTranslateClient) Translate(ctx context.Context, request ports.TranslationRequest) (ports.TranslationResult, error) {
	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}
	baseURL := strings.TrimSpace(c.BaseURL)
	if baseURL == "" {
		baseURL = defaultTranslateBaseURL
	}
	endpoint, err := url.Parse(baseURL)
	if err != nil {
		return ports.TranslationResult{}, fmt.Errorf("parse translate url: %w", err)
	}
	query := endpoint.Query()
	query.Set("client", "gtx")
	query.Set("sl", "auto")
	query.Set("tl", request.TargetLanguage)
	query.Set("dt", "t")
	query.Set("q", request.Text)
	endpoint.RawQuery = query.Encode()

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return ports.TranslationResult{}, fmt.Errorf("create translate request: %w", err)
	}
	response, err := client.Do(httpRequest)
	if err != nil {
		return ports.TranslationResult{}, fmt.Errorf("send translate request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return ports.TranslationResult{}, fmt.Errorf("translate response status %d", response.StatusCode)
	}
	var payload any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return ports.TranslationResult{}, fmt.Errorf("decode translate response: %w", err)
	}
	text := extractGoogleTranslateText(payload)
	if strings.TrimSpace(text) == "" {
		return ports.TranslationResult{}, errors.New("empty translate response")
	}
	return ports.TranslationResult{Text: text}, nil
}

func extractGoogleTranslateText(payload any) string {
	root, ok := payload.([]any)
	if !ok || len(root) == 0 {
		return ""
	}
	chunks, ok := root[0].([]any)
	if !ok {
		return ""
	}
	var builder strings.Builder
	for _, rawChunk := range chunks {
		chunk, ok := rawChunk.([]any)
		if !ok || len(chunk) == 0 {
			continue
		}
		text, ok := chunk[0].(string)
		if ok {
			builder.WriteString(text)
		}
	}
	return builder.String()
}
