package repositories

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestBirthdayMongoIntegrationConfigLifecycleAndDuplicateAlignment(t *testing.T) {
	database := xpAccrualIntegrationDatabase(t)
	repository, err := NewBirthdayConfigRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new birthday repository: %v", err)
	}
	ctx := context.Background()
	config := domain.BirthdayConfig{
		GuildID: " guild-1 ", Message: "happy birthday", UTCOffset: " +08:00 ", ChannelID: " channel-1 ",
		EveryoneCanSetBirthdayDate: true, RoleID: " role-1 ",
	}
	if err := repository.SaveBirthdayConfig(ctx, config); err != nil {
		t.Fatalf("save birthday config: %v", err)
	}
	loaded, err := repository.FindBirthdayConfig(ctx, " guild-1 ")
	if err != nil || loaded.GuildID != "guild-1" || loaded.UTCOffset != "+08:00" || loaded.ChannelID != "channel-1" || loaded.RoleID != "role-1" {
		t.Fatalf("birthday config = %#v err=%v", loaded, err)
	}

	if _, err := database.Collection(BirthdayConfigCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "msg", Value: "duplicate"}, {Key: "utc", Value: "+00:00"}, {Key: "channel", Value: "old"},
	}); err != nil {
		t.Fatalf("seed duplicate birthday config: %v", err)
	}
	config.Message = "updated"
	config.ChannelID = " channel-2 "
	config.RoleID = " "
	config.EveryoneCanSetBirthdayDate = false
	if err := repository.SaveBirthdayConfig(ctx, config); err != nil {
		t.Fatalf("update duplicate birthday configs: %v", err)
	}
	matched, err := database.Collection(BirthdayConfigCollectionName).CountDocuments(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "msg", Value: "updated"}, {Key: "channel", Value: "channel-2"}, {Key: "role", Value: nil},
	})
	if err != nil || matched != 2 {
		t.Fatalf("aligned birthday configs = %d err=%v", matched, err)
	}

	if _, err := repository.FindBirthdayConfig(ctx, "missing"); !errors.Is(err, ports.ErrBirthdayConfigMissing) {
		t.Fatalf("missing birthday config error = %v", err)
	}
	if err := repository.SaveBirthdayConfig(ctx, domain.BirthdayConfig{}); !errors.Is(err, domain.ErrInvalidBirthdayConfig) {
		t.Fatalf("invalid birthday config error = %v", err)
	}
}

func TestBirthdayMongoIntegrationProfileLifecycleAndDuplicateAlignment(t *testing.T) {
	database := xpAccrualIntegrationDatabase(t)
	repository, err := NewBirthdayConfigRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new birthday repository: %v", err)
	}
	ctx := context.Background()
	year, month, day, hour, minute := 2000, 2, 29, 9, 30
	profile := domain.BirthdayProfile{
		GuildID: " guild-1 ", UserID: " user-1 ", BirthdayYear: &year, BirthdayMonth: &month,
		BirthdayDay: &day, SendHour: &hour, SendMinute: &minute,
	}
	if err := repository.SaveBirthdayProfile(ctx, profile); err != nil {
		t.Fatalf("save birthday profile: %v", err)
	}
	loaded, err := repository.FindBirthdayProfile(ctx, " guild-1 ", " user-1 ")
	if err != nil || loaded.GuildID != "guild-1" || loaded.UserID != "user-1" || loaded.BirthdayYear == nil || *loaded.BirthdayYear != 2000 {
		t.Fatalf("birthday profile = %#v err=%v", loaded, err)
	}

	if _, err := database.Collection(BirthdayProfileCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "user", Value: "user-1"}, {Key: "birthday_month", Value: 1},
	}); err != nil {
		t.Fatalf("seed duplicate birthday profile: %v", err)
	}
	month, day, hour, minute = 12, 31, 23, 55
	profile.BirthdayYear = nil
	profile.AllowAdmin = true
	if err := repository.SaveBirthdayProfile(ctx, profile); err != nil {
		t.Fatalf("update duplicate birthday profiles: %v", err)
	}
	matched, err := database.Collection(BirthdayProfileCollectionName).CountDocuments(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "user", Value: "user-1"},
		{Key: "birthday_year", Value: nil}, {Key: "birthday_month", Value: 12}, {Key: "birthday_day", Value: 31},
		{Key: "send_msg_hour", Value: 23}, {Key: "send_msg_min", Value: 55}, {Key: "allow", Value: true},
	})
	if err != nil || matched != 2 {
		t.Fatalf("aligned birthday profiles = %d err=%v", matched, err)
	}
	profiles, err := repository.ListBirthdayProfiles(ctx, " guild-1 ")
	if err != nil || len(profiles) != 2 {
		t.Fatalf("birthday profiles = %#v err=%v", profiles, err)
	}
	if err := repository.DeleteBirthdayProfile(ctx, " guild-1 ", " user-1 "); err != nil {
		t.Fatalf("delete birthday profiles: %v", err)
	}
	if err := repository.DeleteBirthdayProfile(ctx, "guild-1", "user-1"); !errors.Is(err, ports.ErrBirthdayProfileMissing) {
		t.Fatalf("missing birthday profile delete error = %v", err)
	}
	if _, err := repository.FindBirthdayProfile(ctx, "guild-1", "user-1"); !errors.Is(err, ports.ErrBirthdayProfileMissing) {
		t.Fatalf("missing birthday profile error = %v", err)
	}
	if err := repository.SaveBirthdayProfile(ctx, domain.BirthdayProfile{}); !errors.Is(err, domain.ErrInvalidBirthdayProfile) {
		t.Fatalf("invalid birthday profile error = %v", err)
	}
	if _, err := repository.ListBirthdayProfiles(ctx, " "); !errors.Is(err, domain.ErrInvalidBirthdayProfile) {
		t.Fatalf("invalid birthday profile list error = %v", err)
	}
}

func TestBirthdayRepositoryWithoutProfilesFailsClosed(t *testing.T) {
	database := xpAccrualIntegrationDatabase(t)
	repository, err := NewBirthdayConfigRepository(database.Collection(BirthdayConfigCollectionName))
	if err != nil {
		t.Fatalf("new config-only birthday repository: %v", err)
	}
	if _, err := repository.FindBirthdayProfile(context.Background(), "guild-1", "user-1"); err == nil {
		t.Fatal("expected missing birthday profile collection error")
	}
}
