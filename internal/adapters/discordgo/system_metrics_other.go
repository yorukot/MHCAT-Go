//go:build !linux

package discordgo

import (
	"context"
	"runtime"
)

type portableSystemMetricsSampler struct{}

func defaultSystemMetricsSampler() SystemMetricsSampler {
	return portableSystemMetricsSampler{}
}

func (portableSystemMetricsSampler) Sample(ctx context.Context) (SystemMetrics, error) {
	if err := ctx.Err(); err != nil {
		return SystemMetrics{}, err
	}
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return SystemMetrics{
		CPUModel:      runtime.GOOS + "/" + runtime.GOARCH,
		ProcessHeapMB: int64(stats.HeapSys / 1024 / 1024),
	}, nil
}
