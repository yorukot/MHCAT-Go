package domain

import "testing"

func TestJoinRoleConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  JoinRoleConfig
		wantErr bool
	}{
		{name: "valid default give to", config: JoinRoleConfig{GuildID: "guild", RoleID: "role"}},
		{name: "valid bot give to", config: JoinRoleConfig{GuildID: "guild", RoleID: "role", GiveTo: JoinRoleGiveBots}},
		{name: "missing guild", config: JoinRoleConfig{RoleID: "role"}, wantErr: true},
		{name: "missing role", config: JoinRoleConfig{GuildID: "guild"}, wantErr: true},
		{name: "invalid give to", config: JoinRoleConfig{GuildID: "guild", RoleID: "role", GiveTo: "everyone"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("validate: %v", err)
			}
		})
	}
}

func TestNormalizeJoinRoleGiveTo(t *testing.T) {
	if got := NormalizeJoinRoleGiveTo(""); got != JoinRoleGiveAllUsers {
		t.Fatalf("empty = %q", got)
	}
	if got := NormalizeJoinRoleGiveTo(JoinRoleGiveBots); got != JoinRoleGiveBots {
		t.Fatalf("bot = %q", got)
	}
	if got := NormalizeJoinRoleGiveTo("unknown"); got != JoinRoleGiveAllUsers {
		t.Fatalf("unknown = %q", got)
	}
}

func TestLeaveMessageConfigValidation(t *testing.T) {
	channel := LeaveMessageConfig{GuildID: "guild", ChannelID: "channel"}
	if err := channel.ValidateChannel(); err != nil {
		t.Fatalf("channel validate: %v", err)
	}
	content := LeaveMessageConfig{GuildID: "guild", MessageContent: "bye", Title: "bye title", Color: "#df1f2f"}
	if err := content.ValidateContent(); err != nil {
		t.Fatalf("content validate: %v", err)
	}
	random := LeaveMessageConfig{GuildID: "guild", MessageContent: "bye", Title: "bye title", Color: "Random"}
	if err := random.ValidateContent(); err != nil {
		t.Fatalf("random validate: %v", err)
	}
	invalid := LeaveMessageConfig{GuildID: "guild", MessageContent: "bye", Title: "bye title", Color: "not-a-color"}
	if err := invalid.ValidateContent(); err == nil {
		t.Fatal("expected invalid color error")
	}
}
