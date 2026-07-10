package notifications

import (
	"testing"
	"time"
)

func TestValidateDirectCron(t *testing.T) {
	now := time.Date(2026, time.July, 10, 0, 1, 0, 0, time.UTC)
	for _, test := range []struct {
		name  string
		value string
		want  directCronValidation
	}{
		{name: "thirty minute step", value: "*/30 * * * *", want: directCronValid},
		{name: "weekday seven", value: "0 9 * * 7", want: directCronValid},
		{name: "weekday one through seven", value: "0 9 * * 1-7", want: directCronValid},
		{name: "weekday range through seven", value: "0 9 * * 5-7", want: directCronValid},
		{name: "weekday stepped range through seven", value: "0 9 * * 5-7/2", want: directCronValid},
		{name: "every minute", value: "* * * * *", want: directCronTooFrequent},
		{name: "five minute step", value: "*/5 * * * *", want: directCronTooFrequent},
		{name: "cancel starts wizard", value: "cancel", want: directCronInvalid},
		{name: "localized cancel starts wizard", value: "取消", want: directCronInvalid},
		{name: "out of range", value: "0 25 * * *", want: directCronInvalid},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := validateDirectCron(test.value, now); got != test.want {
				t.Fatalf("validateDirectCron(%q) = %v, want %v", test.value, got, test.want)
			}
		})
	}
}

func TestAutoNotificationWeekExpressionPreservesLegacyAllDaysShortcut(t *testing.T) {
	got, ok := autoNotificationWeekExpression([]string{"0", "6", "5", "4", "3", "2", "1"})
	if !ok || got != "*" {
		t.Fatalf("week expression = %q, %v", got, ok)
	}
	got, ok = autoNotificationWeekExpression([]string{"3", "1"})
	if !ok || got != "1,3" {
		t.Fatalf("partial week expression = %q, %v", got, ok)
	}
}
