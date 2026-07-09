package discordgo

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	dgo "github.com/bwmarrin/discordgo"
)

type Session struct {
	mu        sync.Mutex
	session   *dgo.Session
	opened    bool
	ready     chan struct{}
	readyOnce sync.Once
}

func NewSession(token string, intents dgo.Intent) (*Session, error) {
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("discord token is required")
	}
	session, err := dgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("create discord session: %w", err)
	}
	session.Identify.Intents = intents
	wrapped := &Session{session: session, ready: make(chan struct{})}
	session.AddHandler(func(_ *dgo.Session, _ *dgo.Ready) {
		wrapped.readyOnce.Do(func() { close(wrapped.ready) })
	})
	return wrapped, nil
}

func (s *Session) Open() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.opened {
		return nil
	}
	if err := s.session.Open(); err != nil {
		return fmt.Errorf("open discord gateway: %w", err)
	}
	s.opened = true
	return nil
}

func (s *Session) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.opened {
		return nil
	}
	if err := s.session.Close(); err != nil {
		return fmt.Errorf("close discord session: %w", err)
	}
	s.opened = false
	return nil
}

func (s *Session) Opened() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.opened
}

func (s *Session) Ready() <-chan struct{} {
	if s == nil || s.ready == nil {
		closed := make(chan struct{})
		close(closed)
		return closed
	}
	return s.ready
}
