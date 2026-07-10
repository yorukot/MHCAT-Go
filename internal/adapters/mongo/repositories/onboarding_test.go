package repositories

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo/documents"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestLeaveMessagePrepareInsertUpdate(t *testing.T) {
	update, err := leaveMessagePrepareInsertUpdate(domain.LeaveMessageConfig{GuildID: "guild", ChannelID: "channel"})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if !bsonDHas(update, "$set", "channel", "channel") {
		t.Fatalf("missing channel set: %#v", update)
	}
	if !bsonDHas(update, "$setOnInsert", "guild", "guild") {
		t.Fatalf("missing guild setOnInsert: %#v", update)
	}
}

func TestLeaveMessageContentUpdate(t *testing.T) {
	document := documents.LeaveMessageDocumentFromDomain(domain.LeaveMessageConfig{
		GuildID:        "guild",
		MessageContent: "bye",
		Title:          "bye title",
		Color:          "#df1f2f",
	})
	update, err := leaveMessageContentUpdate(document)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if !bsonDHas(update, "$set", "message_content", "bye") || !bsonDHas(update, "$set", "title", "bye title") || !bsonDHas(update, "$set", "color", "#df1f2f") {
		t.Fatalf("missing content set: %#v", update)
	}
}

func TestVerificationConfigUpdate(t *testing.T) {
	document := documents.VerificationDocumentFromDomain(domain.VerificationConfig{
		GuildID:        "guild",
		RoleID:         "role",
		RenameTemplate: "  {name} | MHCAT  ",
	})
	update, err := verificationConfigUpdate(document)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if !bsonDHas(update, "$set", "role", "role") || !bsonDHas(update, "$set", "name", "  {name} | MHCAT  ") {
		t.Fatalf("missing verification set: %#v", update)
	}
}

func TestVerificationConfigInsertUpdate(t *testing.T) {
	document := documents.VerificationDocumentFromDomain(domain.VerificationConfig{GuildID: "guild", RoleID: "role"})
	update, err := verificationConfigInsertUpdate(document)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if !bsonDHas(update, "$setOnInsert", "guild", "guild") || !bsonDHas(update, "$setOnInsert", "role", "role") || !bsonDHas(update, "$setOnInsert", "name", nil) {
		t.Fatalf("missing verification insert fields: %#v", update)
	}
}

func bsonDHas(update bson.D, operation string, field string, value any) bool {
	for _, op := range update {
		if op.Key != operation {
			continue
		}
		fields, ok := op.Value.(bson.D)
		if !ok {
			return false
		}
		for _, item := range fields {
			if item.Key == field && item.Value == value {
				return true
			}
		}
	}
	return false
}
