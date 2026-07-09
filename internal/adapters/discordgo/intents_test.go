package discordgo

import (
	"testing"

	dgo "github.com/bwmarrin/discordgo"
)

func TestDefaultIntentsAreMinimal(t *testing.T) {
	intents := BuildIntents(IntentOptions{})
	if intents != dgo.IntentsGuilds {
		t.Fatalf("expected guilds only, got %v", intents)
	}
}

func TestMessageContentOffByDefault(t *testing.T) {
	intents := BuildIntents(IntentOptions{})
	if intents&dgo.IntentsMessageContent != 0 {
		t.Fatal("message content intent should be off by default")
	}
}

func TestGuildMembersOffByDefault(t *testing.T) {
	intents := BuildIntents(IntentOptions{})
	if intents&dgo.IntentsGuildMembers != 0 {
		t.Fatal("guild members intent should be off by default")
	}
}

func TestExplicitMessageContentIntent(t *testing.T) {
	intents := BuildIntents(IntentOptions{MessageContent: true})
	expected := dgo.IntentsGuilds | dgo.IntentsMessageContent
	if intents != expected {
		t.Fatalf("expected %v, got %v", expected, intents)
	}
}

func TestExplicitGuildMembersIntent(t *testing.T) {
	intents := BuildIntents(IntentOptions{GuildMembers: true})
	expected := dgo.IntentsGuilds | dgo.IntentsGuildMembers
	if intents != expected {
		t.Fatalf("expected %v, got %v", expected, intents)
	}
}

func TestEventIntentsOffByDefault(t *testing.T) {
	intents := BuildIntents(IntentOptions{})
	if intents&dgo.IntentsGuildMessages != 0 || intents&dgo.IntentsGuildMessageReactions != 0 || intents&dgo.IntentsGuildVoiceStates != 0 {
		t.Fatalf("event intents should be off by default: %v", intents)
	}
}

func TestExplicitEventIntents(t *testing.T) {
	intents := BuildIntents(IntentOptions{GuildMessages: true, MessageReactions: true, VoiceStates: true})
	expected := dgo.IntentsGuilds | dgo.IntentsGuildMessages | dgo.IntentsGuildMessageReactions | dgo.IntentsGuildVoiceStates
	if intents != expected {
		t.Fatalf("expected %v, got %v", expected, intents)
	}
	if intents&dgo.IntentsMessageContent != 0 {
		t.Fatal("message content should remain separate from guild message intent")
	}
}
