package events

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type Type string

const (
	TypeReady          Type = "ready"
	TypeResumed        Type = "resumed"
	TypeMessageCreate  Type = "message_create"
	TypeMessageUpdate  Type = "message_update"
	TypeMessageDelete  Type = "message_delete"
	TypeChannelUpdate  Type = "channel_update"
	TypeReactionAdd    Type = "reaction_add"
	TypeReactionRemove Type = "reaction_remove"
	TypeMemberAdd      Type = "member_add"
	TypeMemberRemove   Type = "member_remove"
	TypeVoiceState     Type = "voice_state"
)

var ErrNoHandler = errors.New("discord event handler not found")
var ErrStopPropagation = errors.New("discord event propagation stopped")

type Event struct {
	Type          Type
	ID            string
	GuildID       string
	GuildName     string
	GuildIconURL  string
	BotUserID     string
	BotAvatarURL  string
	ChannelID     string
	MessageID     string
	UserID        string
	Username      string
	UserTag       string
	AvatarURL     string
	Content       string
	OldContent    string
	HasOldContent bool
	IsBot         bool
	CreatedAt     time.Time
	Attachments   []Attachment

	Reaction      *Reaction
	Member        *Member
	ChannelUpdate *ChannelUpdate
	VoiceState    *VoiceState
}

type Attachment struct {
	URL string
}

type Reaction struct {
	EmojiName string
	EmojiID   string
}

type Member struct {
	UserID           string
	Username         string
	UserTag          string
	RoleIDs          []string
	JoinedAt         time.Time
	AccountCreatedAt time.Time
	IsBot            bool
	AvatarURL        string
}

type ChannelUpdate struct {
	ChannelID               string
	GuildID                 string
	OldTopic                string
	NewTopic                string
	HasOldChannel           bool
	OldPermissionOverwrites []PermissionOverwrite
	NewPermissionOverwrites []PermissionOverwrite
}

type PermissionOverwrite struct {
	ID    string
	Type  int
	Allow int64
	Deny  int64
}

type VoiceState struct {
	UserID        string
	GuildID       string
	ChannelID     string
	BeforeChannel string
}

type Handler func(ctx context.Context, event Event) error
type ShutdownFunc func(ctx context.Context) error

type Dispatcher struct {
	mu           sync.RWMutex
	handlers     map[Type][]Handler
	shutdowns    []ShutdownFunc
	shutdownOnce sync.Once
	shutdownErr  error
	logger       *slog.Logger
}

func NewDispatcher(logger *slog.Logger) *Dispatcher {
	if logger == nil {
		logger = slog.Default()
	}
	return &Dispatcher{handlers: map[Type][]Handler{}, logger: logger}
}

func (d *Dispatcher) Register(eventType Type, handler Handler) {
	if d == nil || handler == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.handlers == nil {
		d.handlers = map[Type][]Handler{}
	}
	d.handlers[eventType] = append(d.handlers[eventType], handler)
}

func (d *Dispatcher) RegisterShutdown(fn ShutdownFunc) {
	if d == nil || fn == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.shutdowns = append(d.shutdowns, fn)
}

func (d *Dispatcher) HasHandlers(eventType Type) bool {
	if d == nil {
		return false
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.handlers[eventType]) > 0
}

func (d *Dispatcher) Dispatch(ctx context.Context, event Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if d == nil {
		return fmt.Errorf("%w: dispatcher is nil", ErrNoHandler)
	}
	d.mu.RLock()
	handlers := append([]Handler(nil), d.handlers[event.Type]...)
	d.mu.RUnlock()
	if len(handlers) == 0 {
		return fmt.Errorf("%w: %s", ErrNoHandler, event.Type)
	}
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			if errors.Is(err, ErrStopPropagation) {
				return ctx.Err()
			}
			return err
		}
	}
	return ctx.Err()
}

func (d *Dispatcher) DispatchSafe(ctx context.Context, event Event) {
	if err := d.Dispatch(ctx, event); err != nil && !errors.Is(err, ErrNoHandler) {
		d.logger.WarnContext(ctx, "discord event handler failed", "type", event.Type, "error", err.Error())
	}
}

func (d *Dispatcher) Shutdown(ctx context.Context) error {
	if d == nil {
		return nil
	}
	d.shutdownOnce.Do(func() {
		d.mu.RLock()
		shutdowns := append([]ShutdownFunc(nil), d.shutdowns...)
		d.mu.RUnlock()
		var errs []error
		for i := len(shutdowns) - 1; i >= 0; i-- {
			if shutdowns[i] == nil {
				continue
			}
			if err := shutdowns[i](ctx); err != nil {
				errs = append(errs, err)
			}
		}
		d.shutdownErr = errors.Join(errs...)
	})
	return d.shutdownErr
}
