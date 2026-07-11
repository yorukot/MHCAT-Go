package discordgo

import (
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
