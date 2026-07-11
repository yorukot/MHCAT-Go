package repositories

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestRoleSelectionMongoIntegrationLifecycleAndBulkValidation(t *testing.T) {
	database := xpAccrualIntegrationDatabase(t)
	repository, err := NewRoleSelectionRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new role selection repository: %v", err)
	}
	ctx := context.Background()
	reaction := domain.RoleReactionConfig{GuildID: " guild-1 ", MessageID: " message-1 ", React: " emoji ", RoleID: " role-1 "}
	if err := repository.SaveRoleReactionConfig(ctx, reaction); err != nil {
		t.Fatalf("save role reaction: %v", err)
	}
	loadedReaction, err := repository.GetRoleReactionConfig(ctx, " guild-1 ", " message-1 ", " emoji ")
	if err != nil || loadedReaction.RoleID != "role-1" || loadedReaction.GuildID != "guild-1" {
		t.Fatalf("role reaction = %#v err=%v", loadedReaction, err)
	}
	if _, err := database.Collection(RoleReactionCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "message", Value: "message-1"}, {Key: "react", Value: "emoji"}, {Key: "role", Value: "old"},
	}); err != nil {
		t.Fatalf("seed duplicate role reaction: %v", err)
	}
	reaction.RoleID = "role-2"
	if err := repository.SaveRoleReactionConfig(ctx, reaction); err != nil {
		t.Fatalf("align role reactions: %v", err)
	}
	matched, err := database.Collection(RoleReactionCollectionName).CountDocuments(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "message", Value: "message-1"}, {Key: "react", Value: "emoji"}, {Key: "role", Value: "role-2"},
	})
	if err != nil || matched != 2 {
		t.Fatalf("aligned role reactions = %d err=%v", matched, err)
	}
	if err := repository.DeleteRoleReactionConfig(ctx, " guild-1 ", " message-1 ", " emoji "); err != nil {
		t.Fatalf("delete role reactions: %v", err)
	}
	if err := repository.DeleteRoleReactionConfig(ctx, "guild-1", "message-1", "emoji"); !errors.Is(err, ports.ErrRoleReactionConfigMissing) {
		t.Fatalf("missing role reaction delete error = %v", err)
	}

	buttons := []domain.RoleButtonConfig{
		{GuildID: " guild-1 ", Number: " 1 ", RoleID: " role-1 "},
		{GuildID: "guild-1", Number: "2", RoleID: "role-2"},
		{GuildID: "guild-1", Number: "3", RoleID: "role-3"},
	}
	if err := repository.SaveRoleButtonConfigs(ctx, buttons...); err != nil {
		t.Fatalf("bulk save role buttons: %v", err)
	}
	loadedButton, err := repository.GetRoleButtonConfig(ctx, " guild-1 ", " 2 ")
	if err != nil || loadedButton.RoleID != "role-2" || loadedButton.Number != "2" {
		t.Fatalf("role button = %#v err=%v", loadedButton, err)
	}
	if _, err := database.Collection(RoleButtonCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "number", Value: "2"}, {Key: "role", Value: "old"},
	}); err != nil {
		t.Fatalf("seed duplicate role button: %v", err)
	}
	if err := repository.SaveRoleButtonConfigs(ctx, domain.RoleButtonConfig{GuildID: "guild-1", Number: "2", RoleID: "role-new"}); err != nil {
		t.Fatalf("align role buttons: %v", err)
	}
	matched, err = database.Collection(RoleButtonCollectionName).CountDocuments(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "number", Value: "2"}, {Key: "role", Value: "role-new"},
	})
	if err != nil || matched != 2 {
		t.Fatalf("aligned role buttons = %d err=%v", matched, err)
	}

	err = repository.SaveRoleButtonConfigs(ctx,
		domain.RoleButtonConfig{GuildID: "guild-1", Number: "4", RoleID: "role-4"},
		domain.RoleButtonConfig{},
	)
	if !errors.Is(err, domain.ErrInvalidRoleSelectionConfig) {
		t.Fatalf("invalid role button batch error = %v", err)
	}
	if _, err := repository.GetRoleButtonConfig(ctx, "guild-1", "4"); !errors.Is(err, ports.ErrRoleButtonConfigMissing) {
		t.Fatalf("partially written role button error = %v", err)
	}
	if err := repository.SaveRoleButtonConfigs(ctx); err != nil {
		t.Fatalf("empty role button batch: %v", err)
	}
}

func TestAnnouncementMongoIntegrationLifecycleAndDuplicateAlignment(t *testing.T) {
	if _, err := NewAnnouncementConfigRepository(nil, nil); err == nil {
		t.Fatal("expected nil announcement collections error")
	}
	if _, err := NewAnnouncementConfigRepositoryFromDatabase(nil); err == nil {
		t.Fatal("expected nil announcement database error")
	}
	database := xpAccrualIntegrationDatabase(t)
	repository, err := NewAnnouncementConfigRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new announcement repository: %v", err)
	}
	ctx := context.Background()
	created, err := repository.SetAnnouncementChannel(ctx, domain.AnnouncementChannelConfig{GuildID: " guild-1 ", ChannelID: " channel-1 "})
	if err != nil || !created {
		t.Fatalf("create announcement channel created=%t err=%v", created, err)
	}
	channel, err := repository.GetAnnouncementChannel(ctx, " guild-1 ")
	if err != nil || channel.GuildID != "guild-1" || channel.ChannelID != "channel-1" {
		t.Fatalf("announcement channel = %#v err=%v", channel, err)
	}
	if _, err := database.Collection(GuildConfigCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "announcement_id", Value: "old"},
	}); err != nil {
		t.Fatalf("seed duplicate guild announcement: %v", err)
	}
	created, err = repository.SetAnnouncementChannel(ctx, domain.AnnouncementChannelConfig{GuildID: "guild-1", ChannelID: "channel-2"})
	if err != nil || created {
		t.Fatalf("update announcement channel created=%t err=%v", created, err)
	}
	matched, err := database.Collection(GuildConfigCollectionName).CountDocuments(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "announcement_id", Value: "channel-2"},
	})
	if err != nil || matched != 2 {
		t.Fatalf("aligned announcement channels = %d err=%v", matched, err)
	}

	bound := domain.BoundAnnouncementConfig{
		GuildID: " guild-1 ", ChannelID: " bound-1 ", Tag: "@here", Color: "#53FF53", Title: "Announcement",
	}
	created, err = repository.SetBoundAnnouncement(ctx, bound)
	if err != nil || !created {
		t.Fatalf("create bound announcement created=%t err=%v", created, err)
	}
	freshRepository, err := NewAnnouncementConfigRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new fresh announcement repository: %v", err)
	}
	loaded, err := freshRepository.GetBoundAnnouncement(ctx, " guild-1 ", " bound-1 ")
	if err != nil || loaded.GuildID != "guild-1" || loaded.ChannelID != "bound-1" || loaded.Title != "Announcement" {
		t.Fatalf("bound announcement = %#v err=%v", loaded, err)
	}
	if _, err := database.Collection(BoundAnnouncementConfigCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "announcement_id", Value: "bound-1"},
		{Key: "tag", Value: "old"}, {Key: "color", Value: "Random"}, {Key: "title", Value: "old"},
	}); err != nil {
		t.Fatalf("seed duplicate bound announcement: %v", err)
	}
	bound.Title = "Updated"
	created, err = repository.SetBoundAnnouncement(ctx, bound)
	if err != nil || created {
		t.Fatalf("update bound announcement created=%t err=%v", created, err)
	}
	matched, err = database.Collection(BoundAnnouncementConfigCollectionName).CountDocuments(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "announcement_id", Value: "bound-1"}, {Key: "title", Value: "Updated"},
	})
	if err != nil || matched != 2 {
		t.Fatalf("aligned bound announcements = %d err=%v", matched, err)
	}
	loaded, err = repository.GetBoundAnnouncement(ctx, "guild-1", "bound-1")
	if err != nil || loaded.Title != "Updated" {
		t.Fatalf("cached updated bound announcement = %#v err=%v", loaded, err)
	}
	if err := repository.DeleteBoundAnnouncement(ctx, " guild-1 ", " bound-1 "); err != nil {
		t.Fatalf("delete bound announcements: %v", err)
	}
	if _, err := repository.GetBoundAnnouncement(ctx, "guild-1", "bound-1"); !errors.Is(err, ports.ErrBoundAnnouncementConfigMissing) {
		t.Fatalf("cached deleted bound announcement error = %v", err)
	}
	if err := repository.DeleteBoundAnnouncement(ctx, "guild-1", "bound-1"); !errors.Is(err, ports.ErrBoundAnnouncementConfigMissing) {
		t.Fatalf("second bound announcement delete error = %v", err)
	}
	if _, err := repository.SetAnnouncementChannel(ctx, domain.AnnouncementChannelConfig{}); !errors.Is(err, domain.ErrInvalidAnnouncementConfig) {
		t.Fatalf("invalid announcement channel error = %v", err)
	}
}
