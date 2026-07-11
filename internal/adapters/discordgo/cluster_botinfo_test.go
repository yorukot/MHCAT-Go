package discordgo

import (
	"context"
	"encoding/json"
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
