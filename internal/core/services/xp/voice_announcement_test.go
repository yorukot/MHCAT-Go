package xp

import "testing"

func TestLegacyVoiceXPLevelUpAnnouncementUsesDefaultAndPlaceholders(t *testing.T) {
	if got := LegacyVoiceXPLevelUpAnnouncement("", 4, " user-1 "); got != "🆙恭喜<@user-1> 的語音等級成功升級到 4" {
		t.Fatalf("default announcement = %q", got)
	}
	got := LegacyVoiceXPLevelUpAnnouncement("(user) {user} (leavel) {level}", 5, "user-2")
	if got != "<@user-2> <@user-2> 5 5" {
		t.Fatalf("custom announcement = %q", got)
	}
}
