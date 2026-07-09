package utility_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/utility"
)

func TestStatusProviderReturnsExpectedData(t *testing.T) {
	service := utility.StatusService{Provider: fakeBotInfoProvider{info: ports.BotInfo{
		Name:             "MHCAT",
		ShardID:          0,
		ShardCount:       2,
		GuildCount:       12,
		UserCount:        345,
		Latency:          42 * time.Millisecond,
		Uptime:           3*time.Hour + 2*time.Minute,
		GatewayConnected: true,
	}}}
	got, err := service.BotStatus(context.Background())
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	for _, want := range []string{"MHCAT status", "Gateway: connected", "Guilds: 12", "Users: 345", "Shard: 0/2", "Latency: 42ms"} {
		if !strings.Contains(got, want) {
			t.Fatalf("status missing %q:\n%s", want, got)
		}
	}
}

func TestStatusDegradedProviderError(t *testing.T) {
	service := utility.StatusService{Provider: fakeBotInfoProvider{err: errors.New("internal failure")}}
	got, err := service.BotStatus(context.Background())
	if err != nil {
		t.Fatalf("status should return safe degraded response, got error %v", err)
	}
	if !strings.Contains(got, "Status: degraded") || strings.Contains(got, "internal failure") {
		t.Fatalf("unsafe degraded response:\n%s", got)
	}
}

func TestStatusNoProviderSafe(t *testing.T) {
	service := utility.StatusService{}
	got, err := service.BotStatus(context.Background())
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if !strings.Contains(got, "provider is not configured") {
		t.Fatalf("unexpected response:\n%s", got)
	}
}

type fakeBotInfoProvider struct {
	info ports.BotInfo
	err  error
}

func (f fakeBotInfoProvider) BotInfo(context.Context) (ports.BotInfo, error) {
	return f.info, f.err
}
