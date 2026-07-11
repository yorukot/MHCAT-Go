package discordgo

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	dgo "github.com/bwmarrin/discordgo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

func TestClusterBotInfoProviderAggregatesFreshShardFiles(t *testing.T) {
	directory := t.TempDir()
	state := dgo.NewState()
	state.Guilds = []*dgo.Guild{{ID: "guild-0", MemberCount: 10}}
	localSession := &Session{session: &dgo.Session{State: state}, opened: true}
	provider := &ClusterBotInfoProvider{
		Local: BotInfoProvider{
			Session: localSession, Name: "MHCAT", StartedAt: time.Now().Add(-time.Minute),
			ShardID: 0, ShardCount: 2, Metrics: fixedSystemMetricsSampler(),
		},
		Directory: directory,
		MaxAge:    time.Minute,
	}
	writeShardStatusForTest(t, directory, shardStatusFile{
		UpdatedAt: time.Now(),
		Info: ports.BotInfo{
			Name: "MHCAT", ShardID: 1, ShardCount: 2, GuildCount: 2, UserCount: 20,
			ProcessHeapMB: 30, ProcessRSSMB: 40, Uptime: 2 * time.Minute, GatewayConnected: true,
		},
	})

	info, err := provider.BotInfo(context.Background())
	if err != nil {
		t.Fatalf("bot info: %v", err)
	}
	if info.GuildCount != 3 || info.UserCount != 30 || info.ProcessHeapMB != 94 || info.ProcessRSSMB != 136 {
		t.Fatalf("aggregate = %#v", info)
	}
	if info.ShardCount != 2 || info.Uptime < 2*time.Minute || !info.GatewayConnected {
		t.Fatalf("aggregate runtime = %#v", info)
	}
	shard, err := provider.ShardInfo(context.Background())
	if err != nil || shard.ShardID != 0 || shard.ShardCount != 2 || shard.GuildCount != 1 {
		t.Fatalf("local shard info = %#v err=%v", shard, err)
	}
	infos, err := provider.ShardInfos(context.Background())
	if err != nil {
		t.Fatalf("shard infos: %v", err)
	}
	if len(infos) != 2 || infos[0].ShardID != 0 || infos[1].ShardID != 1 {
		t.Fatalf("shards = %#v", infos)
	}
}

func TestClusterBotInfoProviderIgnoresStaleAndWrongClusterFiles(t *testing.T) {
	directory := t.TempDir()
	provider := &ClusterBotInfoProvider{
		Local:     BotInfoProvider{ShardID: 0, ShardCount: 2, Metrics: fixedSystemMetricsSampler()},
		Directory: directory,
		MaxAge:    time.Second,
	}
	writeShardStatusForTest(t, directory, shardStatusFile{UpdatedAt: time.Now().Add(-time.Hour), Info: ports.BotInfo{ShardID: 1, ShardCount: 2}})
	data, _ := json.Marshal(shardStatusFile{UpdatedAt: time.Now(), Info: ports.BotInfo{ShardID: 2, ShardCount: 3}})
	if err := os.WriteFile(filepath.Join(directory, "wrong.json"), data, 0o600); err != nil {
		t.Fatalf("write wrong status: %v", err)
	}
	infos, err := provider.ShardInfos(context.Background())
	if err != nil {
		t.Fatalf("shard infos: %v", err)
	}
	if len(infos) != 1 || infos[0].ShardID != 0 {
		t.Fatalf("shards = %#v", infos)
	}
}

func TestClusterBotInfoProviderUsesLowFrequencyDefaults(t *testing.T) {
	provider := &ClusterBotInfoProvider{}
	if provider.interval() != 30*time.Second {
		t.Fatalf("interval = %v", provider.interval())
	}
	if provider.maxAge() < 3*provider.interval() {
		t.Fatalf("max age %v is too close to interval %v", provider.maxAge(), provider.interval())
	}
}

func TestClusterBotInfoProviderStartStopLifecycle(t *testing.T) {
	directory := t.TempDir()
	samples := make(chan struct{}, 4)
	provider := &ClusterBotInfoProvider{
		Local: BotInfoProvider{
			ShardID: 0, ShardCount: 2,
			Metrics: &countingProcessMetricsSampler{samples: samples},
		},
		Directory: directory,
		Interval:  5 * time.Millisecond,
		MaxAge:    time.Minute,
	}
	if !provider.Start(context.Background()) || provider.Start(context.Background()) {
		t.Fatal("unexpected start result")
	}
	for range 2 {
		select {
		case <-samples:
		case <-time.After(time.Second):
			t.Fatal("status publisher did not run")
		}
	}
	if err := provider.Stop(context.Background()); err != nil {
		t.Fatalf("stop: %v", err)
	}
	if err := provider.Stop(context.Background()); err != nil {
		t.Fatalf("second stop: %v", err)
	}
	if _, err := os.Stat(provider.statusPath(0)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("status file remains: %v", err)
	}
}

type countingProcessMetricsSampler struct {
	samples chan struct{}
}

func (s *countingProcessMetricsSampler) Sample(context.Context) (SystemMetrics, error) {
	return SystemMetrics{}, nil
}

func (s *countingProcessMetricsSampler) SampleProcess(context.Context) (int64, int64, error) {
	s.samples <- struct{}{}
	return 1, 2, nil
}

func writeShardStatusForTest(t *testing.T, directory string, status shardStatusFile) {
	t.Helper()
	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("marshal status: %v", err)
	}
	path := filepath.Join(directory, "shard-"+strconv.Itoa(status.Info.ShardID)+".json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write status: %v", err)
	}
}
