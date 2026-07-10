package economy

import (
	"context"
	"sync"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type coinGameTimeoutScheduler interface {
	Schedule(key string, generation uint64, deadline time.Time, handler func(context.Context))
	Cancel(key string)
	Stop(ctx context.Context) error
}

type coinGameScheduledTimeout struct {
	generation uint64
	timer      *time.Timer
}

type coinGameTimeoutManager struct {
	mu      sync.Mutex
	clock   ports.Clock
	ctx     context.Context
	cancel  context.CancelFunc
	entries map[string]*coinGameScheduledTimeout
	wg      sync.WaitGroup
	stopped bool
}

func newCoinGameTimeoutManager(clock ports.Clock) *coinGameTimeoutManager {
	if clock == nil {
		clock = ports.SystemClock{}
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &coinGameTimeoutManager{
		clock:   clock,
		ctx:     ctx,
		cancel:  cancel,
		entries: map[string]*coinGameScheduledTimeout{},
	}
}

func (m *coinGameTimeoutManager) Schedule(key string, generation uint64, deadline time.Time, handler func(context.Context)) {
	if m == nil || key == "" || generation == 0 || deadline.IsZero() || handler == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.stopped {
		return
	}
	if current := m.entries[key]; current != nil {
		if current.generation > generation {
			return
		}
		m.stopEntryLocked(key, current)
	}
	entry := &coinGameScheduledTimeout{generation: generation}
	m.entries[key] = entry
	m.wg.Add(1)
	delay := deadline.Sub(m.clock.Now())
	if delay < 0 {
		delay = 0
	}
	entry.timer = time.AfterFunc(delay, func() {
		m.run(key, entry, handler)
	})
}

func (m *coinGameTimeoutManager) Cancel(key string) {
	if m == nil || key == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if entry := m.entries[key]; entry != nil {
		m.stopEntryLocked(key, entry)
	}
}

func (m *coinGameTimeoutManager) Stop(ctx context.Context) error {
	if m == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	m.mu.Lock()
	if !m.stopped {
		m.stopped = true
		m.cancel()
		for key, entry := range m.entries {
			m.stopEntryLocked(key, entry)
		}
	}
	m.mu.Unlock()

	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *coinGameTimeoutManager) run(key string, entry *coinGameScheduledTimeout, handler func(context.Context)) {
	m.mu.Lock()
	if m.entries[key] != entry {
		m.mu.Unlock()
		m.wg.Done()
		return
	}
	delete(m.entries, key)
	stopped := m.stopped
	ctx := m.ctx
	m.mu.Unlock()

	if !stopped {
		handler(ctx)
	}
	m.wg.Done()
}

func (m *coinGameTimeoutManager) stopEntryLocked(key string, entry *coinGameScheduledTimeout) {
	if m.entries[key] == entry {
		delete(m.entries, key)
	}
	if entry != nil && entry.timer != nil && entry.timer.Stop() {
		m.wg.Done()
	}
}
