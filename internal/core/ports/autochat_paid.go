package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

var (
	ErrAutoChatPaidBalanceUnavailable = errors.New("paid autochat balance is unavailable")
	ErrAutoChatPaidBusy               = errors.New("paid autochat request is busy")
	ErrAutoChatPaidResponseMissing    = errors.New("paid autochat response is missing")
	ErrAutoChatPaidStateConflict      = errors.New("paid autochat state conflicts with singleton requirements")
)

type AutoChatPaidRepository interface {
	QueuePaidAutoChat(ctx context.Context, request domain.AutoChatPaidRequest) (domain.AutoChatPaidDispatch, error)
	GetPaidAutoChatResponse(ctx context.Context, guildID string, requestTimeMilli int64) (domain.AutoChatPaidResponse, error)
}
