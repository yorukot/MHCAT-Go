package discordgo

import (
	"context"
	"runtime"
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
}

func (p BotInfoProvider) BotInfo(ctx context.Context) (ports.BotInfo, error) {
	if err := ctx.Err(); err != nil {
		return ports.BotInfo{}, err
	}
	info := ports.BotInfo{
		Name:            p.Name,
		ShardID:         p.ShardID,
		ShardCount:      p.ShardCount,
		CPUModel:        runtime.GOOS + "/" + runtime.GOARCH,
		MemoryUsedMB:    processMemoryUsedMB(),
		MemoryTotalMB:   processMemoryTotalMB(),
		CPUUsagePercent: 0,
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

func processMemoryUsedMB() int64 {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return int64(stats.Alloc / 1024 / 1024)
}

func processMemoryTotalMB() int64 {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return int64(stats.Sys / 1024 / 1024)
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
