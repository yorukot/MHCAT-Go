package app

import (
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/config"
)

func GatewaySmokeTimeout(cfg config.Config) time.Duration {
	if cfg.Staging.Mode && cfg.Staging.AllowGatewaySmoke && cfg.Staging.SmokeTimeout > 0 {
		return cfg.Staging.SmokeTimeout
	}
	return cfg.DiscordGatewaySmokeTimeout
}
