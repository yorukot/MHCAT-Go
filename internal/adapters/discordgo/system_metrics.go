package discordgo

import "context"

type SystemMetrics struct {
	CPUModel          string
	CPUUsagePercent   float64
	HostMemoryUsedMB  int64
	HostMemoryTotalMB int64
	ProcessHeapMB     int64
	ProcessRSSMB      int64
}

type SystemMetricsSampler interface {
	Sample(ctx context.Context) (SystemMetrics, error)
}
