package notifications

import "testing"

func TestNormalizeLegacyAutoNotificationCronWeekdaySeven(t *testing.T) {
	for _, test := range []struct {
		input string
		want  string
	}{
		{input: "0 9 * * 7", want: "0 9 * * 0"},
		{input: "0 9 * * 1,7", want: "0 9 * * 0,1"},
		{input: "0 9 * * 5-7", want: "0 9 * * 0,5,6"},
		{input: "0 9 * * 5-7/2", want: "0 9 * * 0,5"},
		{input: "0 9 * * 6-7/2", want: "0 9 * * 6"},
		{input: "0 9 * * 1-7", want: "0 9 * * *"},
		{input: "0 9 * * 0-7/2", want: "0 9 * * 0,2,4,6"},
		{input: "  */30   * * * *  ", want: "*/30 * * * *"},
	} {
		t.Run(test.input, func(t *testing.T) {
			if got := NormalizeLegacyAutoNotificationCron(test.input); got != test.want {
				t.Fatalf("NormalizeLegacyAutoNotificationCron(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestValidLegacyAutoNotificationCronMatchesNumericValidator(t *testing.T) {
	for _, test := range []struct {
		value string
		want  bool
	}{
		{value: "*/30 * * * *", want: true},
		{value: "0,15,30,45 0-23/2 1,31 1-12 0-7", want: true},
		{value: "  */30   * * * *  ", want: true},
		{value: "0 9 * * MON"},
		{value: "0 9 * JAN 1"},
		{value: "@daily"},
		{value: "0 0 0 * * *"},
		{value: "*/ * * * *"},
		{value: "0 25 * * *"},
		{value: "0 9 * * 1-0"},
	} {
		if got := ValidLegacyAutoNotificationCron(test.value); got != test.want {
			t.Fatalf("ValidLegacyAutoNotificationCron(%q) = %v, want %v", test.value, got, test.want)
		}
	}
}
