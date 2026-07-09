package discordgo

import (
	"context"
	"testing"

	dgo "github.com/bwmarrin/discordgo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

func TestRegisterInteractionHandlerNilSafe(t *testing.T) {
	var session *Session
	remove := session.RegisterInteractionHandler(nil)
	remove()
}

func TestReadyNilSessionReturnsClosedChannel(t *testing.T) {
	var session *Session
	select {
	case <-session.Ready():
	default:
		t.Fatal("nil session ready channel should be closed")
	}
}

func TestRegisterInteractionHandlerDoesNotOpenGateway(t *testing.T) {
	session := &Session{session: &dgo.Session{}, ready: make(chan struct{})}
	called := false
	remove := session.RegisterInteractionHandler(func(context.Context, interactions.Interaction, responses.Responder) error {
		called = true
		return nil
	})
	remove()
	if called {
		t.Fatal("handler should not be called during registration")
	}
	if session.Opened() {
		t.Fatal("gateway opened during handler registration")
	}
}
