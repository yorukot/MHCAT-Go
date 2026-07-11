package external

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

func TestNewGoogleTranslateClientUsesBoundedDefaults(t *testing.T) {
	client := NewGoogleTranslateClient()
	if client.Client == nil || client.Client.Timeout != 10*time.Second || client.BaseURL != defaultTranslateBaseURL {
		t.Fatalf("translate client defaults = %#v", client)
	}
}

func TestGoogleTranslateClientParsesResponse(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if r.Form.Get("client") != "gtx" || r.Form.Get("sl") != "auto" || r.Form.Get("tl") != "en" || r.Form.Get("dt") != "t" || r.Form.Get("q") != "你好" {
			t.Fatalf("form = %#v", r.Form)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`[[["hello ","你好",null,null,1],["world","世界",null,null,1]],null,"zh-TW"]`)),
			Header:     make(http.Header),
		}, nil
	})}
	client := GoogleTranslateClient{Client: httpClient, BaseURL: "https://translate.example.invalid/translate"}
	result, err := client.Translate(context.Background(), ports.TranslationRequest{Text: "你好", TargetLanguage: "en"})
	if err != nil {
		t.Fatalf("translate: %v", err)
	}
	if result.Text != "hello world" {
		t.Fatalf("text = %q", result.Text)
	}
}

func TestGoogleTranslateClientSendsMaximumDiscordInputInRequestBody(t *testing.T) {
	text := strings.Repeat("界", 6000)
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.RawQuery != "" {
			t.Fatalf("raw query contains translation input: %q", r.URL.RawQuery)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if r.Form.Get("q") != text {
			t.Fatalf("translation body length = %d", len([]rune(r.Form.Get("q"))))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`[[["world","界",null,null,1]],null,"zh-TW"]`)),
			Header:     make(http.Header),
		}, nil
	})}
	client := GoogleTranslateClient{Client: httpClient, BaseURL: "https://translate.example.invalid/translate"}
	result, err := client.Translate(context.Background(), ports.TranslationRequest{Text: text, TargetLanguage: "en"})
	if err != nil {
		t.Fatalf("translate: %v", err)
	}
	if result.Text != "world" {
		t.Fatalf("text = %q", result.Text)
	}
}

func TestGoogleTranslateClientRejectsBadStatus(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadGateway,
			Body:       io.NopCloser(strings.NewReader("bad")),
			Header:     make(http.Header),
		}, nil
	})}
	client := GoogleTranslateClient{Client: httpClient, BaseURL: "https://translate.example.invalid/translate"}
	if _, err := client.Translate(context.Background(), ports.TranslationRequest{Text: "hello", TargetLanguage: "ja"}); err == nil {
		t.Fatal("expected error")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}
