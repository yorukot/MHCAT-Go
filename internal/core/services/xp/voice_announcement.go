package xp

import (
	"strconv"
	"strings"
)

func LegacyVoiceXPLevelUpAnnouncement(message string, level int64, userID string) string {
	levelText := strconv.FormatInt(level, 10)
	userID = strings.TrimSpace(userID)
	mention := "<@" + userID + ">"
	if message == "" {
		return "🆙恭喜" + mention + " 的語音等級成功升級到 " + levelText
	}
	out := strings.Replace(message, "(leavel)", levelText, 1)
	out = strings.Replace(out, "{level}", levelText, 1)
	out = strings.Replace(out, "(user)", mention, 1)
	out = strings.Replace(out, "{user}", mention, 1)
	return out
}
