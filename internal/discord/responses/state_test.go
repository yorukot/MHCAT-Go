package responses_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
)

func TestResponderReplyOnce(t *testing.T) {
	responder := fakediscord.NewResponder()
	if err := responder.Reply(context.Background(), responses.Message{Content: "ok"}); err != nil {
		t.Fatalf("reply: %v", err)
	}
	if len(responder.Replies) != 1 {
		t.Fatalf("expected one reply, got %d", len(responder.Replies))
	}
}

func TestResponderDoubleReplyFails(t *testing.T) {
	responder := fakediscord.NewResponder()
	if err := responder.Reply(context.Background(), responses.Message{Content: "ok"}); err != nil {
		t.Fatalf("reply: %v", err)
	}
	if err := responder.Reply(context.Background(), responses.Message{Content: "again"}); !errors.Is(err, responses.ErrAlreadyResponded) {
		t.Fatalf("expected ErrAlreadyResponded, got %v", err)
	}
}

func TestResponderDeferThenEditOriginal(t *testing.T) {
	responder := fakediscord.NewResponder()
	deadline := time.Now().Add(time.Minute)
	if err := responder.Defer(context.Background(), responses.DeferOptions{Ephemeral: true, Deadline: deadline}); err != nil {
		t.Fatalf("defer: %v", err)
	}
	if err := responder.EditOriginal(context.Background(), responses.Message{Content: "done"}); err != nil {
		t.Fatalf("edit: %v", err)
	}
	if got := responder.State.Status(); got != responses.StatusDeferred {
		t.Fatalf("status = %s", got)
	}
	if !responder.State.Ephemeral() {
		t.Fatal("expected ephemeral state to be preserved")
	}
	if responder.State.Deadline().IsZero() {
		t.Fatal("expected defer deadline to be recorded")
	}
}

func TestResponderDeferThenFollowUp(t *testing.T) {
	responder := fakediscord.NewResponder()
	if err := responder.Defer(context.Background(), responses.DeferOptions{}); err != nil {
		t.Fatalf("defer: %v", err)
	}
	if err := responder.FollowUp(context.Background(), responses.Message{Content: "later"}); err != nil {
		t.Fatalf("follow up: %v", err)
	}
}

func TestResponderUpdateMessageThenFollowUp(t *testing.T) {
	responder := fakediscord.NewResponder()
	if err := responder.UpdateMessage(context.Background(), responses.Message{Content: "updated"}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := responder.FollowUp(context.Background(), responses.Message{Content: "done", Ephemeral: true}); err != nil {
		t.Fatalf("follow up: %v", err)
	}
	if len(responder.Updates) != 1 || len(responder.Follow) != 1 {
		t.Fatalf("updates=%#v follow=%#v", responder.Updates, responder.Follow)
	}
}

func TestResponderDeferUpdateThenEditOriginal(t *testing.T) {
	responder := fakediscord.NewResponder()
	if err := responder.DeferUpdate(context.Background()); err != nil {
		t.Fatalf("defer update: %v", err)
	}
	if err := responder.EditOriginal(context.Background(), responses.Message{Content: "updated"}); err != nil {
		t.Fatalf("edit original: %v", err)
	}
	if responder.DeferredUpdates != 1 || len(responder.Edits) != 1 {
		t.Fatalf("deferred updates=%d edits=%#v", responder.DeferredUpdates, responder.Edits)
	}
	if got := responder.State.Status(); got != responses.StatusDeferred {
		t.Fatalf("status = %s", got)
	}
}

func TestResponderShowModalOnce(t *testing.T) {
	responder := fakediscord.NewResponder()
	modal := responses.Modal{
		CustomID: "mhcat:v1:ticket:setup:c=111,r=222",
		Title:    "私人頻道系統!",
		Rows: []responses.ModalRow{{Inputs: []responses.TextInput{{
			CustomID: "ticketcolor",
			Label:    "請輸入嵌入顏色",
			Style:    responses.TextInputStyleShort,
			Required: true,
		}}}},
	}
	if err := responder.ShowModal(context.Background(), modal); err != nil {
		t.Fatalf("show modal: %v", err)
	}
	if len(responder.Modals) != 1 || responder.State.Status() != responses.StatusReplied {
		t.Fatalf("modals=%#v status=%s", responder.Modals, responder.State.Status())
	}
	if err := responder.Reply(context.Background(), responses.Message{Content: "again"}); !errors.Is(err, responses.ErrAlreadyResponded) {
		t.Fatalf("expected ErrAlreadyResponded, got %v", err)
	}
}

func TestResponderInvalidModalFails(t *testing.T) {
	responder := fakediscord.NewResponder()
	if err := responder.ShowModal(context.Background(), responses.Modal{Title: "missing id"}); !errors.Is(err, responses.ErrInvalidModal) {
		t.Fatalf("expected ErrInvalidModal, got %v", err)
	}
}

func TestResponderUpdateAfterInitialResponseFails(t *testing.T) {
	responder := fakediscord.NewResponder()
	if err := responder.Reply(context.Background(), responses.Message{Content: "ok"}); err != nil {
		t.Fatalf("reply: %v", err)
	}
	if err := responder.UpdateMessage(context.Background(), responses.Message{Content: "again"}); !errors.Is(err, responses.ErrAlreadyResponded) {
		t.Fatalf("expected ErrAlreadyResponded, got %v", err)
	}
}

func TestResponderFollowUpBeforeInitialResponseFails(t *testing.T) {
	responder := fakediscord.NewResponder()
	if err := responder.FollowUp(context.Background(), responses.Message{Content: "too soon"}); !errors.Is(err, responses.ErrNoInitialResponse) {
		t.Fatalf("expected ErrNoInitialResponse, got %v", err)
	}
}

func TestResponderEditBeforeInitialResponseFails(t *testing.T) {
	responder := fakediscord.NewResponder()
	if err := responder.EditOriginal(context.Background(), responses.Message{Content: "too soon"}); !errors.Is(err, responses.ErrNoInitialResponse) {
		t.Fatalf("expected ErrNoInitialResponse, got %v", err)
	}
}

func TestSafeErrorResponseRedactsInternalDetails(t *testing.T) {
	responder := fakediscord.NewResponder()
	internal := errors.New("database password=secret failed")
	if err := responder.Error(context.Background(), internal); err != nil {
		t.Fatalf("error response: %v", err)
	}
	if len(responder.Errors) != 1 {
		t.Fatalf("expected one error response, got %d", len(responder.Errors))
	}
	if strings.Contains(responder.Errors[0].Content, "secret") || strings.Contains(responder.Errors[0].Content, "database") {
		t.Fatalf("error response leaked internal detail: %q", responder.Errors[0].Content)
	}
	if !responder.Errors[0].Ephemeral {
		t.Fatal("safe error response should be ephemeral")
	}
}

func TestResponderEphemeralStatePreserved(t *testing.T) {
	responder := fakediscord.NewResponder()
	if err := responder.Defer(context.Background(), responses.DeferOptions{Ephemeral: true}); err != nil {
		t.Fatalf("defer: %v", err)
	}
	if err := responder.EditOriginal(context.Background(), responses.Message{Content: "done"}); err != nil {
		t.Fatalf("edit: %v", err)
	}
	if !responder.State.Ephemeral() {
		t.Fatal("expected ephemeral state to remain true")
	}
}
