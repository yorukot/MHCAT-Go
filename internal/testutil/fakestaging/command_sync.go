package fakestaging

func CommandSyncEnv() map[string]string {
	return map[string]string{
		"MHCAT_DISCORD_TOKEN":                  "test-token",
		"MHCAT_DISCORD_APPLICATION_ID":         "app-1",
		"MHCAT_COMMAND_SYNC_SCOPE":             "guild",
		"MHCAT_COMMAND_SYNC_GUILD_ID":          "guild-1",
		"MHCAT_STAGING_MODE":                   "true",
		"MHCAT_STAGING_GUILD_ID":               "guild-1",
		"MHCAT_STAGING_ALLOWED_APPLICATION_ID": "app-1",
	}
}
