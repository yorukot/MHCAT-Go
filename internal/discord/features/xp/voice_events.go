package xp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/xp"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
)

const (
	voiceXPReconcileTimeout     = 15 * time.Second
	voiceXPReconcileConcurrency = 4
)

type voiceXPGuildCoordinator struct {
	mu        sync.Mutex
	guilds    map[string]chan struct{}
	snapshots chan struct{}
}

func newVoiceXPGuildCoordinator() *voiceXPGuildCoordinator {
	return &voiceXPGuildCoordinator{
		guilds:    make(map[string]chan struct{}),
		snapshots: make(chan struct{}, voiceXPReconcileConcurrency),
	}
}

func (c *voiceXPGuildCoordinator) lockGuild(ctx context.Context, guildID string) (func(), error) {
	c.mu.Lock()
	lock := c.guilds[guildID]
	if lock == nil {
		lock = make(chan struct{}, 1)
		c.guilds[guildID] = lock
	}
	c.mu.Unlock()
	select {
	case lock <- struct{}{}:
		return func() { <-lock }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *voiceXPGuildCoordinator) acquireSnapshot(ctx context.Context) (func(), error) {
	select {
	case c.snapshots <- struct{}{}:
		return func() { <-c.snapshots }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (m VoiceEventModule) VoiceStateHandler() events.Handler {
	coordinator := m.coordinator
	if coordinator == nil {
		coordinator = newVoiceXPGuildCoordinator()
	}
	return func(ctx context.Context, event events.Event) error {
		if event.Type != events.TypeVoiceState || event.VoiceState == nil {
			return nil
		}
		voice := event.VoiceState
		guildID := strings.TrimSpace(voice.GuildID)
		if guildID == "" {
			guildID = strings.TrimSpace(event.GuildID)
		}
		userID := strings.TrimSpace(voice.UserID)
		if userID == "" {
			userID = strings.TrimSpace(event.UserID)
		}
		isBot := event.IsBot
		if event.Member != nil {
			isBot = event.Member.IsBot
			if event.Member.UserID != "" {
				userID = strings.TrimSpace(event.Member.UserID)
			}
		}
		channelID := strings.TrimSpace(voice.ChannelID)
		beforeChannelID := strings.TrimSpace(voice.BeforeChannel)
		if guildID == "" || userID == "" || isBot || channelID == beforeChannelID {
			return nil
		}
		unlock, err := coordinator.lockGuild(ctx, guildID)
		if err != nil {
			return err
		}
		defer unlock()
		if channelID == "" {
			if m.worker != nil {
				m.worker.Stop(guildID, userID)
			}
			return m.service.Leave(ctx, guildID, userID)
		}
		if err := m.service.Join(ctx, guildID, userID); err != nil {
			return err
		}
		if m.worker != nil {
			if event.Member != nil {
				m.worker.StartOrUpdate(guildID, userID, voiceEventRoleIDs(event.Member))
			} else {
				m.worker.Start(guildID, userID, nil)
			}
		}
		return ctx.Err()
	}
}

func (m VoiceEventModule) GuildAvailableHandler() events.Handler {
	coordinator := m.coordinator
	if coordinator == nil {
		coordinator = newVoiceXPGuildCoordinator()
	}
	logger := m.logger
	if logger == nil {
		logger = slog.Default()
	}
	return func(ctx context.Context, event events.Event) error {
		if event.Type != events.TypeGuildAvailable || m.voiceStates == nil {
			return nil
		}
		guildID := strings.TrimSpace(event.GuildID)
		if guildID == "" {
			return nil
		}
		unlock, err := coordinator.lockGuild(ctx, guildID)
		if err != nil {
			return m.failVoiceXPReconciliation(ctx, logger, guildID, "guild_lock", err)
		}
		defer unlock()
		releaseSlot, err := coordinator.acquireSnapshot(ctx)
		if err != nil {
			return m.failVoiceXPReconciliation(ctx, logger, guildID, "concurrency_limit", err)
		}
		defer releaseSlot()
		reconcileCtx, cancel := context.WithTimeout(ctx, voiceXPReconcileTimeout)
		defer cancel()

		states, err := m.voiceStates.GuildVoiceStates(reconcileCtx, guildID)
		if err != nil {
			return m.failVoiceXPReconciliation(reconcileCtx, logger, guildID, "discord_cache", err)
		}
		activeStates := activeVoiceXPStates(states)
		activeUserIDs := make([]string, 0, len(activeStates))
		for _, state := range activeStates {
			activeUserIDs = append(activeUserIDs, state.UserID)
		}
		if err := m.service.Reconcile(reconcileCtx, guildID, activeUserIDs); err != nil {
			return m.failVoiceXPReconciliation(reconcileCtx, logger, guildID, "mongo", err)
		}
		started, stopped := 0, 0
		if m.worker != nil {
			started, stopped = m.worker.ReconcileGuild(guildID, activeStates)
		}
		logger.DebugContext(reconcileCtx, "voice xp guild reconciled",
			"guild_id", guildID,
			"active_users", len(activeStates),
			"workers_started", started,
			"workers_stopped", stopped,
		)
		return reconcileCtx.Err()
	}
}

func (m VoiceEventModule) failVoiceXPReconciliation(ctx context.Context, logger *slog.Logger, guildID string, stage string, err error) error {
	stopped := 0
	if m.worker != nil {
		stopped = m.worker.StopGuild(guildID)
	}
	logger.WarnContext(ctx, "voice xp guild reconciliation failed",
		"guild_id", guildID,
		"stage", stage,
		"workers_stopped", stopped,
		"error", err.Error(),
	)
	return fmt.Errorf("reconcile voice xp guild %s at %s: %w", guildID, stage, err)
}

func activeVoiceXPStates(states []ports.DiscordVoiceState) []ports.DiscordVoiceState {
	active := make(map[string]ports.DiscordVoiceState, len(states))
	for _, state := range states {
		state.UserID = strings.TrimSpace(state.UserID)
		state.ChannelID = strings.TrimSpace(state.ChannelID)
		if state.UserID == "" || state.ChannelID == "" || state.IsBot {
			continue
		}
		state.RoleIDs = trimmedRoleIDs(state.RoleIDs)
		active[state.UserID] = state
	}
	userIDs := make([]string, 0, len(active))
	for userID := range active {
		userIDs = append(userIDs, userID)
	}
	sort.Strings(userIDs)
	out := make([]ports.DiscordVoiceState, 0, len(userIDs))
	for _, userID := range userIDs {
		out = append(out, active[userID])
	}
	return out
}

func voiceEventRoleIDs(member *events.Member) []string {
	if member == nil {
		return nil
	}
	return member.RoleIDs
}

func (m VoiceEventModule) TickVoiceXP(ctx context.Context, guildID string, userID string, currentRoleIDs []string) (coreservice.VoiceAccrualResult, error) {
	result, err := m.accrual.Tick(ctx, guildID, userID)
	if err != nil {
		return result, err
	}
	if !result.Active || !result.Leveled {
		return result, ctx.Err()
	}
	var firstErr error
	announcementSent, err := m.sendLevelUpAnnouncement(ctx, result.Profile.GuildID, result.Profile.UserID, result.Profile.Level)
	if err != nil {
		firstErr = err
	}
	if err := m.applyRewardRoles(ctx, result.Profile.GuildID, result.Profile.UserID, result.Profile.Level, currentRoleIDs); err != nil && firstErr == nil {
		firstErr = err
	}
	if announcementSent {
		err = m.applyCoinReward(ctx, result.Profile.GuildID, result.Profile.UserID, result.Profile.Level)
	}
	if err != nil && firstErr == nil {
		firstErr = err
	}
	if firstErr != nil {
		return result, firstErr
	}
	return result, ctx.Err()
}

func (m VoiceEventModule) sendLevelUpAnnouncement(ctx context.Context, guildID string, userID string, level int64) (bool, error) {
	if m.configs == nil || m.messages == nil {
		return false, ctx.Err()
	}
	config, err := m.configs.GetVoiceXPConfig(ctx, guildID)
	if errors.Is(err, ports.ErrVoiceXPConfigMissing) {
		return false, ctx.Err()
	}
	if err != nil {
		return false, err
	}
	channelID := strings.TrimSpace(config.ChannelID)
	if channelID == "" {
		return false, ctx.Err()
	}
	channelName := channelID
	if m.channels != nil {
		channel, err := m.channels.FindChannelByID(ctx, guildID, channelID)
		if errors.Is(err, ports.ErrChannelNotFound) {
			return false, m.sendVoiceOwnerFallback(ctx, guildID, ":x: 有人的語音頻道等級升級了，但升等頻道已經被刪除了!")
		}
		if err != nil {
			return false, err
		}
		if strings.TrimSpace(channel.Name) != "" {
			channelName = strings.TrimSpace(channel.Name)
		}
	}
	content := coreservice.LegacyVoiceXPLevelUpAnnouncement(config.Message, level, userID)
	_, err = m.messages.SendMessage(ctx, channelID, ports.OutboundMessage{
		Content:         content,
		AllowedMentions: ports.AllowedMentions{UserIDs: []string{strings.TrimSpace(userID)}},
	})
	if err != nil {
		if m.direct != nil && m.guilds != nil {
			return false, m.sendVoiceOwnerFallback(ctx, guildID, ":x: 有人的語音頻道等級升級了，但是我沒有權限在"+channelName+"發送消息!\n因為你是該伺服器擁有者，所以我找你報告: P")
		}
		return false, err
	}
	return true, ctx.Err()
}

func (m VoiceEventModule) applyRewardRoles(ctx context.Context, guildID string, userID string, level int64, currentRoleIDs []string) error {
	if m.rewardRoles.Repository == nil || m.rewardRoles.RolePort == nil {
		return ctx.Err()
	}
	return m.rewardRoles.ApplyLevelUp(ctx, guildID, userID, level, currentRoleIDs)
}

func (m VoiceEventModule) applyCoinReward(ctx context.Context, guildID string, userID string, level int64) error {
	if m.coinRewards.Repository == nil {
		return ctx.Err()
	}
	_, err := m.coinRewards.ApplyLevelUp(ctx, guildID, userID, level)
	return err
}

func (m VoiceEventModule) sendVoiceOwnerFallback(ctx context.Context, guildID string, content string) error {
	if m.direct == nil || m.guilds == nil {
		return ctx.Err()
	}
	guild, err := m.guilds.GuildInfo(ctx, strings.TrimSpace(guildID))
	if err != nil {
		return err
	}
	ownerID := strings.TrimSpace(guild.OwnerID)
	if ownerID == "" {
		return ctx.Err()
	}
	_, err = m.direct.SendDirectMessage(ctx, ownerID, ports.OutboundMessage{
		Content:         content,
		AllowedMentions: ports.AllowedMentions{},
	})
	return err
}
