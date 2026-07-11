package sharding

import (
	"strconv"
	"strings"
)

func OwnsGuild(guildID string, shardID int, shardCount int) bool {
	if shardCount <= 1 {
		return shardID == 0
	}
	if shardID < 0 || shardID >= shardCount {
		return false
	}
	snowflake, err := strconv.ParseUint(strings.TrimSpace(guildID), 10, 64)
	if err != nil {
		return false
	}
	return int((snowflake>>22)%uint64(shardCount)) == shardID
}
