package discordgo

import (
	"testing"
	"time"

	dgo "github.com/bwmarrin/discordgo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
)

func TestEventFromMessage(t *testing.T) {
	timestamp := time.Unix(1_700_000_000, 0)
	event := eventFromMessage(events.TypeMessageCreate, &dgo.Message{
		ID:          "message-1",
		GuildID:     "guild-1",
		ChannelID:   "channel-1",
		Content:     "hello",
		Timestamp:   timestamp,
		Author:      &dgo.User{ID: "user-1", Username: "Yoru", Discriminator: "1234", Avatar: "avatar-hash", Bot: true},
		Member:      &dgo.Member{Roles: []string{"role-1"}},
		Attachments: []*dgo.MessageAttachment{{URL: "https://example.test/file.png"}},
	}, &dgo.User{ID: "bot-1", Username: "MHCAT", Avatar: "bot-avatar"})
	if event.Type != events.TypeMessageCreate || event.MessageID != "message-1" || event.UserID != "user-1" || !event.IsBot || !event.CreatedAt.Equal(timestamp) {
		t.Fatalf("event = %#v", event)
	}
	if event.Username != "Yoru" || event.UserTag != "Yoru#1234" || event.AvatarURL == "" || event.BotUserID != "bot-1" || event.BotAvatarURL == "" {
		t.Fatalf("author metadata = %#v", event)
	}
	if len(event.Attachments) != 1 || event.Attachments[0].URL != "https://example.test/file.png" {
		t.Fatalf("attachments = %#v", event.Attachments)
	}
	if event.Member == nil || event.Member.UserID != "user-1" || len(event.Member.RoleIDs) != 1 || event.Member.RoleIDs[0] != "role-1" {
		t.Fatalf("member metadata = %#v", event.Member)
	}
}

func TestEventFromMessageUpdateIncludesCachedOldContent(t *testing.T) {
	event := eventFromMessageUpdate(&dgo.MessageUpdate{
		Message:      &dgo.Message{ID: "message-1", GuildID: "guild-1", ChannelID: "channel-1", Content: "new"},
		BeforeUpdate: &dgo.Message{Content: "old", Author: &dgo.User{ID: "user-1", Username: "Yoru"}},
	}, nil)
	if event.Type != events.TypeMessageUpdate || event.Content != "new" || event.OldContent != "old" || !event.HasOldContent || event.UserID != "user-1" || event.Username != "Yoru" {
		t.Fatalf("event = %#v", event)
	}
}

func TestEventFromMessageDeletePrefersCachedMessage(t *testing.T) {
	event := eventFromMessageDelete(&dgo.MessageDelete{
		Message:      &dgo.Message{ID: "message-1", GuildID: "guild-1", ChannelID: "channel-1"},
		BeforeDelete: &dgo.Message{Content: "deleted", Author: &dgo.User{ID: "user-1", Username: "Yoru"}},
	}, nil)
	if event.Type != events.TypeMessageDelete || event.MessageID != "message-1" || event.GuildID != "guild-1" || event.ChannelID != "channel-1" || event.Content != "deleted" || event.UserID != "user-1" {
		t.Fatalf("event = %#v", event)
	}
}

func TestEventFromReaction(t *testing.T) {
	event := eventFromReaction(events.TypeReactionAdd, &dgo.MessageReaction{
		UserID:    "user-1",
		MessageID: "message-1",
		ChannelID: "channel-1",
		GuildID:   "guild-1",
		Emoji:     dgo.Emoji{Name: "cat", ID: "emoji-1"},
	}, &dgo.Member{User: &dgo.User{ID: "user-1", Username: "Yoru"}, Roles: []string{"role-1"}})
	if event.Reaction == nil || event.Reaction.EmojiID != "emoji-1" || event.Member == nil || event.Member.UserID != "user-1" {
		t.Fatalf("event = %#v", event)
	}
}

func TestEventFromMemberIncludesGuildNameFromState(t *testing.T) {
	state := dgo.NewState()
	if err := state.GuildAdd(&dgo.Guild{ID: "guild-1", Name: "測試伺服器"}); err != nil {
		t.Fatalf("guild add: %v", err)
	}
	state.User = &dgo.User{ID: "bot-1", Username: "MHCAT", Avatar: "bot-avatar"}
	member := &dgo.Member{GuildID: "guild-1", User: &dgo.User{ID: "113779359301998592", Username: "Yoru"}}
	event := eventFromMember(events.TypeMemberAdd, member, guildFromState(&dgo.Session{State: state}, member), botFromState(&dgo.Session{State: state}))
	if event.GuildID != "guild-1" || event.GuildName != "測試伺服器" || event.BotUserID != "bot-1" || event.BotAvatarURL == "" || event.Member == nil || event.Member.AccountCreatedAt.IsZero() {
		t.Fatalf("event = %#v", event)
	}
}

func TestEventFromVoiceState(t *testing.T) {
	event := eventFromVoiceState(&dgo.VoiceState{
		GuildID:   "guild-1",
		UserID:    "user-1",
		ChannelID: "new",
		Member:    &dgo.Member{User: &dgo.User{ID: "user-1", Username: "Bot", Bot: true}},
	}, &dgo.VoiceState{ChannelID: "old"})
	if event.Type != events.TypeVoiceState || event.VoiceState == nil || event.VoiceState.ChannelID != "new" || event.VoiceState.BeforeChannel != "old" {
		t.Fatalf("event = %#v", event)
	}
	if !event.IsBot || event.Member == nil || !event.Member.IsBot || event.UserID != "user-1" {
		t.Fatalf("event = %#v", event)
	}
}
