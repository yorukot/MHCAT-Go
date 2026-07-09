package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStagingDisabledByDefault(t *testing.T) {
	cfg, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":          "token",
		"MHCAT_DISCORD_APPLICATION_ID": "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":  "guild",
	}))
	if err != nil {
		t.Fatalf("load command sync: %v", err)
	}
	if cfg.Staging.Mode {
		t.Fatal("staging mode must default false")
	}
	if !cfg.Staging.RequireGuildScope {
		t.Fatal("staging should require guild scope by default")
	}
	if cfg.Staging.AllowCommandApply || cfg.Staging.AllowGatewaySmoke {
		t.Fatalf("staging unsafe flags enabled by default: %#v", cfg.Staging)
	}
	if cfg.Staging.SmokeTimeout != 60*time.Second {
		t.Fatalf("unexpected smoke timeout: %v", cfg.Staging.SmokeTimeout)
	}
	if strings.Join(cfg.Staging.ExpectedCommands, ",") != "help,ping,info" {
		t.Fatalf("unexpected expected commands: %#v", cfg.Staging.ExpectedCommands)
	}
}

func TestStagingCommandApplyRequiresGuildID(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":               "token",
		"MHCAT_DISCORD_APPLICATION_ID":      "app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":       "guild",
		"MHCAT_STAGING_MODE":                "true",
		"MHCAT_STAGING_ALLOW_COMMAND_APPLY": "true",
		"MHCAT_COMMAND_SYNC_DRY_RUN":        "false",
	}))
	if !errors.Is(err, ErrInvalidStagingConfig) && !strings.Contains(errString(err), "MHCAT_STAGING_GUILD_ID") {
		t.Fatalf("expected staging guild id error, got %v", err)
	}
}

func TestStagingApplyRejectsGlobalScope(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(stagingApplyEnv(map[string]string{
		"MHCAT_COMMAND_SYNC_SCOPE":    "global",
		"MHCAT_COMMAND_SYNC_GUILD_ID": "",
	})))
	if !strings.Contains(errString(err), "global scope") && !strings.Contains(errString(err), "guild scope") {
		t.Fatalf("expected global scope rejection, got %v", err)
	}
}

func TestStagingApplyRejectsDeletion(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(stagingApplyEnv(map[string]string{
		"MHCAT_COMMAND_SYNC_ALLOW_DELETE": "true",
	})))
	if !strings.Contains(errString(err), "deletion") {
		t.Fatalf("expected delete rejection, got %v", err)
	}
}

func TestStagingApplyRejectsBulkOverwrite(t *testing.T) {
	_, err := LoadCommandSyncWithLookup(mapLookup(stagingApplyEnv(map[string]string{
		"MHCAT_COMMAND_SYNC_ALLOW_BULK_OVERWRITE": "true",
	})))
	if !strings.Contains(errString(err), "bulk overwrite") {
		t.Fatalf("expected bulk overwrite rejection, got %v", err)
	}
}

func TestStagingSmokeRequiresAllowFlag(t *testing.T) {
	_, err := LoadWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":              "token",
		"MHCAT_MONGODB_URI":                "mongodb://localhost:27017/mhcat",
		"MHCAT_MONGODB_DATABASE":           "mhcat",
		"MHCAT_DISCORD_ENABLE_GATEWAY":     "true",
		"MHCAT_DISCORD_GATEWAY_SMOKE_TEST": "true",
		"MHCAT_STAGING_MODE":               "true",
	}))
	if !strings.Contains(errString(err), "MHCAT_STAGING_ALLOW_GATEWAY_SMOKE") {
		t.Fatalf("expected gateway smoke allow error, got %v", err)
	}
}

func TestStagingConfigErrorsDoNotLeakToken(t *testing.T) {
	rawToken := strings.Repeat("a", 24) + "." + strings.Repeat("b", 6) + "." + strings.Repeat("c", 32)
	_, err := LoadCommandSyncWithLookup(mapLookup(map[string]string{
		"MHCAT_DISCORD_TOKEN":                  rawToken,
		"MHCAT_DISCORD_APPLICATION_ID":         "wrong-app",
		"MHCAT_COMMAND_SYNC_GUILD_ID":          "guild",
		"MHCAT_STAGING_MODE":                   "true",
		"MHCAT_STAGING_GUILD_ID":               "guild",
		"MHCAT_STAGING_ALLOWED_APPLICATION_ID": "expected-app",
	}))
	if err == nil {
		t.Fatal("expected staging application id error")
	}
	if strings.Contains(err.Error(), rawToken) {
		t.Fatalf("raw token leaked in error: %v", err)
	}
}

func TestStagingScriptsContainNoRealSecrets(t *testing.T) {
	for _, name := range []string{"command-sync-dry-run.sh", "command-sync-apply-guild.sh", "gateway-smoke.sh"} {
		path := filepath.Join("..", "..", "scripts", "staging", name)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		content := string(data)
		for _, forbidden := range []string{"discord.com/api/webhooks/", "Bot ", "mongodb+srv://"} {
			if strings.Contains(content, forbidden) {
				t.Fatalf("script %s contains forbidden secret-like literal %q", name, forbidden)
			}
		}
	}
}

func stagingApplyEnv(overrides map[string]string) map[string]string {
	env := map[string]string{
		"MHCAT_DISCORD_TOKEN":               "token",
		"MHCAT_DISCORD_APPLICATION_ID":      "app",
		"MHCAT_COMMAND_SYNC_SCOPE":          "guild",
		"MHCAT_COMMAND_SYNC_GUILD_ID":       "guild",
		"MHCAT_COMMAND_SYNC_DRY_RUN":        "false",
		"MHCAT_STAGING_MODE":                "true",
		"MHCAT_STAGING_GUILD_ID":            "guild",
		"MHCAT_STAGING_ALLOW_COMMAND_APPLY": "true",
	}
	for key, value := range overrides {
		env[key] = value
	}
	return env
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
