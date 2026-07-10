package repositories

import "testing"

func TestAutoChatCollectionNamesMatchLegacyMongooseModels(t *testing.T) {
	if AutoChatConfigCollectionName != "chats" {
		t.Fatalf("config collection = %q", AutoChatConfigCollectionName)
	}
	if AutoChatPaidCollectionName != "chatgpts" {
		t.Fatalf("handoff collection = %q", AutoChatPaidCollectionName)
	}
	if BalanceCollectionName != "chatgpt_gets" {
		t.Fatalf("balance collection = %q", BalanceCollectionName)
	}
}
