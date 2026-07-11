package runtime

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

func TestTrackingResponderMarksEverySuccessfulResponse(t *testing.T) {
	tests := []struct {
		name string
		call func(*trackingResponder) error
	}{
		{name: "reply", call: func(r *trackingResponder) error { return r.Reply(context.Background(), responses.Message{}) }},
		{name: "defer", call: func(r *trackingResponder) error { return r.Defer(context.Background(), responses.DeferOptions{}) }},
		{name: "defer update", call: func(r *trackingResponder) error { return r.DeferUpdate(context.Background()) }},
		{name: "modal", call: func(r *trackingResponder) error { return r.ShowModal(context.Background(), responses.Modal{}) }},
		{name: "update", call: func(r *trackingResponder) error { return r.UpdateMessage(context.Background(), responses.Message{}) }},
		{name: "edit original", call: func(r *trackingResponder) error { return r.EditOriginal(context.Background(), responses.Message{}) }},
		{name: "follow up", call: func(r *trackingResponder) error { return r.FollowUp(context.Background(), responses.Message{}) }},
		{name: "create follow up", call: func(r *trackingResponder) error {
			messageID, err := r.CreateFollowUp(context.Background(), responses.Message{})
			if err == nil && messageID != "message-1" {
				t.Fatalf("message ID = %q", messageID)
			}
			return err
		}},
		{name: "edit follow up", call: func(r *trackingResponder) error {
			return r.EditFollowUp(context.Background(), "message-1", responses.Message{})
		}},
		{name: "error", call: func(r *trackingResponder) error { return r.Error(context.Background(), errors.New("boom")) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tracked := &trackingResponder{next: &runtimeResponderStub{}}
			if err := test.call(tracked); err != nil {
				t.Fatalf("response: %v", err)
			}
			if !tracked.responded {
				t.Fatal("successful response was not tracked")
			}
		})
	}
}

func TestTrackingResponderDoesNotMarkFailedOrDeleteOperations(t *testing.T) {
	responseErr := errors.New("response failed")
	failed := &trackingResponder{next: &runtimeResponderStub{err: responseErr}}
	if err := failed.Reply(context.Background(), responses.Message{}); !errors.Is(err, responseErr) || failed.responded {
		t.Fatalf("failed reply error=%v responded=%v", err, failed.responded)
	}

	deleted := &trackingResponder{next: &runtimeResponderStub{}}
	if err := deleted.DeleteFollowUp(context.Background(), "message-1"); err != nil || deleted.responded {
		t.Fatalf("delete error=%v responded=%v", err, deleted.responded)
	}
}

func TestDispatcherNilAndHandlerGuards(t *testing.T) {
	if _, err := NewDispatcher(nil, nil); !errors.Is(err, ErrRuntimeNotConfigured) {
		t.Fatalf("nil router error = %v", err)
	}
	var nilDispatcher *Dispatcher
	if err := nilDispatcher.Shutdown(context.Background()); err != nil {
		t.Fatalf("nil shutdown: %v", err)
	}
	nilDispatcher.RegisterShutdown(func(context.Context) error { return errors.New("unexpected") })
	if err := nilDispatcher.Handler()(context.Background(), interactions.Interaction{}, &runtimeResponderStub{}); !errors.Is(err, ErrRuntimeNotConfigured) {
		t.Fatalf("nil dispatcher error = %v", err)
	}
	dispatcher, err := NewDispatcher(interactions.NewRouter(), nil)
	if err != nil {
		t.Fatalf("new dispatcher: %v", err)
	}
	dispatcher.RegisterShutdown(nil)
	if err := dispatcher.Dispatch(context.Background(), interactions.Interaction{}, nil); !errors.Is(err, ErrRuntimeNotConfigured) {
		t.Fatalf("nil responder error = %v", err)
	}
}

type runtimeResponderStub struct {
	err error
}

func (r *runtimeResponderStub) Reply(context.Context, responses.Message) error         { return r.err }
func (r *runtimeResponderStub) Defer(context.Context, responses.DeferOptions) error    { return r.err }
func (r *runtimeResponderStub) DeferUpdate(context.Context) error                      { return r.err }
func (r *runtimeResponderStub) ShowModal(context.Context, responses.Modal) error       { return r.err }
func (r *runtimeResponderStub) UpdateMessage(context.Context, responses.Message) error { return r.err }
func (r *runtimeResponderStub) EditOriginal(context.Context, responses.Message) error  { return r.err }
func (r *runtimeResponderStub) FollowUp(context.Context, responses.Message) error      { return r.err }
func (r *runtimeResponderStub) CreateFollowUp(context.Context, responses.Message) (string, error) {
	return "message-1", r.err
}
func (r *runtimeResponderStub) EditFollowUp(context.Context, string, responses.Message) error {
	return r.err
}
func (r *runtimeResponderStub) DeleteFollowUp(context.Context, string) error { return r.err }
func (r *runtimeResponderStub) Error(context.Context, error) error           { return r.err }
