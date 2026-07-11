package discordgo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

const (
	DefaultShardStatusDirectory = "/tmp/mhcat-shards"
	DefaultShardStatusInterval  = 5 * time.Second
	DefaultShardStatusMaxAge    = 20 * time.Second
)

type shardStatusFile struct {
	UpdatedAt time.Time     `json:"updated_at"`
	Info      ports.BotInfo `json:"info"`
}

type ClusterBotInfoProvider struct {
	Local     BotInfoProvider
	Directory string
	Interval  time.Duration
	MaxAge    time.Duration

	mu     sync.Mutex
	cancel context.CancelFunc
	done   chan struct{}
}

func (p *ClusterBotInfoProvider) Start(ctx context.Context) bool {
	if p == nil {
		return false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.cancel != nil {
		return false
	}
	workerCtx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	p.done = make(chan struct{})
	go p.run(workerCtx, p.done)
	return true
}

func (p *ClusterBotInfoProvider) Stop(ctx context.Context) error {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	cancel := p.cancel
	done := p.done
	p.cancel = nil
	p.done = nil
	p.mu.Unlock()
	if cancel == nil {
		return nil
	}
	cancel()
	select {
	case <-done:
	case <-ctx.Done():
		return ctx.Err()
	}
	err := os.Remove(p.statusPath(p.Local.ShardID))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (p *ClusterBotInfoProvider) BotInfo(ctx context.Context) (ports.BotInfo, error) {
	local, err := p.Local.BotInfo(ctx)
	if err != nil {
		return ports.BotInfo{}, err
	}
	infos, err := p.ShardInfos(ctx)
	if err != nil {
		return ports.BotInfo{}, err
	}
	local.GuildCount = 0
	local.UserCount = 0
	local.ProcessHeapMB = 0
	local.ProcessRSSMB = 0
	local.GatewayConnected = true
	for _, info := range infos {
		local.GuildCount += info.GuildCount
		local.UserCount += info.UserCount
		local.ProcessHeapMB += info.ProcessHeapMB
		local.ProcessRSSMB += info.ProcessRSSMB
		if info.Uptime > local.Uptime {
			local.Uptime = info.Uptime
		}
		local.GatewayConnected = local.GatewayConnected && info.GatewayConnected
	}
	return local, nil
}

func (p *ClusterBotInfoProvider) ShardInfo(ctx context.Context) (ports.BotInfo, error) {
	return p.Local.ShardInfo(ctx)
}

func (p *ClusterBotInfoProvider) ShardInfos(ctx context.Context) ([]ports.BotInfo, error) {
	if err := p.publish(ctx); err != nil {
		return nil, err
	}
	directory := p.directory()
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("read shard status directory: %w", err)
	}
	cutoff := time.Now().Add(-p.maxAge())
	infos := make([]ports.BotInfo, 0, p.Local.ShardCount)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(directory, entry.Name()))
		if err != nil {
			continue
		}
		var status shardStatusFile
		if json.Unmarshal(data, &status) != nil || status.UpdatedAt.Before(cutoff) {
			continue
		}
		if status.Info.ShardID < 0 || status.Info.ShardID >= p.Local.ShardCount || status.Info.ShardCount != p.Local.ShardCount {
			continue
		}
		infos = append(infos, status.Info)
	}
	sort.Slice(infos, func(i, j int) bool { return infos[i].ShardID < infos[j].ShardID })
	if len(infos) == 0 {
		return nil, errors.New("no fresh shard status available")
	}
	return infos, nil
}

func (p *ClusterBotInfoProvider) run(ctx context.Context, done chan struct{}) {
	defer close(done)
	_ = p.publish(ctx)
	ticker := time.NewTicker(p.interval())
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = p.publish(ctx)
		}
	}
}

func (p *ClusterBotInfoProvider) publish(ctx context.Context) error {
	info, err := p.Local.ShardInfo(ctx)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(p.directory(), 0o700); err != nil {
		return fmt.Errorf("create shard status directory: %w", err)
	}
	data, err := json.Marshal(shardStatusFile{UpdatedAt: time.Now().UTC(), Info: info})
	if err != nil {
		return err
	}
	target := p.statusPath(info.ShardID)
	temporary := target + fmt.Sprintf(".%d.tmp", os.Getpid())
	if err := os.WriteFile(temporary, data, 0o600); err != nil {
		return fmt.Errorf("write shard status: %w", err)
	}
	if err := os.Rename(temporary, target); err != nil {
		_ = os.Remove(temporary)
		return fmt.Errorf("publish shard status: %w", err)
	}
	return nil
}

func (p *ClusterBotInfoProvider) directory() string {
	if p.Directory != "" {
		return p.Directory
	}
	return DefaultShardStatusDirectory
}

func (p *ClusterBotInfoProvider) statusPath(shardID int) string {
	return filepath.Join(p.directory(), fmt.Sprintf("shard-%d.json", shardID))
}

func (p *ClusterBotInfoProvider) interval() time.Duration {
	if p.Interval > 0 {
		return p.Interval
	}
	return DefaultShardStatusInterval
}

func (p *ClusterBotInfoProvider) maxAge() time.Duration {
	if p.MaxAge > 0 {
		return p.MaxAge
	}
	return DefaultShardStatusMaxAge
}

var _ ports.BotInfoProvider = (*ClusterBotInfoProvider)(nil)
var _ ports.ShardInfoProvider = (*ClusterBotInfoProvider)(nil)
var _ ports.ShardInfosProvider = (*ClusterBotInfoProvider)(nil)
