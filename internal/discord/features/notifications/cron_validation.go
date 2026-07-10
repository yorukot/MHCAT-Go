package notifications

import (
	"time"

	robfigcron "github.com/robfig/cron/v3"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/notifications"
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
	value = coreservice.NormalizeLegacyAutoNotificationCron(value)
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
