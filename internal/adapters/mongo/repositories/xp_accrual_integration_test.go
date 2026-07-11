package repositories

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	mhcatmongo "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
)

func TestAtomicTextXPAccrualMongoIntegration(t *testing.T) {
	database := xpAccrualIntegrationDatabase(t)
	repository, err := NewXPAdminRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	ctx := context.Background()

	profile, leveled, err := repository.AccrueTextXP(ctx, "guild-new", "user-new", 20)
	if err != nil {
		t.Fatalf("create accrual: %v", err)
	}
	if leveled || profile.XP != 20 || profile.Level != 0 {
		t.Fatalf("created profile=%#v leveled=%v", profile, leveled)
	}
	assertStoredTextXP(t, database, "guild-new", "user-new", "20", "0")

	_, err = database.Collection(TextXPCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-legacy"},
		{Key: "member", Value: "user-legacy"},
		{Key: "xp", Value: " 95 "},
		{Key: "leavel", Value: int32(0)},
	})
	if err != nil {
		t.Fatalf("seed legacy profile: %v", err)
	}
	profile, leveled, err = repository.AccrueTextXP(ctx, "guild-legacy", "user-legacy", 5)
	if err != nil || leveled || profile.XP != 100 || profile.Level != 0 {
		t.Fatalf("threshold profile=%#v leveled=%v err=%v", profile, leveled, err)
	}
	profile, leveled, err = repository.AccrueTextXP(ctx, "guild-legacy", "user-legacy", 1)
	if err != nil || !leveled || profile.XP != 0 || profile.Level != 1 {
		t.Fatalf("level profile=%#v leveled=%v err=%v", profile, leveled, err)
	}

	_, err = database.Collection(TextXPCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-bool"},
		{Key: "member", Value: "user-bool"},
		{Key: "xp", Value: true},
		{Key: "leavel", Value: false},
	})
	if err != nil {
		t.Fatalf("seed boolean profile: %v", err)
	}
	profile, leveled, err = repository.AccrueTextXP(ctx, "guild-bool", "user-bool", 5)
	if err != nil || leveled || profile.XP != 5 || profile.Level != 0 {
		t.Fatalf("boolean profile=%#v leveled=%v err=%v", profile, leveled, err)
	}
}

func TestAtomicTextXPAccrualMongoIntegrationConcurrent(t *testing.T) {
	database := xpAccrualIntegrationDatabase(t)
	repository, err := NewXPAdminRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	ctx := context.Background()
	_, err = database.Collection(TextXPCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-race"},
		{Key: "member", Value: "user-race"},
		{Key: "xp", Value: "0"},
		{Key: "leavel", Value: "0"},
	})
	if err != nil {
		t.Fatalf("seed profile: %v", err)
	}

	var leveled atomic.Int64
	errorsCh := make(chan error, 11)
	var workers sync.WaitGroup
	for range 11 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			_, didLevel, err := repository.AccrueTextXP(ctx, "guild-race", "user-race", 10)
			if didLevel {
				leveled.Add(1)
			}
			errorsCh <- err
		}()
	}
	workers.Wait()
	close(errorsCh)
	for err := range errorsCh {
		if err != nil {
			t.Fatalf("concurrent accrual: %v", err)
		}
	}
	if leveled.Load() != 1 {
		t.Fatalf("level-up count = %d", leveled.Load())
	}
	assertStoredTextXP(t, database, "guild-race", "user-race", "0", "1")
}

func TestAtomicVoiceXPAccrualMongoIntegrationConcurrent(t *testing.T) {
	database := xpAccrualIntegrationDatabase(t)
	repository, err := NewXPAdminRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	ctx := context.Background()
	_, err = database.Collection(VoiceXPCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-voice"},
		{Key: "member", Value: "user-voice"},
		{Key: "xp", Value: "0"},
		{Key: "leavel", Value: "0"},
		{Key: "leavejoin", Value: "join"},
	})
	if err != nil {
		t.Fatalf("seed profile: %v", err)
	}

	var leveled atomic.Int64
	errorsCh := make(chan error, 21)
	var workers sync.WaitGroup
	for range 21 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			_, active, didLevel, err := repository.AccrueVoiceXP(ctx, "guild-voice", "user-voice", 5)
			if !active && err == nil {
				err = fmt.Errorf("voice profile unexpectedly inactive")
			}
			if didLevel {
				leveled.Add(1)
			}
			errorsCh <- err
		}()
	}
	workers.Wait()
	close(errorsCh)
	for err := range errorsCh {
		if err != nil {
			t.Fatalf("concurrent voice accrual: %v", err)
		}
	}
	if leveled.Load() != 1 {
		t.Fatalf("voice level-up count = %d", leveled.Load())
	}
	assertStoredVoiceXP(t, database, "guild-voice", "user-voice", "5", "1", "join")

	_, err = database.Collection(VoiceXPCollectionName).UpdateOne(
		ctx,
		xpProfileFilter("guild-voice", "user-voice"),
		bson.D{{Key: "$set", Value: bson.D{{Key: "leavejoin", Value: "leave"}}}},
	)
	if err != nil {
		t.Fatalf("mark left: %v", err)
	}
	profile, active, didLevel, err := repository.AccrueVoiceXP(ctx, "guild-voice", "user-voice", 5)
	if err != nil || active || didLevel || profile.XP != 5 || profile.Level != 1 {
		t.Fatalf("left profile=%#v active=%v leveled=%v err=%v", profile, active, didLevel, err)
	}
}

func xpAccrualIntegrationDatabase(t *testing.T) *drivermongo.Database {
	t.Helper()
	if os.Getenv("MHCAT_RUN_MONGO_INTEGRATION_TESTS") != "true" {
		t.Skip("set MHCAT_RUN_MONGO_INTEGRATION_TESTS=true to run")
	}
	uri := os.Getenv("MHCAT_MONGODB_URI")
	if uri == "" {
		t.Fatal("MHCAT_MONGODB_URI is required")
	}
	databaseName := fmt.Sprintf("mhcat_xp_accrual_test_%d", time.Now().UnixNano())
	client, err := mhcatmongo.NewClient(mhcatmongo.Options{
		URI: uri, Database: databaseName, ConnectTimeout: 10 * time.Second, PingTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("new Mongo client: %v", err)
	}
	if err := client.Connect(context.Background()); err != nil {
		t.Fatalf("connect Mongo: %v", err)
	}
	database, err := client.Database()
	if err != nil {
		t.Fatalf("get database: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := database.Drop(ctx); err != nil {
			t.Errorf("drop database: %v", err)
		}
		if err := client.Disconnect(ctx); err != nil {
			t.Errorf("disconnect Mongo: %v", err)
		}
	})
	return database
}

func assertStoredTextXP(t *testing.T, database *drivermongo.Database, guildID string, userID string, wantXP string, wantLevel string) {
	t.Helper()
	var stored struct {
		XP     string `bson:"xp"`
		Leavel string `bson:"leavel"`
	}
	err := database.Collection(TextXPCollectionName).FindOne(context.Background(), xpProfileFilter(guildID, userID)).Decode(&stored)
	if err != nil {
		t.Fatalf("read stored profile: %v", err)
	}
	if stored.XP != wantXP || stored.Leavel != wantLevel {
		t.Fatalf("stored xp=%q level=%q, want xp=%q level=%q", stored.XP, stored.Leavel, wantXP, wantLevel)
	}
}

func assertStoredVoiceXP(t *testing.T, database *drivermongo.Database, guildID string, userID string, wantXP string, wantLevel string, wantState string) {
	t.Helper()
	var stored struct {
		XP        string `bson:"xp"`
		Leavel    string `bson:"leavel"`
		LeaveJoin string `bson:"leavejoin"`
	}
	err := database.Collection(VoiceXPCollectionName).FindOne(context.Background(), xpProfileFilter(guildID, userID)).Decode(&stored)
	if err != nil {
		t.Fatalf("read stored voice profile: %v", err)
	}
	if stored.XP != wantXP || stored.Leavel != wantLevel || stored.LeaveJoin != wantState {
		t.Fatalf("stored voice xp=%q level=%q state=%q, want xp=%q level=%q state=%q", stored.XP, stored.Leavel, stored.LeaveJoin, wantXP, wantLevel, wantState)
	}
}
