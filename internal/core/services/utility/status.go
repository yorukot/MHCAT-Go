package utility

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type StatusService struct {
	Provider ports.BotInfoProvider
}

func (s StatusService) Info(ctx context.Context) (ports.BotInfo, bool) {
	if s.Provider == nil {
		return ports.BotInfo{Name: "MHCAT"}, true
	}
	info, err := s.Provider.BotInfo(ctx)
	if err != nil {
		return ports.BotInfo{Name: "MHCAT"}, true
	}
	if info.Name == "" {
		info.Name = "MHCAT"
	}
	if info.ShardCount == 0 {
		info.ShardCount = 1
	}
	return info, false
}

func (s StatusService) ShardInfo(ctx context.Context) (ports.BotInfo, bool) {
	if provider, ok := s.Provider.(ports.ShardInfoProvider); ok {
		info, err := provider.ShardInfo(ctx)
		if err != nil {
			return ports.BotInfo{Name: "MHCAT"}, true
		}
		if info.Name == "" {
			info.Name = "MHCAT"
		}
		if info.ShardCount == 0 {
			info.ShardCount = 1
		}
		return info, false
	}
	return s.Info(ctx)
}

func (s StatusService) BotStatus(ctx context.Context) (string, error) {
	info, degraded := s.Info(ctx)
	if degraded && s.Provider == nil {
		return "MHCAT status\nStatus: degraded\nReason: bot info provider is not configured", nil
	}
	if degraded {
		return "MHCAT status\nStatus: degraded\nReason: bot information is temporarily unavailable", nil
	}
	connection := "disconnected"
	if info.GatewayConnected {
		connection = "connected"
	}
	lines := []string{
		fmt.Sprintf("%s status", info.Name),
		fmt.Sprintf("Gateway: %s", connection),
		fmt.Sprintf("Guilds: %d", info.GuildCount),
		fmt.Sprintf("Users: %d", info.UserCount),
		fmt.Sprintf("Shard: %d/%d", info.ShardID, info.ShardCount),
		fmt.Sprintf("Latency: %s", formatDuration(info.Latency)),
		fmt.Sprintf("Uptime: %s", formatDuration(info.Uptime)),
	}
	return strings.Join(lines, "\n"), nil
}

func formatDuration(value time.Duration) string {
	if value < 0 {
		value = 0
	}
	if value == 0 {
		return "0s"
	}
	return value.Truncate(time.Millisecond).String()
}
