package autochat

import (
	"context"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

const LegacyAutoChatPricePerUnit = 0.00003

type PaidHandoffService struct {
	configs  ports.AutoChatConfigReader
	balances ports.BalanceRepository
	handoff  ports.AutoChatPaidRepository
}

func NewPaidHandoffService(configs ports.AutoChatConfigReader, balances ports.BalanceRepository, handoff ports.AutoChatPaidRepository) (PaidHandoffService, error) {
	if configs == nil || balances == nil || handoff == nil {
		return PaidHandoffService{}, errors.New("paid autochat repositories are required")
	}
	return PaidHandoffService{configs: configs, balances: balances, handoff: handoff}, nil
}

func (s PaidHandoffService) Submit(ctx context.Context, guildID string, channelID string, content string, now time.Time) (domain.AutoChatPaidSubmission, error) {
	if err := ctx.Err(); err != nil {
		return domain.AutoChatPaidSubmission{}, err
	}
	guildID = strings.TrimSpace(guildID)
	channelID = strings.TrimSpace(channelID)
	if guildID == "" || channelID == "" || now.IsZero() {
		return domain.AutoChatPaidSubmission{State: domain.AutoChatPaidIgnored}, nil
	}
	config, err := s.configs.GetAutoChatConfig(ctx, guildID)
	if errors.Is(err, ports.ErrAutoChatConfigMissing) {
		return domain.AutoChatPaidSubmission{State: domain.AutoChatPaidIgnored}, nil
	}
	if err != nil {
		return domain.AutoChatPaidSubmission{}, err
	}
	if config.ChannelID != channelID {
		return domain.AutoChatPaidSubmission{State: domain.AutoChatPaidIgnored}, nil
	}
	balance, err := s.balances.GetBalance(ctx, guildID)
	if errors.Is(err, ports.ErrBalanceMissing) {
		return domain.AutoChatPaidSubmission{State: domain.AutoChatPaidIgnored}, nil
	}
	if err != nil {
		return domain.AutoChatPaidSubmission{}, err
	}
	if !legacyPositiveAutoChatBalance(balance.Amount) {
		return domain.AutoChatPaidSubmission{State: domain.AutoChatPaidIgnored}, nil
	}
	if strings.Contains(content, "@") {
		return domain.AutoChatPaidSubmission{State: domain.AutoChatPaidUnsafeInput}, nil
	}

	request := domain.AutoChatPaidRequest{
		GuildID:          guildID,
		Content:          content,
		Cost:             LegacyAutoChatMessageCost(content),
		RequestedAtMilli: now.UnixMilli(),
	}
	dispatch, err := s.handoff.QueuePaidAutoChat(ctx, request)
	if errors.Is(err, ports.ErrAutoChatPaidBalanceUnavailable) {
		return domain.AutoChatPaidSubmission{State: domain.AutoChatPaidIgnored}, nil
	}
	if errors.Is(err, ports.ErrAutoChatPaidBusy) {
		return domain.AutoChatPaidSubmission{State: domain.AutoChatPaidBusy}, nil
	}
	if err != nil {
		return domain.AutoChatPaidSubmission{}, err
	}
	return domain.AutoChatPaidSubmission{State: domain.AutoChatPaidQueued, Dispatch: dispatch}, ctx.Err()
}

func (s PaidHandoffService) Response(ctx context.Context, guildID string, requestTimeMilli int64) (domain.AutoChatPaidResponse, error) {
	if err := ctx.Err(); err != nil {
		return domain.AutoChatPaidResponse{}, err
	}
	return s.handoff.GetPaidAutoChatResponse(ctx, strings.TrimSpace(guildID), requestTimeMilli)
}

func LegacyAutoChatMessageCost(content string) float64 {
	units := utf16.Encode([]rune(content))
	length := 0
	for _, unit := range units {
		if unit <= 0xff {
			length++
		} else {
			length += 2
		}
	}
	return float64(length) * LegacyAutoChatPricePerUnit
}

func legacyPositiveAutoChatBalance(value string) bool {
	amount, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	return err == nil && amount > 0 && !math.IsNaN(amount) && !math.IsInf(amount, 0)
}
