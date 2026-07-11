package main

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/config"
)

func TestNewDiscordClientRejectsMissingCredentials(t *testing.T) {
	if _, err := newDiscordClient(config.CommandSyncConfig{}); err == nil {
		t.Fatal("Discord command client accepted missing credentials")
	}
}
