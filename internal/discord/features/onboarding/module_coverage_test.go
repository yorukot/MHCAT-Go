package onboarding

import "testing"

func TestOnboardingModuleCommandsFollowEnabledFeatures(t *testing.T) {
	module := Module{joinRoleConfigured: true, leaveConfigured: true, verificationEnabled: true, flowEnabled: true, accountAgeEnabled: true}
	if len(module.Commands()) == 0 {
		t.Fatal("enabled onboarding module commands must not be empty")
	}
}
