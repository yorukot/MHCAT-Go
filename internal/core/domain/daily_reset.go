package domain

type DailyResetResult struct {
	ExcludedGuilds       int
	CoinsMatched         int64
	CoinsModified        int64
	WorkGuilds           int
	WorkEnergyIncrements int64
	WorkEnergyClamps     int64
}
