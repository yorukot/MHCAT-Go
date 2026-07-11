package mongo

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestNewClientValidation(t *testing.T) {
	if _, err := NewClient(Options{}); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestMongoIntegrationPing(t *testing.T) {
	if os.Getenv("MHCAT_RUN_MONGO_INTEGRATION_TESTS") != "true" {
		t.Skip("set MHCAT_RUN_MONGO_INTEGRATION_TESTS=true to run")
	}
	uri := os.Getenv("MHCAT_MONGODB_URI")
	db := os.Getenv("MHCAT_MONGODB_DATABASE")
	if uri == "" || db == "" {
		t.Fatal("MHCAT_MONGODB_URI and MHCAT_MONGODB_DATABASE are required")
	}
	client, err := NewClient(Options{
		URI:            uri,
		Database:       db,
		ConnectTimeout: 10 * time.Second,
		PingTimeout:    5 * time.Second,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("connect: %v", err)
	}
	if client.DatabaseName() != db {
		t.Fatalf("database name = %q, want %q", client.DatabaseName(), db)
	}
	if health := client.Health(ctx); !health.OK || health.Message != "ok" {
		t.Fatalf("connected health = %#v", health)
	}
	if _, err := client.Collection(" "); err == nil {
		t.Fatal("expected blank collection error")
	}
	collectionName := fmt.Sprintf("mhcat_client_test_%d", time.Now().UnixNano())
	collection, err := client.Collection(collectionName)
	if err != nil {
		t.Fatalf("get collection: %v", err)
	}
	if _, err := collection.InsertOne(ctx, bson.D{{Key: "value", Value: 1}}); err != nil {
		t.Fatalf("seed index collection: %v", err)
	}
	database, err := client.Database()
	if err != nil {
		t.Fatalf("get database: %v", err)
	}
	plan := IndexPlan{Indexes: []IndexSpec{{
		Collection: collectionName,
		Name:       "value_lookup",
		Keys:       []IndexKey{{Field: "value", Order: 1}},
	}}}
	operations := []IndexOperation{{Operation: IndexOperationCreate, Collection: collectionName, IndexName: "value_lookup"}}
	if err := EnsureIndexes(ctx, database, plan, operations); err != nil {
		t.Fatalf("ensure indexes: %v", err)
	}
	indexes, err := ListIndexes(ctx, database, []string{collectionName})
	if err != nil {
		t.Fatalf("list indexes: %v", err)
	}
	if len(indexes[collectionName]) != 2 || indexes[collectionName][1].Name != "value_lookup" {
		t.Fatalf("indexes = %#v", indexes[collectionName])
	}
	if err := EnsureIndexes(ctx, database, IndexPlan{}, operations); err == nil {
		t.Fatal("expected missing index spec error")
	}
	if err := collection.Drop(ctx); err != nil {
		t.Fatalf("drop index collection: %v", err)
	}
	if err := client.Disconnect(ctx); err != nil {
		t.Fatalf("disconnect: %v", err)
	}
	if health := client.Health(ctx); health.OK || health.Message == "" {
		t.Fatalf("disconnected health = %#v", health)
	}
}
