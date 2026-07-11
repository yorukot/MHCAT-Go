package repositories

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	mhcatmongo "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
)

func TestDeleteDataMongoIntegrationTargetsAndDuplicateCleanup(t *testing.T) {
	database := deleteDataIntegrationDatabase(t)
	repository, err := NewDeleteDataRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new delete-data repository: %v", err)
	}

	for _, target := range domain.LegacyDeleteDataTargets() {
		t.Run(string(target), func(t *testing.T) {
			for _, candidate := range domain.LegacyDeleteDataTargets() {
				collectionName, ok := DeleteDataCollectionName(candidate)
				if !ok {
					t.Fatalf("collection for %q is missing", candidate)
				}
				collection := database.Collection(collectionName)
				if _, err := collection.InsertMany(context.Background(), []any{
					bson.D{{Key: "guild", Value: "guild-1"}, {Key: "marker", Value: 1}},
					bson.D{{Key: "guild", Value: "guild-1"}, {Key: "marker", Value: 2}},
					bson.D{{Key: "guild", Value: "guild-2"}, {Key: "marker", Value: 3}},
				}); err != nil {
					t.Fatalf("seed %s: %v", collectionName, err)
				}
			}

			request := domain.DeleteDataRequest{GuildID: "guild-1", Target: target}
			if err := repository.DeleteGuildConfig(context.Background(), request); err != nil {
				t.Fatalf("delete %q: %v", target, err)
			}
			for _, candidate := range domain.LegacyDeleteDataTargets() {
				collectionName, _ := DeleteDataCollectionName(candidate)
				collection := database.Collection(collectionName)
				guildOneCount, err := collection.CountDocuments(context.Background(), bson.D{{Key: "guild", Value: "guild-1"}})
				if err != nil {
					t.Fatalf("count guild-1 in %s: %v", collectionName, err)
				}
				wantGuildOneCount := int64(2)
				if candidate == target {
					wantGuildOneCount = 0
				}
				if guildOneCount != wantGuildOneCount {
					t.Fatalf("guild-1 count in %s = %d, want %d", collectionName, guildOneCount, wantGuildOneCount)
				}
				guildTwoCount, err := collection.CountDocuments(context.Background(), bson.D{{Key: "guild", Value: "guild-2"}})
				if err != nil || guildTwoCount != 1 {
					t.Fatalf("guild-2 count in %s = %d, err=%v", collectionName, guildTwoCount, err)
				}
			}
			if err := repository.DeleteGuildConfig(context.Background(), request); !errors.Is(err, ports.ErrDeleteDataTargetMissing) {
				t.Fatalf("second delete error = %v", err)
			}
			for _, candidate := range domain.LegacyDeleteDataTargets() {
				collectionName, _ := DeleteDataCollectionName(candidate)
				if _, err := database.Collection(collectionName).DeleteMany(context.Background(), bson.D{}); err != nil {
					t.Fatalf("clear %s: %v", collectionName, err)
				}
			}
		})
	}
}

func deleteDataIntegrationDatabase(t *testing.T) *drivermongo.Database {
	t.Helper()
	if os.Getenv("MHCAT_RUN_MONGO_INTEGRATION_TESTS") != "true" {
		t.Skip("set MHCAT_RUN_MONGO_INTEGRATION_TESTS=true to run")
	}
	uri := os.Getenv("MHCAT_MONGODB_URI")
	if uri == "" {
		t.Fatal("MHCAT_MONGODB_URI is required")
	}
	client, err := mhcatmongo.NewClient(mhcatmongo.Options{
		URI:            uri,
		Database:       fmt.Sprintf("mhcat_delete_data_test_%d", time.Now().UnixNano()),
		ConnectTimeout: 10 * time.Second,
		PingTimeout:    5 * time.Second,
	})
	if err != nil {
		t.Fatalf("new Mongo client: %v", err)
	}
	if err := client.Connect(context.Background()); err != nil {
		t.Fatalf("connect Mongo: %v", err)
	}
	database, err := client.Database()
	if err != nil {
		t.Fatalf("get Mongo database: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := database.Drop(ctx); err != nil {
			t.Errorf("drop integration database: %v", err)
		}
		if err := client.Disconnect(ctx); err != nil {
			t.Errorf("disconnect integration Mongo: %v", err)
		}
	})
	return database
}
