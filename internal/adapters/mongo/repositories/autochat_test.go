package repositories

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

func TestAutoChatConfigCollectionName(t *testing.T) {
	if AutoChatConfigCollectionName != "chats" {
		t.Fatalf("autochat collection = %s, want chats", AutoChatConfigCollectionName)
	}
}

func TestNewAutoChatConfigRepositoryRequiresCollection(t *testing.T) {
	if _, err := NewAutoChatConfigRepository(nil); err == nil {
		t.Fatal("expected nil collection error")
	}
}

func TestNewAutoChatConfigRepositoryFromDatabaseRequiresDatabase(t *testing.T) {
	if _, err := NewAutoChatConfigRepositoryFromDatabase(nil); err == nil {
		t.Fatal("expected nil database error")
	}
}

func TestAutoChatConfigGetRejectsBlankGuild(t *testing.T) {
	repo := &AutoChatConfigRepository{}
	if _, err := repo.GetAutoChatConfig(context.Background(), " "); !errors.Is(err, domain.ErrInvalidAutoChatConfig) {
		t.Fatalf("expected ErrInvalidAutoChatConfig, got %v", err)
	}
}
