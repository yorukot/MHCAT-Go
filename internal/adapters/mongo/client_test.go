package mongo

import (
	"context"
	"os"
	"testing"
	"time"
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
	if err := client.Disconnect(ctx); err != nil {
		t.Fatalf("disconnect: %v", err)
	}
}
