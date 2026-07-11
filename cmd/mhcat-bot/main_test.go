package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

func TestBotRunDelegatesOutputAndReturnsSuccess(t *testing.T) {
	var stdout, stderr bytes.Buffer
	called := false
	code := run(context.Background(), &stdout, &stderr, func(ctx context.Context, output io.Writer) error {
		called = true
		if ctx == nil {
			t.Fatal("runner received nil context")
		}
		_, err := io.WriteString(output, "started")
		return err
	})
	if code != 0 || !called || stdout.String() != "started" || stderr.Len() != 0 {
		t.Fatalf("code=%d called=%v stdout=%q stderr=%q", code, called, stdout.String(), stderr.String())
	}
}

func TestBotRunReportsRunnerAndConfigurationErrors(t *testing.T) {
	for _, test := range []struct {
		name   string
		runner appRunner
		want   string
	}{
		{name: "missing runner", want: "app runner is required"},
		{name: "runner error", runner: func(context.Context, io.Writer) error { return errors.New("startup failed") }, want: "startup failed"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := run(context.Background(), &stdout, &stderr, test.runner); code == 0 || stdout.Len() != 0 || !strings.Contains(stderr.String(), test.want) {
				t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
			}
		})
	}
}

func TestBotEntrypointDoesNotUseCommandSyncOrApplicationCommandMutation(t *testing.T) {
	data, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	content := string(data)
	forbidden := []string{
		"cmd/mhcat-command-sync",
		"internal/discord/commands/sync",
		"ApplicationCommand" + "Create",
		"ApplicationCommand" + "Edit",
		"ApplicationCommand" + "Delete",
		"ApplicationCommand" + "BulkOverwrite",
	}
	for _, value := range forbidden {
		if strings.Contains(content, value) {
			t.Fatalf("bot entrypoint contains forbidden command mutation path %q", value)
		}
	}
}
