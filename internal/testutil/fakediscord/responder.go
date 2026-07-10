package fakediscord

import (
	"context"
	"fmt"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

type FollowUpEdit struct {
	MessageID string
	Message   responses.Message
}

type Responder struct {
	State           *responses.State
	Replies         []responses.Message
	Defers          []responses.DeferOptions
	DeferredUpdates int
	Modals          []responses.Modal
	Updates         []responses.Message
	Edits           []responses.Message
	Follow          []responses.Message
	FollowIDs       []string
	FollowEdits     []FollowUpEdit
	Errors          []responses.Message
}

func NewResponder() *Responder {
	return &Responder{State: responses.NewState()}
}

func (r *Responder) Reply(ctx context.Context, msg responses.Message) error {
	if err := r.State.MarkReply(ctx, msg); err != nil {
		return err
	}
	r.Replies = append(r.Replies, msg)
	return nil
}

func (r *Responder) Defer(ctx context.Context, opts responses.DeferOptions) error {
	if err := r.State.MarkDefer(ctx, opts); err != nil {
		return err
	}
	r.Defers = append(r.Defers, opts)
	return nil
}

func (r *Responder) DeferUpdate(ctx context.Context) error {
	if err := r.State.MarkDeferUpdate(ctx); err != nil {
		return err
	}
	r.DeferredUpdates++
	return nil
}

func (r *Responder) ShowModal(ctx context.Context, modal responses.Modal) error {
	if err := r.State.MarkModal(ctx, modal); err != nil {
		return err
	}
	r.Modals = append(r.Modals, modal)
	return nil
}

func (r *Responder) UpdateMessage(ctx context.Context, msg responses.Message) error {
	if err := r.State.MarkUpdateMessage(ctx, msg); err != nil {
		return err
	}
	r.Updates = append(r.Updates, msg)
	return nil
}

func (r *Responder) EditOriginal(ctx context.Context, msg responses.Message) error {
	if err := r.State.MarkEditOriginal(ctx, msg); err != nil {
		return err
	}
	r.Edits = append(r.Edits, msg)
	return nil
}

func (r *Responder) FollowUp(ctx context.Context, msg responses.Message) error {
	_, err := r.CreateFollowUp(ctx, msg)
	return err
}

func (r *Responder) CreateFollowUp(ctx context.Context, msg responses.Message) (string, error) {
	if err := r.State.MarkFollowUp(ctx, msg); err != nil {
		return "", err
	}
	messageID := fmt.Sprintf("follow-up-%d", len(r.Follow)+1)
	r.Follow = append(r.Follow, msg)
	r.FollowIDs = append(r.FollowIDs, messageID)
	return messageID, nil
}

func (r *Responder) EditFollowUp(ctx context.Context, messageID string, msg responses.Message) error {
	if err := r.State.MarkEditFollowUp(ctx, messageID); err != nil {
		return err
	}
	r.FollowEdits = append(r.FollowEdits, FollowUpEdit{MessageID: messageID, Message: msg})
	return nil
}

func (r *Responder) Error(ctx context.Context, err error) error {
	msg := responses.SafeErrorMessage(err)
	if r.State.Status() == responses.StatusInitial {
		if replyErr := r.Reply(ctx, msg); replyErr != nil {
			return replyErr
		}
	} else if followErr := r.FollowUp(ctx, msg); followErr != nil {
		return followErr
	}
	r.Errors = append(r.Errors, msg)
	return nil
}
