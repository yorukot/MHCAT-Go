package utility_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/utility"
)

func TestPingResponseGoodLatency(t *testing.T) {
	clock := fakeClock{now: time.Unix(100, 150_000_000)}
	service := utility.PingService{Clock: clock}
	got := service.Response(time.Unix(100, 0))
	want := "<:icons_idelping:1084881570013388860> **Pong!** `150`ms"
	if got != want {
		t.Fatalf("response = %q, want %q", got, want)
	}
}

func TestPingResponseZeroWhenCreatedAtMissing(t *testing.T) {
	service := utility.PingService{Clock: fakeClock{now: time.Unix(100, 0)}}
	got := service.Response(time.Time{})
	if got != "<:icons_goodping:1084881470075703367> **Pong!** `0`ms" {
		t.Fatalf("response = %q", got)
	}
}

func TestPingResponsePreservesLegacyNegativeLatency(t *testing.T) {
	service := utility.PingService{Clock: fakeClock{now: time.Unix(100, 0)}}
	got := service.Response(time.Unix(100, 50_000_000))
	want := "<:icons_goodping:1084881470075703367> **Pong!** `-50`ms"
	if got != want {
		t.Fatalf("response = %q, want %q", got, want)
	}
}

func TestPingResponseLegacyThresholdBoundaries(t *testing.T) {
	now := time.Unix(100, 0)
	service := utility.PingService{Clock: fakeClock{now: now}}
	tests := []struct {
		milliseconds int64
		icon         string
	}{
		{milliseconds: 125, icon: "<:icons_goodping:1084881470075703367>"},
		{milliseconds: 126, icon: "<:icons_idelping:1084881570013388860>"},
		{milliseconds: 180, icon: "<:icons_idelping:1084881570013388860>"},
		{milliseconds: 181, icon: "<:icons_badping:1084881519581069482>"},
	}
	for _, test := range tests {
		createdAt := now.Add(-time.Duration(test.milliseconds) * time.Millisecond)
		want := fmt.Sprintf("%s **Pong!** `%d`ms", test.icon, test.milliseconds)
		if got := service.Response(createdAt); got != want {
			t.Errorf("latency %dms response = %q, want %q", test.milliseconds, got, want)
		}
	}
}

type fakeClock struct {
	now time.Time
}

func (f fakeClock) Now() time.Time {
	return f.now
}
