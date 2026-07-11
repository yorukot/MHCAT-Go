package repositories

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestOnboardingRepositoryConstructorsRejectNilCollections(t *testing.T) {
	if _, err := NewJoinRoleConfigRepository(nil); err == nil {
		t.Fatal("expected nil join role collection error")
	}
	if _, err := NewJoinMessageConfigRepository(nil); err == nil {
		t.Fatal("expected nil join message collection error")
	}
	if _, err := NewLeaveMessageConfigRepository(nil); err == nil {
		t.Fatal("expected nil leave message collection error")
	}
	if _, err := NewVerificationConfigRepository(nil); err == nil {
		t.Fatal("expected nil verification collection error")
	}
	if _, err := NewAccountAgeConfigRepository(nil); err == nil {
		t.Fatal("expected nil account age collection error")
	}
}

func TestOnboardingMongoIntegrationLifecycle(t *testing.T) {
	database := xpAccrualIntegrationDatabase(t)
	joinRoles, err := NewJoinRoleConfigRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new join role repository: %v", err)
	}
	joinMessages, err := NewJoinMessageConfigRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new join message repository: %v", err)
	}
	leaveMessages, err := NewLeaveMessageConfigRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new leave message repository: %v", err)
	}
	verification, err := NewVerificationConfigRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new verification repository: %v", err)
	}
	accountAge, err := NewAccountAgeConfigRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new account age repository: %v", err)
	}
	ctx := context.Background()

	joinRole := domain.JoinRoleConfig{GuildID: "guild-1", RoleID: "role-1", GiveTo: domain.JoinRoleGiveMembers}
	if err := joinRoles.CreateJoinRoleConfig(ctx, joinRole); err != nil {
		t.Fatalf("create join role: %v", err)
	}
	if err := joinRoles.CreateJoinRoleConfig(ctx, joinRole); !errors.Is(err, ports.ErrJoinRoleConfigExists) {
		t.Fatalf("duplicate join role error = %v", err)
	}
	roles, err := joinRoles.ListJoinRoleConfigs(ctx, "guild-1")
	if err != nil || len(roles) != 1 || roles[0] != joinRole {
		t.Fatalf("join roles = %#v err=%v", roles, err)
	}
	if err := joinRoles.DeleteJoinRoleConfig(ctx, "guild-1", "role-1"); err != nil {
		t.Fatalf("delete join role: %v", err)
	}
	if err := joinRoles.DeleteJoinRoleConfig(ctx, "guild-1", "role-1"); !errors.Is(err, ports.ErrJoinRoleConfigMissing) {
		t.Fatalf("missing join role error = %v", err)
	}

	if _, err := database.Collection(JoinMessageCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "enable", Value: true}, {Key: "channel", Value: "welcome-1"},
		{Key: "message_content", Value: "welcome"}, {Key: "color", Value: "Blue"}, {Key: "img", Value: "https://example.test/welcome.png"},
	}); err != nil {
		t.Fatalf("seed join message: %v", err)
	}
	joinMessage, err := joinMessages.GetJoinMessageConfig(ctx, "guild-1")
	if err != nil || !joinMessage.Deliverable() || joinMessage.ImageURL == "" {
		t.Fatalf("join message = %#v err=%v", joinMessage, err)
	}

	prepared, err := leaveMessages.PrepareLeaveMessageConfig(ctx, "guild-1", "leave-1")
	if err != nil || prepared.ChannelID != "leave-1" {
		t.Fatalf("prepared leave message = %#v err=%v", prepared, err)
	}
	prepared.MessageContent = "goodbye"
	prepared.Title = "Left"
	prepared.Color = "Red"
	if err := leaveMessages.SaveLeaveMessageContent(ctx, prepared); err != nil {
		t.Fatalf("save leave message content: %v", err)
	}
	prepared, err = leaveMessages.PrepareLeaveMessageConfig(ctx, "guild-1", "leave-2")
	if err != nil || prepared.ChannelID != "leave-2" || prepared.MessageContent != "goodbye" {
		t.Fatalf("updated leave message = %#v err=%v", prepared, err)
	}
	loadedLeave, err := leaveMessages.GetLeaveMessageConfig(ctx, "guild-1")
	if err != nil || loadedLeave != prepared {
		t.Fatalf("loaded leave message = %#v err=%v", loadedLeave, err)
	}
	if err := leaveMessages.SaveLeaveMessageContent(ctx, domain.LeaveMessageConfig{GuildID: "missing", MessageContent: "x", Title: "y", Color: "Blue"}); !errors.Is(err, ports.ErrLeaveMessageConfigMissing) {
		t.Fatalf("missing leave message error = %v", err)
	}

	verificationConfig := domain.VerificationConfig{GuildID: "guild-1", RoleID: "verified-1", RenameTemplate: "member-{name}"}
	if err := verification.SaveVerificationConfig(ctx, verificationConfig); err != nil {
		t.Fatalf("save verification config: %v", err)
	}
	verificationConfig.RoleID = "verified-2"
	verificationConfig.RenameTemplate = ""
	if err := verification.SaveVerificationConfig(ctx, verificationConfig); err != nil {
		t.Fatalf("update verification config: %v", err)
	}
	loadedVerification, err := verification.GetVerificationConfig(ctx, "guild-1")
	if err != nil || loadedVerification != verificationConfig {
		t.Fatalf("verification config = %#v err=%v", loadedVerification, err)
	}

	ageConfig, err := accountAge.SaveAccountAgeRequirement(ctx, "guild-1", 3600.5)
	if err != nil || ageConfig.RequiredSeconds != 3600.5 || ageConfig.ChannelID != "" {
		t.Fatalf("account age requirement = %#v err=%v", ageConfig, err)
	}
	ageConfig, err = accountAge.SetAccountAgeLogChannel(ctx, "guild-1", "age-log")
	if err != nil || ageConfig.ChannelID != "age-log" {
		t.Fatalf("account age log = %#v err=%v", ageConfig, err)
	}
	ageConfig, err = accountAge.SaveAccountAgeRequirement(ctx, "guild-1", 7200)
	if err != nil || ageConfig.RequiredSeconds != 7200 || ageConfig.ChannelID != "age-log" {
		t.Fatalf("updated account age = %#v err=%v", ageConfig, err)
	}
	loadedAge, err := accountAge.GetAccountAgeConfig(ctx, "guild-1")
	if err != nil || loadedAge != ageConfig {
		t.Fatalf("loaded account age = %#v err=%v", loadedAge, err)
	}
	if err := accountAge.DeleteAccountAgeLogChannel(ctx, "guild-1"); err != nil {
		t.Fatalf("delete account age log: %v", err)
	}
	loadedAge, err = accountAge.GetAccountAgeConfig(ctx, "guild-1")
	if err != nil || loadedAge.ChannelID != "" {
		t.Fatalf("account age after log delete = %#v err=%v", loadedAge, err)
	}
	if err := accountAge.DeleteAccountAgeConfig(ctx, "guild-1"); err != nil {
		t.Fatalf("delete account age config: %v", err)
	}
	if err := accountAge.DeleteAccountAgeConfig(ctx, "guild-1"); !errors.Is(err, ports.ErrAccountAgeConfigMissing) {
		t.Fatalf("missing account age error = %v", err)
	}
}

func TestVoiceRoomMongoIntegrationLifecycle(t *testing.T) {
	database := xpAccrualIntegrationDatabase(t)
	configs, err := NewVoiceRoomConfigRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new voice room config repository: %v", err)
	}
	locks, err := NewVoiceRoomLockRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new voice room lock repository: %v", err)
	}
	states, err := NewVoiceRoomStateRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new voice room state repository: %v", err)
	}
	ctx := context.Background()

	configOne := domain.VoiceRoomConfig{GuildID: "guild-1", TriggerChannelID: "trigger-1", ParentID: "category-1", Name: "Room {name}", Limit: 4, Lock: true}
	if err := configs.SaveVoiceRoomConfig(ctx, configOne); err != nil {
		t.Fatalf("save first voice room config: %v", err)
	}
	configOne.Limit = 8
	if err := configs.SaveVoiceRoomConfig(ctx, configOne); err != nil {
		t.Fatalf("update first voice room config: %v", err)
	}
	loadedConfig, err := configs.GetVoiceRoomConfigByTrigger(ctx, "guild-1", "trigger-1")
	if err != nil || loadedConfig != configOne {
		t.Fatalf("voice room config = %#v err=%v", loadedConfig, err)
	}
	configTwo := domain.VoiceRoomConfig{GuildID: "guild-1", TriggerChannelID: "trigger-2", ParentID: "category-1", Name: "Second", Limit: 0}
	if err := configs.SaveVoiceRoomConfig(ctx, configTwo); err != nil {
		t.Fatalf("save second voice room config: %v", err)
	}
	if err := configs.DeleteVoiceRoomConfigByTrigger(ctx, "guild-1", "trigger-1"); err != nil {
		t.Fatalf("delete voice room trigger: %v", err)
	}
	if err := configs.DeleteVoiceRoomConfigByTrigger(ctx, "guild-1", "trigger-1"); !errors.Is(err, ports.ErrVoiceRoomConfigMissing) {
		t.Fatalf("missing voice room trigger error = %v", err)
	}
	if err := configs.DeleteVoiceRoomConfigsByParent(ctx, "guild-1", "category-1"); err != nil {
		t.Fatalf("delete voice room category: %v", err)
	}
	if err := configs.DeleteVoiceRoomConfigsByParent(ctx, "guild-1", "category-1"); !errors.Is(err, ports.ErrVoiceRoomConfigMissing) {
		t.Fatalf("missing voice room category error = %v", err)
	}

	lock := domain.VoiceRoomLock{GuildID: "guild-1", ChannelID: "voice-1", Password: "secret", OwnerID: "owner-1", TextChannelID: "text-1", AllowedUserIDs: []string{"owner-1"}}
	if err := locks.SaveVoiceRoomLock(ctx, lock); err != nil {
		t.Fatalf("save voice room lock: %v", err)
	}
	if err := locks.AllowVoiceRoomLockUser(ctx, "guild-1", "voice-1", "user-2"); err != nil {
		t.Fatalf("allow voice room user: %v", err)
	}
	if err := locks.AllowVoiceRoomLockUser(ctx, "guild-1", "voice-1", "user-2"); err != nil {
		t.Fatalf("allow duplicate voice room user: %v", err)
	}
	loadedLock, err := locks.GetVoiceRoomLock(ctx, "guild-1", "voice-1")
	if err != nil || loadedLock.OwnerID != "owner-1" || !loadedLock.HasPassword() || len(loadedLock.AllowedUserIDs) != 2 {
		t.Fatalf("voice room lock = %#v err=%v", loadedLock, err)
	}
	if err := locks.DeleteVoiceRoomLock(ctx, "guild-1", "voice-1"); err != nil {
		t.Fatalf("delete voice room lock: %v", err)
	}
	if err := locks.DeleteVoiceRoomLock(ctx, "guild-1", "voice-1"); !errors.Is(err, ports.ErrVoiceRoomLockMissing) {
		t.Fatalf("missing voice room lock error = %v", err)
	}

	state := domain.VoiceRoomState{GuildID: "guild-1", ChannelID: "voice-1"}
	if err := states.SaveVoiceRoomState(ctx, state); err != nil {
		t.Fatalf("save voice room state: %v", err)
	}
	if err := states.SaveVoiceRoomState(ctx, state); err != nil {
		t.Fatalf("save duplicate voice room state: %v", err)
	}
	loadedState, err := states.GetVoiceRoomState(ctx, "guild-1", "voice-1")
	if err != nil || loadedState != state {
		t.Fatalf("voice room state = %#v err=%v", loadedState, err)
	}
	if err := states.DeleteVoiceRoomState(ctx, "guild-1", "voice-1"); err != nil {
		t.Fatalf("delete voice room state: %v", err)
	}
	if err := states.DeleteVoiceRoomState(ctx, "guild-1", "voice-1"); !errors.Is(err, ports.ErrVoiceRoomStateMissing) {
		t.Fatalf("missing voice room state error = %v", err)
	}
}
