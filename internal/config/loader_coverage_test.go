package config

import "testing"

func TestTopLevelLoadersRejectMissingRequiredConfiguration(t *testing.T) {
	t.Setenv("MHCAT_DISCORD_TOKEN", "")
	t.Setenv("DISCORD_TOKEN", "")
	t.Setenv("MHCAT_MONGODB_URI", "")
	t.Setenv("MONGOOSE_CONNECTION_STRING", "")
	t.Setenv("MHCAT_MONGODB_DATABASE", "")
	tests := []struct {
		name string
		run  func() error
	}{
		{name: "application", run: func() error { _, err := Load(); return err }},
		{name: "command sync", run: func() error { _, err := LoadCommandSync(); return err }},
		{name: "daily reset", run: func() error { _, err := LoadDailyReset(); return err }},
		{name: "mongo admin", run: func() error { _, err := LoadMongoAdmin(); return err }},
		{name: "scheduler lease", run: func() error { _, err := LoadSchedulerLease(); return err }},
		{name: "work payout", run: func() error { _, err := LoadWorkPayout(); return err }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.run(); err == nil {
				t.Fatal("missing required configuration was accepted")
			}
		})
	}
}
