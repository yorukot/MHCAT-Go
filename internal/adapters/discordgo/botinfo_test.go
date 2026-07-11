package discordgo

import (
	"context"
	"testing"
	"time"

	dgo "github.com/bwmarrin/discordgo"
)

func TestBotInfoProviderDegradesWithoutSession(t *testing.T) {
	provider := BotInfoProvider{Name: "MHCAT", ShardCount: 1, Metrics: fixedSystemMetricsSampler()}
	info, err := provider.BotInfo(context.Background())
	if err != nil {
		t.Fatalf("bot info: %v", err)
	}
	if info.GatewayConnected {
		t.Fatal("gateway should be disconnected without session")
	}
	if info.Name != "MHCAT" || info.ShardCount != 1 {
		t.Fatalf("info = %#v", info)
	}
}

func TestBotInfoProviderReadsCachedCounts(t *testing.T) {
	state := dgo.NewState()
	state.Guilds = []*dgo.Guild{
		{ID: "guild-1", MemberCount: 10},
		{ID: "guild-2", MemberCount: 20},
	}
	session := &Session{session: &dgo.Session{State: state}}
	provider := BotInfoProvider{Session: session, StartedAt: time.Now().Add(-time.Minute), ShardID: 1, ShardCount: 2, Metrics: fixedSystemMetricsSampler()}
	info, err := provider.BotInfo(context.Background())
	if err != nil {
		t.Fatalf("bot info: %v", err)
	}
	if info.GuildCount != 2 || info.UserCount != 30 {
		t.Fatalf("counts = guilds %d users %d", info.GuildCount, info.UserCount)
	}
	if info.ShardID != 1 || info.ShardCount != 2 {
		t.Fatalf("shard = %d/%d", info.ShardID, info.ShardCount)
	}
	if info.Uptime <= 0 {
		t.Fatalf("expected positive uptime")
	}
	if info.CPUModel != "test-cpu" || info.CPUUsagePercent != 12.5 || info.MemoryUsedMB != 768 || info.MemoryTotalMB != 1024 || info.ProcessHeapMB != 64 || info.ProcessRSSMB != 96 {
		t.Fatalf("metrics = %#v", info)
	}
}

type staticSystemMetricsSampler struct {
	metrics SystemMetrics
}

func (s staticSystemMetricsSampler) Sample(context.Context) (SystemMetrics, error) {
	return s.metrics, nil
}

func fixedSystemMetricsSampler() SystemMetricsSampler {
	return staticSystemMetricsSampler{metrics: SystemMetrics{
		CPUModel:          "test-cpu",
		CPUUsagePercent:   12.5,
		HostMemoryUsedMB:  768,
		HostMemoryTotalMB: 1024,
		ProcessHeapMB:     64,
		ProcessRSSMB:      96,
	}}
}
