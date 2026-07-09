package app

import (
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/config"
)

func TestGatewaySmokeTimeoutUsesStagingTimeoutWhenAllowed(t *testing.T) {
	cfg := validTestConfig()
	cfg.DiscordGatewaySmokeTimeout = 10 * time.Second
	cfg.Staging = config.StagingConfig{
		Mode:              true,
		AllowGatewaySmoke: true,
		SmokeTimeout:      55 * time.Second,
	}
	if got := GatewaySmokeTimeout(cfg); got != 55*time.Second {
		t.Fatalf("timeout = %v", got)
	}
}

func TestGatewaySmokeTimeoutFallsBackToGatewayTimeout(t *testing.T) {
	cfg := validTestConfig()
	cfg.DiscordGatewaySmokeTimeout = 10 * time.Second
	cfg.Staging = config.StagingConfig{SmokeTimeout: 55 * time.Second}
	if got := GatewaySmokeTimeout(cfg); got != 10*time.Second {
		t.Fatalf("timeout = %v", got)
	}
}
