package notifications

import (
	"strings"
	"time"

	robfigcron "github.com/robfig/cron/v3"
)

type directCronValidation uint8

const (
	directCronInvalid directCronValidation = iota
	directCronValid
	directCronTooFrequent
)

var autoNotificationCronParser = robfigcron.NewParser(
	robfigcron.Minute | robfigcron.Hour | robfigcron.Dom | robfigcron.Month | robfigcron.Dow,
)

func validateDirectCron(value string, now time.Time) directCronValidation {
	value = normalizeSundaySeven(value)
	schedule, err := autoNotificationCronParser.Parse(value)
	if err != nil {
		return directCronInvalid
	}
	if now.IsZero() {
		now = time.Now()
	}
	first := schedule.Next(now)
	second := schedule.Next(first)
	if first.IsZero() || second.IsZero() {
		return directCronInvalid
	}
	if second.Sub(first) < 15*time.Minute {
		return directCronTooFrequent
	}
	return directCronValid
}

func normalizeSundaySeven(value string) string {
	fields := strings.Fields(value)
	if len(fields) != 5 {
		return value
	}
	if fields[4] == "1-7" {
		fields[4] = "*"
		return strings.Join(fields, " ")
	}
	parts := strings.Split(fields[4], ",")
	for index, part := range parts {
		if part == "7" {
			parts[index] = "0"
		}
	}
	fields[4] = strings.Join(parts, ",")
	return strings.Join(fields, " ")
}
