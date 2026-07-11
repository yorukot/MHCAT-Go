package domain

import "testing"

func TestWorkUserStateUsesLegacyEndTimeScalar(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		wantState string
		wantTime  string
	}{
		{name: "decimal", text: "100.5", wantState: "在礦坑打工", wantTime: "<t:100.5:R>"},
		{name: "infinity", text: "Infinity", wantState: "在礦坑打工", wantTime: "<t:Infinity:R>"},
		{name: "null", text: "null", wantState: WorkIdleState, wantTime: "`沒有打工再進行`"},
		{name: "malformed", text: "undefined", wantState: WorkIdleState, wantTime: "`沒有打工再進行`"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := WorkUserState{State: "礦坑", EndTimeUnix: 100, EndTimeText: test.text}
			if got := state.EffectiveState(100); got != test.wantState {
				t.Fatalf("effective state = %q", got)
			}
			if got := state.RemainingTimeText(100); got != test.wantTime {
				t.Fatalf("remaining time = %q", got)
			}
		})
	}
}

func TestWorkUserStatePreservesFutureEmptyState(t *testing.T) {
	state := WorkUserState{EndTimeText: "101"}
	if got := state.EffectiveState(100); got != "在打工" {
		t.Fatalf("effective state = %q", got)
	}
}
