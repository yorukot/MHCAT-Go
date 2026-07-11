package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestProductionPreflightRejectsPositionalArguments(t *testing.T) {
	var stdout, stderr strings.Builder
	code := run(context.Background(), []string{"unexpected"}, nil, &stdout, &stderr, nil)
	if code == 0 || !strings.Contains(stderr.String(), "unexpected positional arguments") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestRecommendedShardCountUsesAuthenticatedGatewayEndpoint(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.String() != "https://discord.com/api/v10/gateway/bot" {
			t.Fatalf("url = %q", request.URL.String())
		}
		if request.Header.Get("Authorization") != "Bot token" {
			t.Fatalf("authorization = %q", request.Header.Get("Authorization"))
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"shards":16}`))}, nil
	})}
	count, err := recommendedShardCount(context.Background(), client, "token")
	if err != nil {
		t.Fatalf("recommended shards: %v", err)
	}
	if count != 16 {
		t.Fatalf("count = %d", count)
	}
}

func TestCheckAssetsRequiresEveryNonemptyFile(t *testing.T) {
	root := t.TempDir()
	for _, path := range requiredAssets {
		fullPath := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte("asset"), 0o644); err != nil {
			t.Fatalf("write asset: %v", err)
		}
	}
	if missing := checkAssets(root); len(missing) != 0 {
		t.Fatalf("missing = %#v", missing)
	}
	if err := os.Truncate(filepath.Join(root, requiredAssets[0]), 0); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	missing := checkAssets(root)
	if len(missing) != 1 || missing[0] != requiredAssets[0] {
		t.Fatalf("missing = %#v", missing)
	}
}
