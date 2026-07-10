package fakemongo

import (
	"context"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type AutoChatPaidRepository struct {
	Requests      []domain.AutoChatPaidRequest
	QueueHook     func(domain.AutoChatPaidRequest)
	QueueResult   domain.AutoChatPaidDispatch
	QueueErr      error
	Response      domain.AutoChatPaidResponse
	ResponseErr   error
	ResponseGuild string
	ResponseTime  int64
}

func (r *AutoChatPaidRepository) QueuePaidAutoChat(ctx context.Context, request domain.AutoChatPaidRequest) (domain.AutoChatPaidDispatch, error) {
	if err := ctx.Err(); err != nil {
		return domain.AutoChatPaidDispatch{}, err
	}
	r.Requests = append(r.Requests, request)
	if r.QueueHook != nil {
		r.QueueHook(request)
	}
	return r.QueueResult, r.QueueErr
}

func (r *AutoChatPaidRepository) GetPaidAutoChatResponse(ctx context.Context, guildID string, requestTimeMilli int64) (domain.AutoChatPaidResponse, error) {
	if err := ctx.Err(); err != nil {
		return domain.AutoChatPaidResponse{}, err
	}
	r.ResponseGuild = guildID
	r.ResponseTime = requestTimeMilli
	return r.Response, r.ResponseErr
}

var _ ports.AutoChatPaidRepository = (*AutoChatPaidRepository)(nil)
