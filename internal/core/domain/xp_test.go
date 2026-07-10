package domain

import "testing"

func TestTextXPConfigValidate(t *testing.T) {
	valid := TextXPConfig{GuildID: "guild-1", ChannelID: "channel-1", Color: "#00ff00"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}
	for _, invalid := range []TextXPConfig{
		{ChannelID: "channel-1"},
		{GuildID: "guild-1"},
		{GuildID: "guild-1", ChannelID: "channel-1", Color: "not-a-color"},
	} {
		if err := invalid.Validate(); err == nil {
			t.Fatalf("invalid config accepted: %#v", invalid)
		}
	}
}

func TestVoiceXPConfigValidate(t *testing.T) {
	valid := VoiceXPConfig{GuildID: "guild-1", ChannelID: "channel-1", Color: "#00ff00"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}
	for _, invalid := range []VoiceXPConfig{
		{ChannelID: "channel-1"},
		{GuildID: "guild-1"},
		{GuildID: "guild-1", ChannelID: "channel-1", Color: "not-a-color"},
	} {
		if err := invalid.Validate(); err == nil {
			t.Fatalf("invalid config accepted: %#v", invalid)
		}
	}
}

func TestValidLegacyColor(t *testing.T) {
	for _, value := range []string{"#fff", "#ffffff", "ffffff", "red", "Green", ""} {
		if !ValidLegacyColor(value) {
			t.Fatalf("expected %q to be valid", value)
		}
	}
	for _, value := range []string{"#ffff", "bad color", "url(javascript:alert(1))"} {
		if ValidLegacyColor(value) {
			t.Fatalf("expected %q to be invalid", value)
		}
	}
}

func TestParseLegacyColorValue(t *testing.T) {
	tests := map[string]int{
		"#df1f2f": 0xDF1F2F,
		"df1":     0xDDFF11,
		"red":     0xFF0000,
	}
	for value, want := range tests {
		got, ok := ParseLegacyColorValue(value)
		if !ok || got != want {
			t.Fatalf("%q parsed to %x/%t, want %x", value, got, ok, want)
		}
	}
	if _, ok := ParseLegacyColorValue("not-a-color"); ok {
		t.Fatal("unexpected valid color")
	}
}

func TestApplyTextXPAdjustmentMatchesLegacyLevelMath(t *testing.T) {
	profile := XPProfile{GuildID: "guild-1", UserID: "user-1", XP: 20, Level: 1}

	got := ApplyTextXPAdjustment(profile, 200)
	if got.Level != 2 || got.XP != 87 {
		t.Fatalf("positive text XP adjustment = %#v", got)
	}

	got = ApplyTextXPAdjustment(XPProfile{GuildID: "guild-1", UserID: "user-1", XP: 20, Level: 1}, -30)
	if got.Level != 0 || got.XP != 90 {
		t.Fatalf("negative text XP adjustment = %#v", got)
	}
}

func TestApplyTextXPMessageMatchesLegacyAccrualLevelBehavior(t *testing.T) {
	got, leveled := ApplyTextXPMessage(XPProfile{GuildID: "guild-1", UserID: "user-1", XP: 95, Level: 0}, 5)
	if leveled || got.Level != 0 || got.XP != 100 {
		t.Fatalf("non-level message = %#v leveled=%t", got, leveled)
	}

	got, leveled = ApplyTextXPMessage(XPProfile{GuildID: "guild-1", UserID: "user-1", XP: 96, Level: 0}, 5)
	if !leveled || got.Level != 1 || got.XP != 0 {
		t.Fatalf("level message = %#v leveled=%t", got, leveled)
	}
}

func TestApplyVoiceXPAdjustmentMatchesLegacyLevelMath(t *testing.T) {
	profile := XPProfile{GuildID: "guild-1", UserID: "user-1", XP: 50, Level: 2}

	got := ApplyVoiceXPAdjustment(profile, 500)
	if got.Level != 3 || got.XP != 250 {
		t.Fatalf("positive voice XP adjustment = %#v", got)
	}
}

func TestApplyXPAdjustmentPreservesLegacyZeroDeltaQuirk(t *testing.T) {
	got := ApplyTextXPAdjustment(XPProfile{GuildID: "guild-1", UserID: "user-1", XP: 70, Level: 4}, 0)
	if got.Level != 5 || got.XP != 0 {
		t.Fatalf("zero delta adjustment = %#v", got)
	}
}
