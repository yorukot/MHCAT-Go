package economy

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	coinResetWarningContent = ":warning: | 一但重製，___**將無法復原**___，如確定要還原請於60秒內輸入`^確認^`(只有一次機會)!!!"
	coinResetConfirmContent = "^確認^"
	coinResetOptionDivisor  = "除以多少"
	coinResetDeleteEmoji    = "<:trashbin:995991389043163257>"
	coinResetErrorPrefix    = "<a:Discord_AnimatedNo:1015989839809757295> | "
	coinResetErrorColor     = 0xED4245
	coinResetSuccessColor   = 0x57F287
)

var defaultCoinResetConfirmationStore = newCoinResetConfirmationStore(ports.SystemClock{}, time.Minute)

func (m Module) CoinResetHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		owner, err := m.coinResetActorIsGuildOwner(ctx, interaction)
		if err != nil {
			message := coinResetErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
			message.Ephemeral = true
			return responder.Reply(ctx, message)
		}
		if !owner {
			message := coinResetErrorMessage("只有服主可以使用這個指令!")
			message.Ephemeral = true
			return responder.Reply(ctx, message)
		}
		divisor := int64(0)
		if value, ok := integerOption(interaction, coinResetOptionDivisor); ok {
			divisor = value
		}
		m.confirmations.Put(coinResetConfirmation{
			GuildID:   interaction.Actor.GuildID,
			ChannelID: interaction.ChannelID,
			UserID:    interaction.Actor.UserID,
			Divisor:   divisor,
		})
		if err := responder.Reply(ctx, coinResetWarningMessage()); err != nil {
			return err
		}
		return m.trackCommand(ctx, interaction, CoinResetCommandName)
	}
}

func (m Module) CoinResetConfirmationHandler() events.Handler {
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
		if event.Content != coinResetConfirmContent {
			_, err := m.messages.SendMessage(ctx, event.ChannelID, coinResetWrongConfirmationOutbound())
			return err
		}
		_, err := m.coinReset.Reset(ctx, domain.CoinResetCommand{
			GuildID: confirmation.GuildID,
			Divisor: confirmation.Divisor,
		})
		if err != nil {
			_, sendErr := m.messages.SendMessage(ctx, event.ChannelID, coinResetGuildErrorOutbound(err))
			return sendErr
		}
		_, err = m.messages.SendMessage(ctx, event.ChannelID, coinResetGuildSuccessOutbound())
		return err
	}
}

func (m Module) coinResetActorIsGuildOwner(ctx context.Context, interaction interactions.Interaction) (bool, error) {
	if m.discord == nil {
		return false, fmt.Errorf("economy coin reset requires discord guild info provider")
	}
	info, err := m.discord.GuildInfo(ctx, interaction.Actor.GuildID)
	if err != nil {
		return false, err
	}
	ownerID := strings.TrimSpace(info.OwnerID)
	if ownerID == "" {
		return false, fmt.Errorf("guild owner id is empty")
	}
	return strings.TrimSpace(interaction.Actor.UserID) == ownerID, nil
}

func coinResetWarningMessage() responses.Message {
	return responses.Message{
		Content:         coinResetWarningContent,
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func coinResetErrorMessage(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: coinResetErrorPrefix + content,
			Color: coinResetErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func coinResetWrongConfirmationOutbound() ports.OutboundMessage {
	return ports.OutboundMessage{
		Embeds: []ports.OutboundEmbed{{
			Title: coinResetErrorPrefix + "你輸入了錯誤的確認!因此視為取消還原",
			Color: coinResetErrorColor,
		}},
		AllowedMentions: ports.AllowedMentions{},
	}
}

func coinResetGuildErrorOutbound(err error) ports.OutboundMessage {
	message := "很抱歉，出現了未知的錯誤，請重試!"
	if errors.Is(err, ports.ErrCoinBalanceNotFound) {
		message = "這伺服器沒有任何的代幣喔!"
	}
	return ports.OutboundMessage{
		Embeds: []ports.OutboundEmbed{{
			Title: coinResetErrorPrefix + message,
			Color: coinResetErrorColor,
		}},
		AllowedMentions: ports.AllowedMentions{},
	}
}

func coinResetGuildSuccessOutbound() ports.OutboundMessage {
	return ports.OutboundMessage{
		Embeds: []ports.OutboundEmbed{{
			Title: coinResetDeleteEmoji + "成功重製伺服器內所有代幣",
			Color: coinResetSuccessColor,
		}},
		AllowedMentions: ports.AllowedMentions{},
	}
}

type coinResetConfirmation struct {
	GuildID   string
	ChannelID string
	UserID    string
	Divisor   int64
	ExpiresAt time.Time
}

type coinResetConfirmationKey struct {
	GuildID   string
	ChannelID string
	UserID    string
}

type coinResetConfirmationStore struct {
	mu      sync.Mutex
	pending map[coinResetConfirmationKey]coinResetConfirmation
	clock   ports.Clock
	ttl     time.Duration
}

func newCoinResetConfirmationStore(clock ports.Clock, ttl time.Duration) *coinResetConfirmationStore {
	if clock == nil {
		clock = ports.SystemClock{}
	}
	if ttl <= 0 {
		ttl = time.Minute
	}
	return &coinResetConfirmationStore{
		pending: map[coinResetConfirmationKey]coinResetConfirmation{},
		clock:   clock,
		ttl:     ttl,
	}
}

func (s *coinResetConfirmationStore) Put(confirmation coinResetConfirmation) {
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
	key := coinResetConfirmationKey{GuildID: confirmation.GuildID, ChannelID: confirmation.ChannelID, UserID: confirmation.UserID}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pending[key] = confirmation
}

func (s *coinResetConfirmationStore) Take(event events.Event) (coinResetConfirmation, bool) {
	if s == nil {
		return coinResetConfirmation{}, false
	}
	key := coinResetConfirmationKey{
		GuildID:   strings.TrimSpace(event.GuildID),
		ChannelID: strings.TrimSpace(event.ChannelID),
		UserID:    strings.TrimSpace(event.UserID),
	}
	if key.GuildID == "" || key.ChannelID == "" || key.UserID == "" {
		return coinResetConfirmation{}, false
	}
	now := event.CreatedAt
	if now.IsZero() {
		now = s.clock.Now()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	confirmation, ok := s.pending[key]
	if !ok {
		return coinResetConfirmation{}, false
	}
	delete(s.pending, key)
	if !now.Before(confirmation.ExpiresAt) {
		return coinResetConfirmation{}, false
	}
	return confirmation, true
}
