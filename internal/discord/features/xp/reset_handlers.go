package xp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	xpResetWarningContent = ":warning: | 一但刪除，___**將無法復原**___，如確定要還原請於60秒內輸入`^確認^`(只有一次機會)!!!"
	xpResetConfirmContent = "^確認^"
	xpResetSuccessColor   = 0x53FF53
)

var defaultResetConfirmationStore = newResetConfirmationStore(ports.SystemClock{}, time.Minute)

type resetKind string

const (
	resetKindText  resetKind = "text"
	resetKindVoice resetKind = "voice"
)

func (m ResetModule) ResetHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		owner, err := m.actorIsGuildOwner(ctx, interaction)
		if err != nil {
			return responder.EditOriginal(ctx, textXPUnknownError(err))
		}
		if !owner {
			return responder.EditOriginal(ctx, textXPErrorMessage("你必須擁有`服主`才能使用"))
		}

		switch interaction.Subcommand {
		case "重製個人聊天經驗":
			userID := xpResetTargetUserID(interaction)
			if err := m.service.ResetTextProfile(ctx, interaction.Actor.GuildID, userID); err != nil {
				return responder.EditOriginal(ctx, xpResetProfileErrorMessage(err))
			}
			if err := responder.EditOriginal(ctx, xpResetProfileSuccessMessage(userID, "聊天")); err != nil {
				return err
			}
		case "重製個人語音經驗":
			userID := xpResetTargetUserID(interaction)
			if err := m.service.ResetVoiceProfile(ctx, interaction.Actor.GuildID, userID); err != nil {
				return responder.EditOriginal(ctx, xpResetProfileErrorMessage(err))
			}
			if err := responder.EditOriginal(ctx, xpResetProfileSuccessMessage(userID, "語音")); err != nil {
				return err
			}
		case "聊天經驗重製":
			m.confirmations.Put(resetConfirmation{
				GuildID:   interaction.Actor.GuildID,
				ChannelID: interaction.ChannelID,
				UserID:    interaction.Actor.UserID,
				Kind:      resetKindText,
			})
			if err := responder.EditOriginal(ctx, xpResetWarningMessage()); err != nil {
				return err
			}
		case "語音經驗重製":
			m.confirmations.Put(resetConfirmation{
				GuildID:   interaction.Actor.GuildID,
				ChannelID: interaction.ChannelID,
				UserID:    interaction.Actor.UserID,
				Kind:      resetKindVoice,
			})
			if err := responder.EditOriginal(ctx, xpResetWarningMessage()); err != nil {
				return err
			}
		default:
			return responder.EditOriginal(ctx, textXPErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
		}
		return m.track(ctx, interaction)
	}
}

func xpResetTargetUserID(interaction interactions.Interaction) string {
	if userID := firstOption(interaction, "使用者"); userID != "" {
		return userID
	}
	return strings.TrimSpace(interaction.Actor.UserID)
}

func (m ResetModule) ConfirmationHandler() events.Handler {
	return func(ctx context.Context, event events.Event) error {
		if event.Type != events.TypeMessageCreate || event.IsBot {
			return nil
		}
		if m.messages == nil || m.confirmations == nil {
			return nil
		}
		confirmation, ok := m.confirmations.Take(event)
		if !ok {
			return nil
		}
		if event.Content != xpResetConfirmContent {
			_, err := m.messages.SendMessage(ctx, event.ChannelID, xpResetWrongConfirmationOutbound())
			return err
		}
		var err error
		switch confirmation.Kind {
		case resetKindText:
			err = m.service.ResetTextGuild(ctx, confirmation.GuildID)
		case resetKindVoice:
			err = m.service.ResetVoiceGuild(ctx, confirmation.GuildID)
		default:
			err = fmt.Errorf("unknown xp reset kind %q", confirmation.Kind)
		}
		if err != nil {
			_, sendErr := m.messages.SendMessage(ctx, event.ChannelID, xpResetGuildErrorOutbound(err, confirmation.Kind))
			if sendErr != nil {
				return sendErr
			}
			return nil
		}
		_, err = m.messages.SendMessage(ctx, event.ChannelID, xpResetGuildSuccessOutbound(confirmation.Kind))
		return err
	}
}

func (m ResetModule) actorIsGuildOwner(ctx context.Context, interaction interactions.Interaction) (bool, error) {
	if m.guilds == nil {
		return false, fmt.Errorf("xp reset requires discord guild info provider")
	}
	info, err := m.guilds.GuildInfo(ctx, interaction.Actor.GuildID)
	if err != nil {
		return false, err
	}
	ownerID := strings.TrimSpace(info.OwnerID)
	if ownerID == "" {
		return false, fmt.Errorf("guild owner id is empty")
	}
	return strings.TrimSpace(interaction.Actor.UserID) == ownerID, nil
}

func xpResetProfileErrorMessage(err error) responses.Message {
	if errors.Is(err, ports.ErrTextXPProfileMissing) || errors.Is(err, ports.ErrVoiceXPProfileMissing) {
		return textXPErrorMessage("這位使用者還沒有任何的經驗值喔!")
	}
	return textXPUnknownError(err)
}

func xpResetProfileSuccessMessage(userID string, label string) responses.Message {
	return responses.Message{
		Content:         fmt.Sprintf("%s | 成功清除<@%s>的%s經驗", doneEmoji, strings.TrimSpace(userID), label),
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func xpResetWarningMessage() responses.Message {
	return responses.Message{
		Content:         xpResetWarningContent,
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func xpResetWrongConfirmationOutbound() ports.OutboundMessage {
	return ports.OutboundMessage{
		Embeds: []ports.OutboundEmbed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | 你輸入了錯誤的確認!因此視為取消還原",
			Color: textXPErrorColor,
		}},
		AllowedMentions: ports.AllowedMentions{},
	}
}

func xpResetGuildErrorOutbound(err error, kind resetKind) ports.OutboundMessage {
	message := "很抱歉，出現了未知的錯誤，請重試!"
	if kind == resetKindText && errors.Is(err, ports.ErrTextXPProfileMissing) {
		message = "伺服器沒有任何聊天經驗的資料!"
	}
	if kind == resetKindVoice && errors.Is(err, ports.ErrVoiceXPProfileMissing) {
		message = "伺服器沒有任何語音經驗的資料!"
	}
	return ports.OutboundMessage{
		Embeds: []ports.OutboundEmbed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | " + message,
			Color: textXPErrorColor,
		}},
		AllowedMentions: ports.AllowedMentions{},
	}
}

func xpResetGuildSuccessOutbound(kind resetKind) ports.OutboundMessage {
	label := "聊天"
	if kind == resetKindVoice {
		label = "語音"
	}
	return ports.OutboundMessage{
		Embeds: []ports.OutboundEmbed{{
			Title: deleteEmoji + "成功刪除伺服器內所有" + label + "經驗",
			Color: xpResetSuccessColor,
		}},
		AllowedMentions: ports.AllowedMentions{},
	}
}

func (m ResetModule) track(ctx context.Context, interaction interactions.Interaction) error {
	if m.usage == nil {
		return nil
	}
	return m.usage.TrackCommand(ctx, ports.UsageEvent{
		CommandName: XPResetCommandName,
		UserID:      interaction.Actor.UserID,
		GuildID:     interaction.Actor.GuildID,
		Feature:     "xp-reset",
	})
}

type resetConfirmation struct {
	GuildID   string
	ChannelID string
	UserID    string
	Kind      resetKind
	ExpiresAt time.Time
}

type resetConfirmationKey struct {
	GuildID   string
	ChannelID string
	UserID    string
}

type resetConfirmationStore struct {
	mu      sync.Mutex
	pending map[resetConfirmationKey]resetConfirmation
	clock   ports.Clock
	ttl     time.Duration
}

func newResetConfirmationStore(clock ports.Clock, ttl time.Duration) *resetConfirmationStore {
	if clock == nil {
		clock = ports.SystemClock{}
	}
	if ttl <= 0 {
		ttl = time.Minute
	}
	return &resetConfirmationStore{
		pending: map[resetConfirmationKey]resetConfirmation{},
		clock:   clock,
		ttl:     ttl,
	}
}

func (s *resetConfirmationStore) Put(confirmation resetConfirmation) {
	if s == nil {
		return
	}
	confirmation.GuildID = strings.TrimSpace(confirmation.GuildID)
	confirmation.ChannelID = strings.TrimSpace(confirmation.ChannelID)
	confirmation.UserID = strings.TrimSpace(confirmation.UserID)
	if confirmation.GuildID == "" || confirmation.ChannelID == "" || confirmation.UserID == "" {
		return
	}
	confirmation.ExpiresAt = s.clock.Now().Add(s.ttl)
	key := resetConfirmationKey{GuildID: confirmation.GuildID, ChannelID: confirmation.ChannelID, UserID: confirmation.UserID}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pending[key] = confirmation
}

func (s *resetConfirmationStore) Take(event events.Event) (resetConfirmation, bool) {
	if s == nil {
		return resetConfirmation{}, false
	}
	key := resetConfirmationKey{
		GuildID:   strings.TrimSpace(event.GuildID),
		ChannelID: strings.TrimSpace(event.ChannelID),
		UserID:    strings.TrimSpace(event.UserID),
	}
	if key.GuildID == "" || key.ChannelID == "" || key.UserID == "" {
		return resetConfirmation{}, false
	}
	now := event.CreatedAt
	if now.IsZero() {
		now = s.clock.Now()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	confirmation, ok := s.pending[key]
	if !ok {
		return resetConfirmation{}, false
	}
	delete(s.pending, key)
	if !now.Before(confirmation.ExpiresAt) {
		return resetConfirmation{}, false
	}
	return confirmation, true
}
