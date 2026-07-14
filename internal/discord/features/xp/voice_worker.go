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
	active  map[string]*voiceXPWorkerEntry
	stopped bool
	running bool
	rootCtx context.Context
	cancel  context.CancelFunc
	done    chan struct{}
	wake    chan struct{}
}

type voiceXPWorkerEntry struct {
	ctx            context.Context
	cancel         context.CancelFunc
	guildID        string
	userID         string
	currentRoleIDs []string
	nextTick       time.Time
	generation     uint64
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
		active:   map[string]*voiceXPWorkerEntry{},
		wake:     make(chan struct{}, 1),
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

	w.mu.Lock()
	if w.stopped {
		w.mu.Unlock()
		return false
	}
	if _, ok := w.active[key]; ok {
		w.mu.Unlock()
		return false
	}
	w.ensureRunningLocked()
	ctx, cancel := context.WithCancel(w.rootCtx)
	w.active[key] = &voiceXPWorkerEntry{
		ctx: ctx, cancel: cancel, guildID: guildID, userID: userID,
		currentRoleIDs: roles, nextTick: time.Now().Add(w.interval), generation: 1,
	}
	w.mu.Unlock()
	w.notify()
	return true
}

func (w *VoiceXPWorker) StartOrUpdate(guildID string, userID string, currentRoleIDs []string) bool {
	if w == nil || w.tick == nil {
		return false
	}
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return false
	}
	key := voiceXPWorkerKey(guildID, userID)
	w.mu.Lock()
	if entry, ok := w.active[key]; ok {
		entry.currentRoleIDs = trimmedRoleIDs(currentRoleIDs)
		entry.generation++
		w.mu.Unlock()
		return false
	}
	w.mu.Unlock()
	return w.Start(guildID, userID, currentRoleIDs)
}

func (w *VoiceXPWorker) ReconcileGuild(guildID string, states []ports.DiscordVoiceState) (int, int) {
	if w == nil || w.tick == nil {
		return 0, 0
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return 0, 0
	}
	desired := make(map[string][]string, len(states))
	for _, state := range states {
		userID := strings.TrimSpace(state.UserID)
		if userID == "" || state.IsBot || strings.TrimSpace(state.ChannelID) == "" {
			continue
		}
		desired[userID] = trimmedRoleIDs(state.RoleIDs)
	}

	var canceled []context.CancelFunc
	started := 0
	stopped := 0
	w.mu.Lock()
	if w.stopped {
		w.mu.Unlock()
		return 0, 0
	}
	for key, entry := range w.active {
		if entry.guildID != guildID {
			continue
		}
		if _, ok := desired[entry.userID]; ok {
			continue
		}
		delete(w.active, key)
		canceled = append(canceled, entry.cancel)
		stopped++
	}
	for userID, roles := range desired {
		key := voiceXPWorkerKey(guildID, userID)
		if entry, ok := w.active[key]; ok {
			entry.currentRoleIDs = roles
			entry.generation++
			continue
		}
		w.ensureRunningLocked()
		ctx, cancel := context.WithCancel(w.rootCtx)
		w.active[key] = &voiceXPWorkerEntry{
			ctx: ctx, cancel: cancel, guildID: guildID, userID: userID,
			currentRoleIDs: roles, nextTick: time.Now().Add(w.interval), generation: 1,
		}
		started++
	}
	w.mu.Unlock()
	for _, cancel := range canceled {
		cancel()
	}
	if started > 0 || stopped > 0 {
		w.notify()
	}
	return started, stopped
}

func (w *VoiceXPWorker) StopGuild(guildID string) int {
	if w == nil {
		return 0
	}
	guildID = strings.TrimSpace(guildID)
	var canceled []context.CancelFunc
	w.mu.Lock()
	for key, entry := range w.active {
		if entry.guildID != guildID {
			continue
		}
		delete(w.active, key)
		canceled = append(canceled, entry.cancel)
	}
	w.mu.Unlock()
	for _, cancel := range canceled {
		cancel()
	}
	if len(canceled) > 0 {
		w.notify()
	}
	return len(canceled)
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
	w.notify()
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
	for key, entry := range w.active {
		entry.cancel()
		delete(w.active, key)
	}
	w.stopped = true
	cancel := w.cancel
	done := w.done
	w.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if done == nil {
		return ctx.Err()
	}
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (w *VoiceXPWorker) ActiveCount() int {
	if w == nil {
		return 0
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.active)
}

func (w *VoiceXPWorker) run(ctx context.Context, done chan struct{}) {
	defer func() {
		w.mu.Lock()
		if w.done == done {
			w.running = false
			w.rootCtx = nil
			w.cancel = nil
		}
		w.mu.Unlock()
		close(done)
	}()
	timer := time.NewTimer(time.Hour)
	stopVoiceXPTimer(timer)
	defer timer.Stop()
	for {
		deadline, ok := w.nextDeadline()
		if !ok {
			select {
			case <-ctx.Done():
				return
			case <-w.wake:
				continue
			}
		}
		resetVoiceXPTimer(timer, time.Until(deadline))
		select {
		case <-ctx.Done():
			return
		case <-w.wake:
			stopVoiceXPTimer(timer)
		case now := <-timer.C:
			w.runDue(now)
		}
	}
}

func (w *VoiceXPWorker) ensureRunningLocked() {
	if w.running {
		return
	}
	w.rootCtx, w.cancel = context.WithCancel(context.Background())
	w.done = make(chan struct{})
	w.running = true
	go w.run(w.rootCtx, w.done)
}

func (w *VoiceXPWorker) runDue(cutoff time.Time) {
	for _, due := range w.collectDue(cutoff) {
		key, entry := due.key, due.entry
		result, err := w.tick(entry.ctx, entry.guildID, entry.userID, due.currentRoleIDs)
		if err != nil {
			if entry.ctx.Err() != nil {
				continue
			}
			if errors.Is(err, ports.ErrVoiceXPProfileMissing) {
				if !w.finish(key, entry, due.generation) {
					w.scheduleNext(key, entry)
				}
				continue
			}
			w.logger.WarnContext(entry.ctx, "voice xp tick failed", "guild_id", entry.guildID, "user_id", entry.userID, "error", err.Error())
			w.scheduleNext(key, entry)
			continue
		}
		if !result.Active {
			if !w.finish(key, entry, due.generation) {
				w.scheduleNext(key, entry)
			}
			continue
		}
		w.scheduleNext(key, entry)
	}
}

type dueVoiceXPWorkerEntry struct {
	key            string
	entry          *voiceXPWorkerEntry
	currentRoleIDs []string
	generation     uint64
}

func (w *VoiceXPWorker) collectDue(cutoff time.Time) []dueVoiceXPWorkerEntry {
	w.mu.Lock()
	defer w.mu.Unlock()
	due := make([]dueVoiceXPWorkerEntry, 0)
	for key, entry := range w.active {
		if entry.nextTick.After(cutoff) {
			continue
		}
		// Keep in-flight entries out of the next deadline calculation. Their real
		// next tick is scheduled from completion so slow Mongo calls never cause
		// catch-up bursts.
		entry.nextTick = cutoff.Add(w.interval)
		due = append(due, dueVoiceXPWorkerEntry{
			key: key, entry: entry,
			currentRoleIDs: append([]string(nil), entry.currentRoleIDs...),
			generation:     entry.generation,
		})
	}
	return due
}

func (w *VoiceXPWorker) nextDeadline() (time.Time, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	var earliest time.Time
	for _, entry := range w.active {
		if earliest.IsZero() || entry.nextTick.Before(earliest) {
			earliest = entry.nextTick
		}
	}
	return earliest, !earliest.IsZero()
}

func (w *VoiceXPWorker) finish(key string, expected *voiceXPWorkerEntry, generation uint64) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	entry, ok := w.active[key]
	if ok && entry == expected && entry.generation == generation {
		delete(w.active, key)
		entry.cancel()
		return true
	}
	return false
}

func (w *VoiceXPWorker) scheduleNext(key string, expected *voiceXPWorkerEntry) {
	w.mu.Lock()
	defer w.mu.Unlock()
	entry, ok := w.active[key]
	if ok && entry == expected {
		entry.nextTick = time.Now().Add(w.interval)
	}
}

func (w *VoiceXPWorker) notify() {
	select {
	case w.wake <- struct{}{}:
	default:
	}
}

func stopVoiceXPTimer(timer *time.Timer) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
}

func resetVoiceXPTimer(timer *time.Timer, delay time.Duration) {
	stopVoiceXPTimer(timer)
	if delay < 0 {
		delay = 0
	}
	timer.Reset(delay)
}

func (m VoiceEventModule) WithRuntimeWorker(interval time.Duration, logger *slog.Logger) VoiceEventModule {
	if logger == nil {
		logger = slog.Default()
	}
	m.logger = logger
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
	seen := make(map[string]struct{}, len(roleIDs))
	for _, roleID := range roleIDs {
		roleID = strings.TrimSpace(roleID)
		if roleID == "" {
			continue
		}
		if _, exists := seen[roleID]; exists {
			continue
		}
		seen[roleID] = struct{}{}
		roles = append(roles, roleID)
	}
	return roles
}
