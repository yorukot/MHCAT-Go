package repositories

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

func TestXPRepositoryConstructorsRejectNilCollections(t *testing.T) {
	if _, err := NewTextXPConfigRepository(nil); err == nil {
		t.Fatal("expected nil text XP config collection error")
	}
	if _, err := NewVoiceXPConfigRepository(nil); err == nil {
		t.Fatal("expected nil voice XP config collection error")
	}
	if _, err := NewTextXPRewardRoleRepository(nil); err == nil {
		t.Fatal("expected nil text XP reward collection error")
	}
	if _, err := NewVoiceXPRewardRoleRepository(nil); err == nil {
		t.Fatal("expected nil voice XP reward collection error")
	}
	if _, err := NewXPAdminRepository(nil, nil); err == nil {
		t.Fatal("expected nil text XP profile collection error")
	}
}

func TestXPConfigMongoIntegrationLifecycle(t *testing.T) {
	database := xpAccrualIntegrationDatabase(t)
	text, err := NewTextXPConfigRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new text XP config repository: %v", err)
	}
	voice, err := NewVoiceXPConfigRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new voice XP config repository: %v", err)
	}
	ctx := context.Background()

	textConfig := domain.TextXPConfig{GuildID: "guild-1", ChannelID: "text-1", Color: "#112233", Message: "level up"}
	if err := text.SaveTextXPConfig(ctx, textConfig); err != nil {
		t.Fatalf("save text XP config: %v", err)
	}
	textConfig.ChannelID = "text-2"
	textConfig.Color = ""
	textConfig.Message = ""
	if err := text.SaveTextXPConfig(ctx, textConfig); err != nil {
		t.Fatalf("update text XP config: %v", err)
	}
	loadedText, err := text.GetTextXPConfig(ctx, " guild-1 ")
	if err != nil || loadedText.ChannelID != "text-2" || loadedText.Color != "" || loadedText.Message != "" {
		t.Fatalf("text XP config = %#v err=%v", loadedText, err)
	}

	voiceConfig := domain.VoiceXPConfig{GuildID: "guild-1", ChannelID: "voice-1", Color: "Blue", Message: "voice level"}
	if err := voice.SaveVoiceXPConfig(ctx, voiceConfig); err != nil {
		t.Fatalf("save voice XP config: %v", err)
	}
	voiceConfig.ChannelID = "voice-2"
	if err := voice.SaveVoiceXPConfig(ctx, voiceConfig); err != nil {
		t.Fatalf("update voice XP config: %v", err)
	}
	loadedVoice, err := voice.GetVoiceXPConfig(ctx, "guild-1")
	if err != nil || loadedVoice.ChannelID != "voice-2" || loadedVoice.Color != "Blue" {
		t.Fatalf("voice XP config = %#v err=%v", loadedVoice, err)
	}

	if err := text.DeleteTextXPConfig(ctx, "guild-1"); err != nil {
		t.Fatalf("delete text XP config: %v", err)
	}
	if _, err := text.GetTextXPConfig(ctx, "guild-1"); !errors.Is(err, ports.ErrTextXPConfigMissing) {
		t.Fatalf("missing text XP config error = %v", err)
	}
	if err := text.DeleteTextXPConfig(ctx, "guild-1"); !errors.Is(err, ports.ErrTextXPConfigMissing) {
		t.Fatalf("second text XP delete error = %v", err)
	}
	if err := voice.DeleteVoiceXPConfig(ctx, "guild-1"); err != nil {
		t.Fatalf("delete voice XP config: %v", err)
	}
	if _, err := voice.GetVoiceXPConfig(ctx, "guild-1"); !errors.Is(err, ports.ErrVoiceXPConfigMissing) {
		t.Fatalf("missing voice XP config error = %v", err)
	}
	if err := voice.DeleteVoiceXPConfig(ctx, "guild-1"); !errors.Is(err, ports.ErrVoiceXPConfigMissing) {
		t.Fatalf("second voice XP delete error = %v", err)
	}
}

func TestXPRewardRoleMongoIntegrationLifecycle(t *testing.T) {
	database := xpAccrualIntegrationDatabase(t)
	text, err := NewTextXPRewardRoleRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new text XP reward role repository: %v", err)
	}
	voice, err := NewVoiceXPRewardRoleRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new voice XP reward role repository: %v", err)
	}
	ctx := context.Background()
	textConfig := domain.XPRewardRoleConfig{GuildID: " guild-1 ", Level: 5, RoleID: " role-text ", DeleteWhenNot: true}
	if err := text.SaveTextXPRewardRole(ctx, textConfig); err != nil {
		t.Fatalf("save text XP reward role: %v", err)
	}
	textConfig.DeleteWhenNot = false
	if err := text.SaveTextXPRewardRole(ctx, textConfig); err != nil {
		t.Fatalf("replace text XP reward role: %v", err)
	}
	textRoles, err := text.ListTextXPRewardRoles(ctx, "guild-1")
	if err != nil || len(textRoles) != 1 || textRoles[0].Level != 5 || textRoles[0].RoleID != "role-text" || textRoles[0].DeleteWhenNot {
		t.Fatalf("text XP reward roles = %#v err=%v", textRoles, err)
	}

	voiceConfig := domain.XPRewardRoleConfig{GuildID: "guild-1", Level: 7, RoleID: "role-voice", DeleteWhenNot: true}
	if err := voice.SaveVoiceXPRewardRole(ctx, voiceConfig); err != nil {
		t.Fatalf("save voice XP reward role: %v", err)
	}
	voiceRoles, err := voice.ListVoiceXPRewardRoles(ctx, "guild-1")
	if err != nil || len(voiceRoles) != 1 || voiceRoles[0] != voiceConfig {
		t.Fatalf("voice XP reward roles = %#v err=%v", voiceRoles, err)
	}

	if err := text.DeleteTextXPRewardRole(ctx, "guild-1", 5, "role-text"); err != nil {
		t.Fatalf("delete text XP reward role: %v", err)
	}
	if err := text.DeleteTextXPRewardRole(ctx, "guild-1", 5, "role-text"); !errors.Is(err, ports.ErrTextXPRewardRoleMissing) {
		t.Fatalf("missing text XP reward role error = %v", err)
	}
	if err := voice.DeleteVoiceXPRewardRole(ctx, "guild-1", 7, "role-voice"); err != nil {
		t.Fatalf("delete voice XP reward role: %v", err)
	}
	if err := voice.DeleteVoiceXPRewardRole(ctx, "guild-1", 7, "role-voice"); !errors.Is(err, ports.ErrVoiceXPRewardRoleMissing) {
		t.Fatalf("missing voice XP reward role error = %v", err)
	}
}

func TestXPAdminMongoIntegrationLifecycle(t *testing.T) {
	database := xpAccrualIntegrationDatabase(t)
	repository, err := NewXPAdminRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new XP admin repository: %v", err)
	}
	ctx := context.Background()
	textOne := domain.XPProfile{GuildID: "guild-1", UserID: "user-1", XP: 25, Level: 2}
	textTwo := domain.XPProfile{GuildID: "guild-1", UserID: "user-2", XP: 50, Level: 3}
	voiceOne := domain.XPProfile{GuildID: "guild-1", UserID: "user-1", XP: 75, Level: 4}
	voiceTwo := domain.XPProfile{GuildID: "guild-1", UserID: "user-2", XP: 100, Level: 5}
	for _, profile := range []domain.XPProfile{textOne, textTwo} {
		if err := repository.SaveTextXPProfile(ctx, profile); err != nil {
			t.Fatalf("save text XP profile %#v: %v", profile, err)
		}
	}
	for _, profile := range []domain.XPProfile{voiceOne, voiceTwo} {
		if err := repository.SaveVoiceXPProfile(ctx, profile); err != nil {
			t.Fatalf("save voice XP profile %#v: %v", profile, err)
		}
	}
	loadedText, err := repository.GetTextXPProfile(ctx, "guild-1", "user-1")
	if err != nil || loadedText.XP != 25 || loadedText.Level != 2 {
		t.Fatalf("text XP profile = %#v err=%v", loadedText, err)
	}
	loadedVoice, err := repository.GetVoiceXPProfile(ctx, "guild-1", "user-1")
	if err != nil || loadedVoice.XP != 75 || loadedVoice.Level != 4 {
		t.Fatalf("voice XP profile = %#v err=%v", loadedVoice, err)
	}
	textProfiles, err := repository.ListTextXPProfiles(ctx, "guild-1")
	if err != nil || len(textProfiles) != 2 {
		t.Fatalf("text XP profiles = %#v err=%v", textProfiles, err)
	}
	voiceProfiles, err := repository.ListVoiceXPProfiles(ctx, "guild-1")
	if err != nil || len(voiceProfiles) != 2 {
		t.Fatalf("voice XP profiles = %#v err=%v", voiceProfiles, err)
	}

	if err := repository.MarkVoiceXPJoined(ctx, "guild-1", "user-1"); err != nil {
		t.Fatalf("mark voice XP joined: %v", err)
	}
	if err := repository.MarkVoiceXPLeft(ctx, "guild-1", "user-2"); err != nil {
		t.Fatalf("mark voice XP left: %v", err)
	}
	joined, err := repository.ListJoinedVoiceXPSessions(ctx)
	if err != nil || len(joined) != 1 || joined[0].UserID != "user-1" || joined[0].LeaveJoin != domain.VoiceXPSessionJoined {
		t.Fatalf("joined voice XP sessions = %#v err=%v", joined, err)
	}

	if err := repository.DeleteTextXPProfile(ctx, "guild-1", "user-1"); err != nil {
		t.Fatalf("delete text XP profile: %v", err)
	}
	if err := repository.DeleteTextXPProfile(ctx, "guild-1", "user-1"); !errors.Is(err, ports.ErrTextXPProfileMissing) {
		t.Fatalf("missing text XP profile error = %v", err)
	}
	if err := repository.DeleteVoiceXPProfile(ctx, "guild-1", "user-1"); err != nil {
		t.Fatalf("delete voice XP profile: %v", err)
	}
	if err := repository.DeleteVoiceXPProfile(ctx, "guild-1", "user-1"); !errors.Is(err, ports.ErrVoiceXPProfileMissing) {
		t.Fatalf("missing voice XP profile error = %v", err)
	}
	if err := repository.DeleteTextXPGuild(ctx, "guild-1"); err != nil {
		t.Fatalf("delete text XP guild: %v", err)
	}
	if err := repository.DeleteTextXPGuild(ctx, "guild-1"); !errors.Is(err, ports.ErrTextXPProfileMissing) {
		t.Fatalf("missing text XP guild error = %v", err)
	}
	if err := repository.DeleteVoiceXPGuild(ctx, "guild-1"); err != nil {
		t.Fatalf("delete voice XP guild: %v", err)
	}
	if err := repository.DeleteVoiceXPGuild(ctx, "guild-1"); !errors.Is(err, ports.ErrVoiceXPProfileMissing) {
		t.Fatalf("missing voice XP guild error = %v", err)
	}
}
