package xp

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/xp"
)

const LegacyVoiceXPInterval = 30 * time.Second

type VoiceXPTickFunc func(ctx context.Context, guildID string, userID string, currentRoleIDs []string) (coreservice.VoiceAccrualResult, error)

type VoiceXPWorker struct {
	interval time.Duration
	tick     VoiceXPTickFunc
	logger   *slog.Logger

	mu      sync.Mutex
	active  map[string]voiceXPWorkerEntry
	stopped bool
}

type voiceXPWorkerEntry struct {
	cancel context.CancelFunc
	done   chan struct{}
}

func NewVoiceXPWorker(interval time.Duration, tick VoiceXPTickFunc, logger *slog.Logger) *VoiceXPWorker {
	if interval <= 0 {
		interval = LegacyVoiceXPInterval
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &VoiceXPWorker{
		interval: interval,
		tick:     tick,
		logger:   logger,
		active:   map[string]voiceXPWorkerEntry{},
	}
}

func (w *VoiceXPWorker) Start(guildID string, userID string, currentRoleIDs []string) bool {
	if w == nil || w.tick == nil {
		return false
	}
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return false
	}
	key := voiceXPWorkerKey(guildID, userID)
	roles := trimmedRoleIDs(currentRoleIDs)
	ctx, cancel := context.WithCancel(context.Background())
	entry := voiceXPWorkerEntry{cancel: cancel, done: make(chan struct{})}

	w.mu.Lock()
	if w.stopped {
		w.mu.Unlock()
		cancel()
		return false
	}
	if _, ok := w.active[key]; ok {
		w.mu.Unlock()
		cancel()
		return false
	}
	w.active[key] = entry
	w.mu.Unlock()

	go w.run(ctx, entry.done, key, guildID, userID, roles)
	return true
}

func (w *VoiceXPWorker) Stop(guildID string, userID string) bool {
	if w == nil {
		return false
	}
	key := voiceXPWorkerKey(guildID, userID)
	w.mu.Lock()
	entry, ok := w.active[key]
	if ok {
		delete(w.active, key)
	}
	w.mu.Unlock()
	if !ok {
		return false
	}
	entry.cancel()
	return true
}

func (w *VoiceXPWorker) StopAll(ctx context.Context) error {
	if w == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	w.mu.Lock()
	entries := make([]voiceXPWorkerEntry, 0, len(w.active))
	for key, entry := range w.active {
		entries = append(entries, entry)
		delete(w.active, key)
	}
	w.stopped = true
	w.mu.Unlock()

	for _, entry := range entries {
		entry.cancel()
	}
	var errs []error
	for _, entry := range entries {
		select {
		case <-entry.done:
		case <-ctx.Done():
			errs = append(errs, ctx.Err())
		}
	}
	return errors.Join(errs...)
}

func (w *VoiceXPWorker) ActiveCount() int {
	if w == nil {
		return 0
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.active)
}

func (w *VoiceXPWorker) run(ctx context.Context, done chan struct{}, key string, guildID string, userID string, currentRoleIDs []string) {
	defer close(done)
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			result, err := w.tick(ctx, guildID, userID, currentRoleIDs)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				if errors.Is(err, ports.ErrVoiceXPProfileMissing) {
					w.finish(key, done)
					return
				}
				w.logger.WarnContext(ctx, "voice xp tick failed", "guild_id", guildID, "user_id", userID, "error", err.Error())
				continue
			}
			if !result.Active {
				w.finish(key, done)
				return
			}
		}
	}
}

func (w *VoiceXPWorker) finish(key string, done chan struct{}) {
	w.mu.Lock()
	defer w.mu.Unlock()
	entry, ok := w.active[key]
	if ok && entry.done == done {
		delete(w.active, key)
	}
}

func (m VoiceEventModule) WithRuntimeWorker(interval time.Duration, logger *slog.Logger) VoiceEventModule {
	m.worker = NewVoiceXPWorker(interval, m.TickVoiceXP, logger)
	return m
}

func (m VoiceEventModule) StopRuntimeWorker(ctx context.Context) error {
	if m.worker == nil {
		return nil
	}
	return m.worker.StopAll(ctx)
}

func voiceXPWorkerKey(guildID string, userID string) string {
	return strings.TrimSpace(guildID) + "\x00" + strings.TrimSpace(userID)
}

func trimmedRoleIDs(roleIDs []string) []string {
	roles := make([]string, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		roleID = strings.TrimSpace(roleID)
		if roleID != "" {
			roles = append(roles, roleID)
		}
	}
	return roles
}
