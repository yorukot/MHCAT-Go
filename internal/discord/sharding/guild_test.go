package sharding

import (
	"fmt"
	"testing"
)

func TestOwnsGuildUsesDiscordSnowflakeFormula(t *testing.T) {
	for shardID := 0; shardID < 16; shardID++ {
		guildID := uint64(shardID) << 22
		if !OwnsGuild(formatSnowflake(guildID), shardID, 16) {
			t.Fatalf("shard %d did not own guild %d", shardID, guildID)
		}
		if OwnsGuild(formatSnowflake(guildID), (shardID+1)%16, 16) {
			t.Fatalf("wrong shard owned guild %d", guildID)
		}
	}
}

func TestOwnsGuildSingleShardAcceptsLegacyTestIDs(t *testing.T) {
	if !OwnsGuild("guild-1", 0, 1) {
		t.Fatal("single shard should own every guild")
	}
}

func TestOwnsGuildRejectsInvalidInput(t *testing.T) {
	for _, tc := range []struct {
		guildID    string
		shardID    int
		shardCount int
	}{
		{guildID: "invalid", shardID: 0, shardCount: 16},
		{guildID: "1", shardID: -1, shardCount: 16},
		{guildID: "1", shardID: 16, shardCount: 16},
		{guildID: "1", shardID: 1, shardCount: 1},
	} {
		if OwnsGuild(tc.guildID, tc.shardID, tc.shardCount) {
			t.Fatalf("unexpected ownership for %#v", tc)
		}
	}
}

func formatSnowflake(value uint64) string {
	return fmt.Sprintf("%d", value)
}
