package external

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

func TestGoogleTranslateClientParsesResponse(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Query().Get("tl") != "en" || r.URL.Query().Get("q") != "你好" {
			t.Fatalf("query = %s", r.URL.RawQuery)
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
