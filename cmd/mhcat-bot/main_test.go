package main

import (
	"os"
	"strings"
	"testing"
)

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
