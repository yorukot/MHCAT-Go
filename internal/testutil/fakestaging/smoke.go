package fakestaging

func GatewaySmokeEnv() map[string]string {
	env := CommandSyncEnv()
	env["MHCAT_MONGODB_URI"] = "mongodb://localhost:27017/mhcat-staging"
	env["MHCAT_MONGODB_DATABASE"] = "mhcat_staging"
	env["MHCAT_DISCORD_ENABLE_GATEWAY"] = "true"
	env["MHCAT_DISCORD_GATEWAY_SMOKE_TEST"] = "true"
	env["MHCAT_STAGING_ALLOW_GATEWAY_SMOKE"] = "true"
	return env
}
