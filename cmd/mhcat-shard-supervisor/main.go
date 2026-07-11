package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	defaultBotPath           = "/usr/local/bin/mhcat-bot"
	defaultSpawnDelay        = 6 * time.Second
	defaultRestartDelay      = 2 * time.Second
	defaultRestartMaxDelay   = 2 * time.Minute
	defaultRestartResetAfter = 5 * time.Minute
	defaultStopTimeout       = 30 * time.Second
	defaultForceKillWait     = 5 * time.Second
)

type supervisorConfig struct {
	BotPath           string
	ShardCount        int
	SpawnDelay        time.Duration
	RestartDelay      time.Duration
	RestartMaxDelay   time.Duration
	RestartResetAfter time.Duration
	StopTimeout       time.Duration
}

type childExit struct {
	shardID int
	err     error
	uptime  time.Duration
}

type supervisor struct {
	cfg    supervisorConfig
	stdout io.Writer
	stderr io.Writer

	mu       sync.Mutex
	children map[int]*exec.Cmd
	exits    chan childExit
}

type lockedWriter struct {
	mu     *sync.Mutex
	writer io.Writer
}

func main() {
	cfg, err := loadSupervisorConfig(os.LookupEnv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mhcat-shard-supervisor: %v\n", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	s := newSupervisor(cfg, os.Stdout, os.Stderr)
	if err := s.run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "mhcat-shard-supervisor: %v\n", err)
		os.Exit(1)
	}
}

func loadSupervisorConfig(lookup func(string) (string, bool)) (supervisorConfig, error) {
	count, err := envInt(lookup, "MHCAT_DISCORD_SHARD_COUNT", 1)
	if err != nil {
		return supervisorConfig{}, err
	}
	if count <= 0 {
		return supervisorConfig{}, errors.New("MHCAT_DISCORD_SHARD_COUNT must be positive")
	}
	spawnDelay, err := envDuration(lookup, "MHCAT_DISCORD_SHARD_SPAWN_DELAY", defaultSpawnDelay)
	if err != nil {
		return supervisorConfig{}, err
	}
	restartDelay, err := envDuration(lookup, "MHCAT_DISCORD_SHARD_RESTART_DELAY", defaultRestartDelay)
	if err != nil {
		return supervisorConfig{}, err
	}
	restartMaxDelay, err := envDuration(lookup, "MHCAT_DISCORD_SHARD_RESTART_MAX_DELAY", defaultRestartMaxDelay)
	if err != nil {
		return supervisorConfig{}, err
	}
	restartResetAfter, err := envDuration(lookup, "MHCAT_DISCORD_SHARD_RESTART_RESET_AFTER", defaultRestartResetAfter)
	if err != nil {
		return supervisorConfig{}, err
	}
	stopTimeout, err := envDuration(lookup, "MHCAT_DISCORD_SHARD_STOP_TIMEOUT", defaultStopTimeout)
	if err != nil {
		return supervisorConfig{}, err
	}
	if spawnDelay < 0 || restartDelay <= 0 || restartMaxDelay < restartDelay || restartResetAfter <= 0 || stopTimeout <= 0 {
		return supervisorConfig{}, errors.New("shard spawn delay must be non-negative; restart delays, reset window, and stop timeout must be positive; restart max delay must not be shorter than restart delay")
	}
	botPath := envString(lookup, "MHCAT_BOT_PATH", defaultBotPath)
	if strings.TrimSpace(botPath) == "" {
		return supervisorConfig{}, errors.New("MHCAT_BOT_PATH must not be empty")
	}
	return supervisorConfig{
		BotPath:           botPath,
		ShardCount:        count,
		SpawnDelay:        spawnDelay,
		RestartDelay:      restartDelay,
		RestartMaxDelay:   restartMaxDelay,
		RestartResetAfter: restartResetAfter,
		StopTimeout:       stopTimeout,
	}, nil
}

func newSupervisor(cfg supervisorConfig, stdout io.Writer, stderr io.Writer) *supervisor {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
	outputMu := &sync.Mutex{}
	return &supervisor{
		cfg:      cfg,
		stdout:   lockedWriter{mu: outputMu, writer: stdout},
		stderr:   lockedWriter{mu: outputMu, writer: stderr},
		children: make(map[int]*exec.Cmd),
		exits:    make(chan childExit, cfg.ShardCount),
	}
}

func (w lockedWriter) Write(payload []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.writer.Write(payload)
}

func (s *supervisor) run(ctx context.Context) error {
	restartDelays := make(map[int]time.Duration, s.cfg.ShardCount)
	for shardID := 0; shardID < s.cfg.ShardCount; shardID++ {
		if err := s.start(shardID); err != nil {
			return errors.Join(err, s.shutdown())
		}
		if shardID+1 < s.cfg.ShardCount && !sleepContext(ctx, s.cfg.SpawnDelay) {
			return s.shutdown()
		}
	}

	for {
		select {
		case <-ctx.Done():
			return s.shutdown()
		case exited := <-s.exits:
			s.mu.Lock()
			delete(s.children, exited.shardID)
			s.mu.Unlock()
			restartDelay := nextRestartDelay(
				s.cfg.RestartDelay,
				s.cfg.RestartMaxDelay,
				restartDelays[exited.shardID],
				exited.uptime,
				s.cfg.RestartResetAfter,
			)
			restartDelays[exited.shardID] = restartDelay
			fmt.Fprintf(s.stderr, "shard %d exited after %s: %v; restarting in %s\n", exited.shardID, exited.uptime.Round(time.Millisecond), exited.err, restartDelay)
			if !sleepContext(ctx, restartDelay) {
				return s.shutdown()
			}
			if err := s.start(exited.shardID); err != nil {
				return errors.Join(err, s.shutdown())
			}
		}
	}
}

func (s *supervisor) start(shardID int) error {
	startedAt := time.Now()
	cmd := exec.Command(s.cfg.BotPath)
	cmd.Stdout = s.stdout
	cmd.Stderr = s.stderr
	cmd.Env = shardEnvironment(os.Environ(), shardID, s.cfg.ShardCount)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start shard %d: %w", shardID, err)
	}
	s.mu.Lock()
	s.children[shardID] = cmd
	s.mu.Unlock()
	fmt.Fprintf(s.stdout, "started shard %d/%d pid=%d\n", shardID, s.cfg.ShardCount, cmd.Process.Pid)
	go func() {
		s.exits <- childExit{shardID: shardID, err: cmd.Wait(), uptime: time.Since(startedAt)}
	}()
	return nil
}

func nextRestartDelay(initial time.Duration, maximum time.Duration, previous time.Duration, uptime time.Duration, resetAfter time.Duration) time.Duration {
	if uptime >= resetAfter {
		previous = 0
	}
	if previous <= 0 {
		return initial
	}
	if previous >= maximum || previous > maximum/2 {
		return maximum
	}
	return previous * 2
}

func (s *supervisor) shutdown() error {
	s.stopAll()
	if s.waitForChildren(s.cfg.StopTimeout) {
		return nil
	}
	s.killAll()
	if !s.waitForChildren(defaultForceKillWait) {
		return errors.New("shard shutdown timed out; killed children did not exit")
	}
	return errors.New("shard shutdown timed out; remaining children killed")
}

func (s *supervisor) waitForChildren(timeout time.Duration) bool {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	for {
		s.mu.Lock()
		remaining := len(s.children)
		s.mu.Unlock()
		if remaining == 0 {
			return true
		}
		select {
		case exited := <-s.exits:
			s.mu.Lock()
			delete(s.children, exited.shardID)
			s.mu.Unlock()
		case <-deadline.C:
			return false
		}
	}
}

func (s *supervisor) stopAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, child := range s.children {
		if child.Process != nil {
			_ = child.Process.Signal(syscall.SIGTERM)
		}
	}
}

func (s *supervisor) killAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, child := range s.children {
		if child.Process != nil {
			_ = child.Process.Kill()
		}
	}
}

func shardEnvironment(base []string, shardID int, shardCount int) []string {
	result := make([]string, 0, len(base)+3)
	leaseOwner := ""
	for _, value := range base {
		if strings.HasPrefix(value, "MHCAT_SCHEDULER_LEASE_OWNER=") {
			leaseOwner = strings.TrimPrefix(value, "MHCAT_SCHEDULER_LEASE_OWNER=")
			continue
		}
		if strings.HasPrefix(value, "MHCAT_DISCORD_SHARD_ID=") || strings.HasPrefix(value, "MHCAT_DISCORD_SHARD_COUNT=") {
			continue
		}
		result = append(result, value)
	}
	result = append(result,
		"MHCAT_DISCORD_SHARD_ID="+strconv.Itoa(shardID),
		"MHCAT_DISCORD_SHARD_COUNT="+strconv.Itoa(shardCount),
	)
	if strings.TrimSpace(leaseOwner) != "" {
		result = append(result, "MHCAT_SCHEDULER_LEASE_OWNER="+leaseOwner+"-shard-"+strconv.Itoa(shardID))
	}
	return result
}

func envString(lookup func(string) (string, bool), key string, fallback string) string {
	value, ok := lookup(key)
	if !ok {
		return fallback
	}
	return strings.TrimSpace(value)
}

func envInt(lookup func(string) (string, bool), key string, fallback int) (int, error) {
	value, ok := lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
}

func envDuration(lookup func(string) (string, bool), key string, fallback time.Duration) (time.Duration, error) {
	value, ok := lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
}

func sleepContext(ctx context.Context, delay time.Duration) bool {
	if delay <= 0 {
		return ctx.Err() == nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
