package repositories

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo/documents"
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

func TestAutoChatConfigUpdatePreservesLegacyFields(t *testing.T) {
	update, err := mhcatAutoChatConfigUpdate(documents.AutoChatConfigDocument{
		Guild:   "guild-1",
		Channel: "channel-1",
	}, true)
	if err != nil {
		t.Fatalf("build update: %v", err)
	}
	set := documentValue(t, update, "$set")
	if value := documentValue(t, set, "channel"); value != "channel-1" {
		t.Fatalf("channel = %#v", value)
	}
	setOnInsert := documentValue(t, update, "$setOnInsert")
	if value := documentValue(t, setOnInsert, "guild"); value != "guild-1" {
		t.Fatalf("guild setOnInsert = %#v", value)
	}
}

func TestAutoChatConfigUpdateOmitsSetOnInsertWhenNotUpserting(t *testing.T) {
	update, err := mhcatAutoChatConfigUpdate(documents.AutoChatConfigDocumentFromDomain(domain.AutoChatConfig{
		GuildID:   "guild-1",
		ChannelID: "channel-1",
	}), false)
	if err != nil {
		t.Fatalf("build update: %v", err)
	}
	if hasKey(update, "$setOnInsert") {
		t.Fatalf("non-upsert update should not include setOnInsert: %#v", update)
	}
}
