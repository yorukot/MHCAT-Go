package stats

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

const LegacyStatsRenameInterval = 20 * time.Minute

type RenameService struct {
	Repository ports.StatsRenameRepository
	Channels   ports.DiscordChannelPort
	GuildStats ports.DiscordGuildStatsReader
	RoleStats  ports.DiscordRoleStatsReader
	Logger     *slog.Logger
}

type RenameResult struct {
	ConfigsChecked     int
	RoleConfigsChecked int
	ChannelsRenamed    int
	ChannelsSkipped    int
	CountersUpdated    int
}

type RenameWorker struct {
	service  RenameService
	interval time.Duration
	logger   *slog.Logger

	mu      sync.Mutex
	cancel  context.CancelFunc
	done    chan struct{}
	running bool
}

func NewRenameWorker(service RenameService, interval time.Duration, logger *slog.Logger) *RenameWorker {
	if interval <= 0 {
		interval = LegacyStatsRenameInterval
	}
	if logger == nil {
		logger = slog.Default()
	}
	service.Logger = logger
	return &RenameWorker{service: service, interval: interval, logger: logger}
}

func (s RenameService) RunOnce(ctx context.Context) (RenameResult, error) {
	if err := ctx.Err(); err != nil {
		return RenameResult{}, err
	}
	if s.Repository == nil || s.Channels == nil || s.GuildStats == nil || s.RoleStats == nil {
		return RenameResult{}, domain.ErrInvalidStatsConfigRequest
	}
	result, err := s.renameBaseStats(ctx)
	if err != nil {
		return result, err
	}
	roleResult, err := s.renameRoleStats(ctx)
	result.RoleConfigsChecked += roleResult.RoleConfigsChecked
	result.ChannelsRenamed += roleResult.ChannelsRenamed
	result.ChannelsSkipped += roleResult.ChannelsSkipped
	result.CountersUpdated += roleResult.CountersUpdated
	if err != nil {
		return result, err
	}
	return result, ctx.Err()
}

func (s RenameService) renameBaseStats(ctx context.Context) (RenameResult, error) {
	configs, err := s.Repository.ListStatsConfigs(ctx)
	if err != nil {
		return RenameResult{}, err
	}
	var result RenameResult
	for _, config := range configs {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		if config.GuildID == "" {
			result.ChannelsSkipped++
			continue
		}
		result.ConfigsChecked++
		snapshot, err := s.GuildStats.GuildStats(ctx, config.GuildID)
		if err != nil {
			s.logWarn(ctx, "skip stats config rename after guild stats failure", "guild_id", config.GuildID, "error", err.Error())
			result.ChannelsSkipped += configuredBaseStatsChannelCount(config)
			continue
		}
		update := domain.StatsConfigCounterUpdate{}
		s.applyBaseCounter(ctx, &result, config.GuildID, config.MemberNumberID, config.MemberNumberName, snapshot.MemberCount, func(value *string) {
			update.MemberNumberName = value
		})
		s.applyBaseCounter(ctx, &result, config.GuildID, config.UserNumberID, config.UserNumberName, snapshot.UserCount, func(value *string) {
			update.UserNumberName = value
		})
		s.applyBaseCounter(ctx, &result, config.GuildID, config.BotNumberID, config.BotNumberName, snapshot.BotCount, func(value *string) {
			update.BotNumberName = value
		})
		s.applyBaseCounter(ctx, &result, config.GuildID, config.ChannelNumberID, config.ChannelNumberName, snapshot.ChannelCount, func(value *string) {
			update.ChannelNumberName = value
		})
		s.applyBaseCounter(ctx, &result, config.GuildID, config.TextNumberID, config.TextNumberName, snapshot.TextChannelCount, func(value *string) {
			update.TextNumberName = value
		})
		s.applyBaseCounter(ctx, &result, config.GuildID, config.VoiceNumberID, config.VoiceNumberName, snapshot.VoiceChannelCount, func(value *string) {
			update.VoiceNumberName = value
		})
		if !update.IsZero() {
			if err := s.Repository.UpdateStatsConfigCounters(ctx, config.GuildID, update); err != nil {
				return result, err
			}
			result.CountersUpdated++
		}
	}
	return result, ctx.Err()
}

func (s RenameService) renameRoleStats(ctx context.Context) (RenameResult, error) {
	configs, err := s.Repository.ListStatsRoleConfigs(ctx)
	if err != nil {
		return RenameResult{}, err
	}
	var result RenameResult
	for _, config := range configs {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		if config.GuildID == "" || config.RoleID == "" || config.ChannelID == "" {
			result.ChannelsSkipped++
			continue
		}
		result.RoleConfigsChecked++
		snapshot, err := s.RoleStats.RoleStats(ctx, config.GuildID, config.RoleID)
		if err != nil {
			s.logWarn(ctx, "skip stats role rename after role stats failure", "guild_id", config.GuildID, "role_id", config.RoleID, "error", err.Error())
			result.ChannelsSkipped++
			continue
		}
		currentValue := domain.StatsRoleCounterValue(snapshot.MemberCount)
		renamed, ok, err := s.renameConfiguredChannel(ctx, config.GuildID, config.ChannelID, config.ChannelName, currentValue)
		if err != nil {
			s.logWarn(ctx, "skip stats role channel rename after discord failure", "guild_id", config.GuildID, "channel_id", config.ChannelID, "error", err.Error())
			result.ChannelsSkipped++
			continue
		}
		if !ok {
			result.ChannelsSkipped++
			continue
		}
		if renamed {
			result.ChannelsRenamed++
		}
		if strings.TrimSpace(config.ChannelName) != currentValue {
			if err := s.Repository.UpdateStatsRoleConfigCounter(ctx, config.GuildID, config.RoleID, currentValue); err != nil {
				return result, err
			}
			result.CountersUpdated++
		}
	}
	return result, ctx.Err()
}

func (s RenameService) applyBaseCounter(ctx context.Context, result *RenameResult, guildID string, channelID string, oldValue string, count int, set func(*string)) {
	channelID = strings.TrimSpace(channelID)
	if channelID == "" {
		return
	}
	currentValue := strconv.Itoa(count)
	renamed, ok, err := s.renameConfiguredChannel(ctx, guildID, channelID, oldValue, currentValue)
	if err != nil {
		s.logWarn(ctx, "skip stats channel rename after discord failure", "guild_id", guildID, "channel_id", channelID, "error", err.Error())
		result.ChannelsSkipped++
		return
	}
	if !ok {
		result.ChannelsSkipped++
		return
	}
	if renamed {
		result.ChannelsRenamed++
	}
	if strings.TrimSpace(oldValue) != currentValue {
		set(domain.StatsCounterValue(count))
	}
}

func (s RenameService) renameConfiguredChannel(ctx context.Context, guildID string, channelID string, oldValue string, currentValue string) (bool, bool, error) {
	channel, err := s.Channels.FindCachedChannelByID(ctx, guildID, channelID)
	if err != nil {
		if errors.Is(err, ports.ErrChannelNotFound) {
			return false, false, nil
		}
		return false, false, err
	}
	nextName := legacyStatsRenamedChannelName(channel.Name, oldValue, currentValue)
	if nextName == channel.Name {
		return false, true, ctx.Err()
	}
	if _, err := s.Channels.RenameChannel(ctx, guildID, channelID, nextName); err != nil {
		if errors.Is(err, ports.ErrChannelNotFound) {
			return false, false, nil
		}
		return false, false, err
	}
	return true, true, ctx.Err()
}

func legacyStatsRenamedChannelName(name string, oldValue string, currentValue string) string {
	oldValue = strings.TrimSpace(oldValue)
	currentValue = strings.TrimSpace(currentValue)
	if currentValue == "" {
		return name
	}
	if oldValue == "" || !strings.Contains(name, oldValue) {
		return currentValue
	}
	return strings.Replace(name, oldValue, currentValue, 1)
}

func configuredBaseStatsChannelCount(config domain.StatsConfig) int {
	count := 0
	for _, channelID := range []string{
		config.MemberNumberID,
		config.UserNumberID,
		config.BotNumberID,
		config.ChannelNumberID,
		config.TextNumberID,
		config.VoiceNumberID,
	} {
		if strings.TrimSpace(channelID) != "" {
			count++
		}
	}
	return count
}

func (s RenameService) logWarn(ctx context.Context, message string, args ...any) {
	logger := s.Logger
	if logger == nil {
		logger = slog.Default()
	}
	logger.WarnContext(ctx, message, args...)
}

func (w *RenameWorker) Start(ctx context.Context) bool {
	if w == nil {
		return false
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.running {
		return false
	}
	runCtx, cancel := context.WithCancel(ctx)
	w.cancel = cancel
	w.done = make(chan struct{})
	w.running = true
	go w.loop(runCtx)
	return true
}

func (w *RenameWorker) Stop(ctx context.Context) error {
	if w == nil {
		return nil
	}
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return nil
	}
	cancel := w.cancel
	done := w.done
	w.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (w *RenameWorker) loop(ctx context.Context) {
	defer func() {
		w.mu.Lock()
		w.running = false
		w.cancel = nil
		close(w.done)
		w.mu.Unlock()
	}()
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.runOnceSafe(ctx)
		}
	}
}

func (w *RenameWorker) runOnceSafe(ctx context.Context) {
	result, err := w.service.RunOnce(ctx)
	if err != nil {
		if ctx.Err() == nil {
			w.logger.WarnContext(ctx, "stats rename worker failed", "error", err.Error())
		}
		return
	}
	w.logger.InfoContext(ctx, "stats rename worker completed",
		"configs", result.ConfigsChecked,
		"role_configs", result.RoleConfigsChecked,
		"renamed", result.ChannelsRenamed,
		"skipped", result.ChannelsSkipped,
		"counter_updates", result.CountersUpdated,
	)
}
