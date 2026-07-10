package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
)

func TestUsageCollectionNameMatchesLegacyMongoosePlural(t *testing.T) {
	if UsageCollectionName != "all_use_counts" {
		t.Fatalf("usage collection = %q", UsageCollectionName)
	}
}

func TestNewUsageTrackerRequiresCollectionAndTimeout(t *testing.T) {
	if _, err := NewUsageTracker(nil, time.Second); err == nil {
		t.Fatal("expected nil collection error")
	}
	if _, err := NewUsageTrackerFromDatabase(nil, time.Second); err == nil {
		t.Fatal("expected nil database error")
	}
	if _, err := NewUsageTracker(new(drivermongo.Collection), 0); err == nil {
		t.Fatal("expected non-positive timeout error")
	}
}

func TestUsageIncrementPipelineNormalizesLegacyCount(t *testing.T) {
	pipeline := usageIncrementPipeline("ping")
	if len(pipeline) != 1 {
		t.Fatalf("pipeline = %#v", pipeline)
	}
	set, ok := usageDocumentValue(t, pipeline[0], "$set").(bson.D)
	if !ok {
		t.Fatalf("set stage = %#v", usageDocumentValue(t, pipeline[0], "$set"))
	}
	if got := usageDocumentValue(t, set, "slashcommand_name"); got != "ping" {
		t.Fatalf("slashcommand_name = %#v", got)
	}
	count := usageDocumentValue(t, set, "count")
	countDocument, ok := count.(bson.D)
	if !ok {
		t.Fatalf("count expression = %#v", count)
	}
	add := usageDocumentValue(t, countDocument, "$add")
	operands, ok := add.(bson.A)
	if !ok || len(operands) != 2 || operands[1] != float64(1) {
		t.Fatalf("add operands = %#v", add)
	}
	convert, ok := operands[0].(bson.D)
	if !ok {
		t.Fatalf("convert expression = %#v", operands[0])
	}
	convertOptions, ok := usageDocumentValue(t, convert, "$convert").(bson.D)
	if !ok {
		t.Fatalf("convert options = %#v", usageDocumentValue(t, convert, "$convert"))
	}
	if usageDocumentValue(t, convertOptions, "input") != "$count" || usageDocumentValue(t, convertOptions, "to") != "double" {
		t.Fatalf("convert options = %#v", convertOptions)
	}
	if usageDocumentValue(t, convertOptions, "onError") != float64(0) || usageDocumentValue(t, convertOptions, "onNull") != float64(0) {
		t.Fatalf("convert fallbacks = %#v", convertOptions)
	}
}

func TestUsageTrackerRejectsInvalidEventBeforeUsingCollection(t *testing.T) {
	tracker := &UsageTracker{timeout: time.Second}
	if err := tracker.TrackCommand(context.Background(), ports.UsageEvent{}); !errors.Is(err, ports.ErrInvalidUsageEvent) {
		t.Fatalf("expected ErrInvalidUsageEvent, got %v", err)
	}
}

func usageDocumentValue(t *testing.T, document bson.D, key string) any {
	t.Helper()
	for _, element := range document {
		if element.Key == key {
			return element.Value
		}
	}
	t.Fatalf("missing key %q in %#v", key, document)
	return nil
}
