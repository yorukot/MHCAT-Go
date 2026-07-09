package utility

import (
	"fmt"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type PingService struct {
	Clock ports.Clock
}

func (s PingService) Response(createdAt time.Time) string {
	now := time.Now()
	if s.Clock != nil {
		now = s.Clock.Now()
	}
	latency := time.Duration(0)
	if !createdAt.IsZero() && now.After(createdAt) {
		latency = now.Sub(createdAt)
	}
	ms := latency.Milliseconds()
	icon := "<:icons_goodping:1084881470075703367>"
	if ms > 180 {
		icon = "<:icons_badping:1084881519581069482>"
	} else if ms > 125 {
		icon = "<:icons_idelping:1084881570013388860>"
	}
	return fmt.Sprintf("%s **Pong!** `%d`ms", icon, ms)
}
