package xp

import "testing"

func TestXPModuleMetadata(t *testing.T) {
	modules := []struct {
		name     string
		commands int
	}{
		{name: (Module{}).Name(), commands: len((Module{}).Commands())},
		{name: (VoiceModule{}).Name(), commands: len((VoiceModule{}).Commands())},
		{name: (DisabledProfileModule{}).Name(), commands: len((DisabledProfileModule{}).Commands())},
		{name: (RewardRoleModule{}).Name(), commands: len((RewardRoleModule{}).Commands())},
		{name: (AdminModule{}).Name(), commands: len((AdminModule{}).Commands())},
		{name: (ResetModule{}).Name(), commands: len((ResetModule{}).Commands())},
		{name: (RankModule{}).Name(), commands: len((RankModule{}).Commands())},
	}
	for _, module := range modules {
		if module.name == "" || module.commands == 0 {
			t.Fatalf("xp module metadata = %#v", module)
		}
	}
	if (TextEventModule{}).Name() != "text-xp-accrual" || (VoiceEventModule{}).Name() != "voice-xp-sessions" {
		t.Fatal("xp event module names are invalid")
	}
	(ResetModule{}).RegisterEventRoutes(nil)
	(TextEventModule{}).RegisterEventRoutes(nil)
	(VoiceEventModule{}).RegisterEventRoutes(nil)
}
