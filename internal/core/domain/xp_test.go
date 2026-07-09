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
