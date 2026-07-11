package discordgo

import (
	"context"
	"time"

	dgo "github.com/bwmarrin/discordgo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type BotInfoProvider struct {
	Session    *Session
	Name       string
	StartedAt  time.Time
	ShardID    int
	ShardCount int
	Metrics    SystemMetricsSampler
}

func (p BotInfoProvider) BotInfo(ctx context.Context) (ports.BotInfo, error) {
	if err := ctx.Err(); err != nil {
		return ports.BotInfo{}, err
	}
	metricsSampler := p.Metrics
	if metricsSampler == nil {
		metricsSampler = defaultSystemMetricsSampler()
	}
	metrics, err := metricsSampler.Sample(ctx)
	if err != nil {
		return ports.BotInfo{}, err
	}
	info := ports.BotInfo{
		Name:            p.Name,
		ShardID:         p.ShardID,
		ShardCount:      p.ShardCount,
		CPUModel:        metrics.CPUModel,
		CPUUsagePercent: metrics.CPUUsagePercent,
		MemoryUsedMB:    metrics.HostMemoryUsedMB,
		MemoryTotalMB:   metrics.HostMemoryTotalMB,
		ProcessHeapMB:   metrics.ProcessHeapMB,
		ProcessRSSMB:    metrics.ProcessRSSMB,
	}
	if info.Name == "" {
		info.Name = "MHCAT"
	}
	if info.ShardCount == 0 {
		info.ShardCount = 1
	}
	if !p.StartedAt.IsZero() {
		info.Uptime = time.Since(p.StartedAt)
	}
	if p.Session == nil {
		return info, nil
	}
	p.Session.mu.Lock()
	session := p.Session.session
	opened := p.Session.opened
	p.Session.mu.Unlock()
	info.GatewayConnected = opened
	if session == nil {
		return info, nil
	}
	if opened {
		info.Latency = session.HeartbeatLatency()
	}
	info.GuildCount, info.UserCount = stateCounts(session.State)
	return info, nil
}

func stateCounts(state *dgo.State) (int, int) {
	if state == nil {
		return 0, 0
	}
	state.RLock()
	defer state.RUnlock()
	users := 0
	for _, guild := range state.Guilds {
		users += guild.MemberCount
	}
	return len(state.Guilds), users
}
