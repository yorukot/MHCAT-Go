package responses

import (
	"context"
	"sync"
	"time"
)

type Status string

const (
	StatusInitial  Status = "initial"
	StatusReplied  Status = "replied"
	StatusDeferred Status = "deferred"
)

type State struct {
	mu        sync.Mutex
	status    Status
	ephemeral bool
	deadline  time.Time
}

func NewState() *State {
	return &State{status: StatusInitial}
}

func (s *State) Status() Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status
}

func (s *State) Ephemeral() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ephemeral
}

func (s *State) Deadline() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.deadline
}

func (s *State) MarkReply(ctx context.Context, msg Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.status != StatusInitial {
		return ErrAlreadyResponded
	}
	s.status = StatusReplied
	s.ephemeral = msg.Ephemeral
	return nil
}

func (s *State) MarkDefer(ctx context.Context, opts DeferOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.status != StatusInitial {
		return ErrAlreadyResponded
	}
	s.status = StatusDeferred
	s.ephemeral = opts.Ephemeral
	s.deadline = opts.Deadline
	return nil
}

func (s *State) MarkDeferUpdate(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.status != StatusInitial {
		return ErrAlreadyResponded
	}
	s.status = StatusDeferred
	return nil
}

func (s *State) MarkModal(ctx context.Context, modal Modal) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.status != StatusInitial {
		return ErrAlreadyResponded
	}
	if modal.CustomID == "" || modal.Title == "" {
		return ErrInvalidModal
	}
	s.status = StatusReplied
	return nil
}

func (s *State) MarkEditOriginal(ctx context.Context, msg Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.status == StatusInitial {
		return ErrNoInitialResponse
	}
	if s.status == StatusDeferred && msg.Ephemeral && !s.ephemeral {
		return ErrEphemeralChanged
	}
	return nil
}

func (s *State) MarkUpdateMessage(ctx context.Context, msg Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.status != StatusInitial {
		return ErrAlreadyResponded
	}
	s.status = StatusReplied
	s.ephemeral = msg.Ephemeral
	return nil
}

func (s *State) MarkFollowUp(ctx context.Context, msg Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.status == StatusInitial {
		return ErrNoInitialResponse
	}
	return nil
}
