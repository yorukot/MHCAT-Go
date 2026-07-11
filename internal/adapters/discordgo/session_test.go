package discordgo

import (
	"errors"
	"net"
	"net/http"
	"testing"

	dgo "github.com/bwmarrin/discordgo"
)

func TestNewSessionConfiguresShardIdentify(t *testing.T) {
	session, err := NewSession("token", dgo.IntentsGuilds, 7, 16)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	if session.ShardID() != 7 || session.ShardCount() != 16 {
		t.Fatalf("shard = %d/%d", session.ShardID(), session.ShardCount())
	}
	if session.session.Identify.Intents != dgo.IntentsGuilds {
		t.Fatalf("intents = %v", session.session.Identify.Intents)
	}
}

func TestSessionOpenCloseLifecycleAndGatewayDialFailure(t *testing.T) {
	session, err := NewSession("token", dgo.IntentsGuilds, 0, 1)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	session.opened = true
	if err := session.Open(); err != nil {
		t.Fatalf("idempotent open: %v", err)
	}
	if err := session.Close(); err != nil || session.Opened() {
		t.Fatalf("close unopened gateway: opened=%v err=%v", session.Opened(), err)
	}
	if err := session.Close(); err != nil {
		t.Fatalf("idempotent close: %v", err)
	}

	dialErr := errors.New("gateway dial blocked")
	session.session.Client = &http.Client{Transport: responderRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodGet || request.URL.Path != "/api/v9/gateway" {
			t.Fatalf("unexpected gateway request: %s %s", request.Method, request.URL.String())
		}
		return responderHTTPResponse(request, http.StatusOK, `{"url":"ws://gateway.invalid"}`), nil
	})}
	session.session.Dialer.NetDial = func(string, string) (net.Conn, error) { return nil, dialErr }
	if err := session.Open(); !errors.Is(err, dialErr) {
		t.Fatalf("gateway open error = %v", err)
	}
	if session.Opened() {
		t.Fatal("failed gateway open changed lifecycle state")
	}
}
