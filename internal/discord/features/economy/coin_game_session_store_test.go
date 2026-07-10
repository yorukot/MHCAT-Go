package economy

import (
	"testing"
	"time"
)

func TestCoinGameSessionStoreDoesNotRebindNewInviteFromExpiredMessage(t *testing.T) {
	clock := &coinGameTestClock{now: time.Unix(100, 0)}
	store := newCoinGameSessionStore(clock)
	store.Put(coinGameSession{
		ID:             "active",
		GuildID:        "guild-1",
		ChannelID:      "channel-1",
		MessageID:      "old-message",
		ChallengerID:   "user-1",
		OpponentID:     "user-2",
		State:          coinGameSessionActive,
		TurnDeadline:   time.Unix(100, 0),
		TurnGeneration: 1,
	})
	store.Put(coinGameSession{
		ID:           "pending",
		GuildID:      "guild-1",
		ChannelID:    "channel-1",
		ChallengerID: "user-1",
		OpponentID:   "user-2",
		State:        coinGameSessionPending,
	})

	if _, ok := store.ClaimForComponent("guild-1", "user-1", "channel-1", "old-message"); ok {
		t.Fatal("expired bound message claimed a different invite")
	}
	store.mu.Lock()
	pendingMessageID := store.sessions["pending"].session.MessageID
	store.mu.Unlock()
	if pendingMessageID != "" {
		t.Fatalf("pending invite rebound to %q", pendingMessageID)
	}
}
