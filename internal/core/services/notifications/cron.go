package notifications

import (
	"strconv"
	"strings"
)

// NormalizeLegacyAutoNotificationCron maps the validator's Sunday value 7
// into the 0-6 range accepted by both robfig/cron and the legacy scheduler.
func NormalizeLegacyAutoNotificationCron(value string) string {
	fields := strings.Fields(value)
	if len(fields) != 5 {
		return value
	}
	weekdays, ok := normalizeLegacyAutoNotificationWeekdays(fields[4])
	if !ok {
		return strings.Join(fields, " ")
	}
	fields[4] = weekdays
	return strings.Join(fields, " ")
}

func normalizeLegacyAutoNotificationWeekdays(value string) (string, bool) {
	selected := [7]bool{}
	for _, condition := range strings.Split(value, ",") {
		parts := strings.Split(condition, "/")
		if len(parts) > 2 || len(parts) == 0 {
			return "", false
		}
		step := 1
		if len(parts) == 2 {
			parsed, err := strconv.Atoi(parts[1])
			if err != nil || parsed <= 0 {
				return "", false
			}
			step = parsed
		}

		start, end, ok := legacyAutoNotificationWeekdayRange(parts[0], len(parts) == 2)
		if !ok {
			return "", false
		}
		for day := start; day <= end; day += step {
			selected[day%7] = true
		}
	}

	result := make([]string, 0, len(selected))
	for day, included := range selected {
		if included {
			result = append(result, strconv.Itoa(day))
		}
	}
	if len(result) == len(selected) {
		return "*", true
	}
	if len(result) == 0 {
		return "", false
	}
	return strings.Join(result, ","), true
}

func legacyAutoNotificationWeekdayRange(value string, stepped bool) (int, int, bool) {
	if value == "*" {
		return 0, 7, true
	}
	if startText, endText, ranged := strings.Cut(value, "-"); ranged {
		start, startErr := strconv.Atoi(startText)
		end, endErr := strconv.Atoi(endText)
		return start, end, startErr == nil && endErr == nil && start >= 0 && start <= end && end <= 7
	}
	start, err := strconv.Atoi(value)
	if err != nil || start < 0 || start > 7 {
		return 0, 0, false
	}
	if stepped {
		return start, 7, true
	}
	return start, start, true
}
