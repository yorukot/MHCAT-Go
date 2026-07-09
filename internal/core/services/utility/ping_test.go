package utility_test

import (
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

type fakeClock struct {
	now time.Time
}

func (f fakeClock) Now() time.Time {
	return f.now
}
