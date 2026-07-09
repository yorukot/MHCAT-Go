package mongo

import (
	"errors"
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestUpdateBuilderInc(t *testing.T) {
	got, err := NewUpdate().Inc("coin", 5).Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	want := bson.D{{Key: "$inc", Value: bson.D{{Key: "coin", Value: int64(5)}}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("update = %#v, want %#v", got, want)
	}
}

func TestUpdateBuilderSet(t *testing.T) {
	got, err := NewUpdate().Set("name", "value").Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	want := bson.D{{Key: "$set", Value: bson.D{{Key: "name", Value: "value"}}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("update = %#v, want %#v", got, want)
	}
}

func TestUpdateBuilderSetOnInsert(t *testing.T) {
	got, err := NewUpdate().SetOnInsert("createdAt", "now").Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	want := bson.D{{Key: "$setOnInsert", Value: bson.D{{Key: "createdAt", Value: "now"}}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("update = %#v, want %#v", got, want)
	}
}

func TestUpdateBuilderUnset(t *testing.T) {
	got, err := NewUpdate().Unset("background").Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	want := bson.D{{Key: "$unset", Value: bson.D{{Key: "background", Value: ""}}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("update = %#v, want %#v", got, want)
	}
}

func TestUpdateBuilderAddToSet(t *testing.T) {
	got, err := NewUpdate().AddToSet("roles", "role-1").Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	want := bson.D{{Key: "$addToSet", Value: bson.D{{Key: "roles", Value: "role-1"}}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("update = %#v, want %#v", got, want)
	}
}

func TestUpdateBuilderPull(t *testing.T) {
	got, err := NewUpdate().Pull("roles", "role-1").Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	want := bson.D{{Key: "$pull", Value: bson.D{{Key: "roles", Value: "role-1"}}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("update = %#v, want %#v", got, want)
	}
}

func TestUpdateBuilderPush(t *testing.T) {
	got, err := NewUpdate().Push("items", "item-1").Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	want := bson.D{{Key: "$push", Value: bson.D{{Key: "items", Value: "item-1"}}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("update = %#v, want %#v", got, want)
	}
}

func TestUpdateBuilderEmptyFails(t *testing.T) {
	_, err := NewUpdate().Build()
	if !errors.Is(err, ErrEmptyUpdate) {
		t.Fatalf("expected ErrEmptyUpdate, got %v", err)
	}
}

func TestUpdateBuilderEmptyFieldFails(t *testing.T) {
	_, err := NewUpdate().Set("", "value").Build()
	if !errors.Is(err, ErrInvalidField) {
		t.Fatalf("expected ErrInvalidField, got %v", err)
	}
}

func TestUpdateBuilderConflictingSameFieldOpsFail(t *testing.T) {
	_, err := NewUpdate().Inc("coin", 1).Set("coin", 2).Build()
	if !errors.Is(err, ErrConflictingUpdate) {
		t.Fatalf("expected ErrConflictingUpdate, got %v", err)
	}
}
