package ports

import (
	"context"
	"time"
)

type BotInfo struct {
	Name             string
	ShardID          int
	ShardCount       int
	GuildCount       int
	UserCount        int
	Latency          time.Duration
	Uptime           time.Duration
	CPUModel         string
	CPUUsagePercent  float64
	MemoryUsedMB     int64
	MemoryTotalMB    int64
	ProcessHeapMB    int64
	ProcessRSSMB     int64
	GatewayConnected bool
}

type BotInfoProvider interface {
	BotInfo(ctx context.Context) (BotInfo, error)
}

type ShardInfoProvider interface {
	ShardInfo(ctx context.Context) (BotInfo, error)
}

type ShardInfosProvider interface {
	ShardInfos(ctx context.Context) ([]BotInfo, error)
}
