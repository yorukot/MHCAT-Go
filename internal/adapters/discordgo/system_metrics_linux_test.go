//go:build linux

package discordgo

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestLinuxSystemMetricsSamplerReadsHostAndProcessMetrics(t *testing.T) {
	metrics, err := (linuxSystemMetricsSampler{CPUInterval: time.Millisecond}).Sample(context.Background())
	if err != nil {
		t.Fatalf("sample metrics: %v", err)
	}
	if metrics.CPUModel == "" || metrics.CPUUsagePercent < 0 || metrics.CPUUsagePercent > 100 {
		t.Fatalf("cpu metrics = %#v", metrics)
	}
	if metrics.HostMemoryTotalMB <= 0 || metrics.HostMemoryUsedMB < 0 || metrics.HostMemoryUsedMB > metrics.HostMemoryTotalMB {
		t.Fatalf("host memory metrics = %#v", metrics)
	}
	if metrics.ProcessHeapMB <= 0 || metrics.ProcessRSSMB <= 0 {
		t.Fatalf("process memory metrics = %#v", metrics)
	}
}

func TestLinuxSystemMetricsSamplerHonorsCancellationDuringCPUSample(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := (linuxSystemMetricsSampler{CPUInterval: time.Second}).Sample(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("sample error = %v, want context canceled", err)
	}
}

func TestRoundedMiBMatchesJavaScriptRound(t *testing.T) {
	for _, test := range []struct {
		bytes uint64
		want  int64
	}{
		{bytes: 0, want: 0},
		{bytes: bytesPerMiB/2 - 1, want: 0},
		{bytes: bytesPerMiB / 2, want: 1},
		{bytes: bytesPerMiB + bytesPerMiB/2, want: 2},
	} {
		if got := roundedMiB(test.bytes); got != test.want {
			t.Errorf("roundedMiB(%d) = %d, want %d", test.bytes, got, test.want)
		}
	}
}
