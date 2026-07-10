package discordgo

import (
	"strings"
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
	if event.Username != "Yoru" || event.UserTag != "Yoru#1234" || event.AvatarURL == "" || event.AvatarIsDefault || event.BotUserID != "bot-1" || event.BotAvatarURL == "" {
		t.Fatalf("author metadata = %#v", event)
	}
	if len(event.Attachments) != 1 || event.Attachments[0].URL != "https://example.test/file.png" {
		t.Fatalf("attachments = %#v", event.Attachments)
	}
	if event.Member == nil || event.Member.UserID != "user-1" || len(event.Member.RoleIDs) != 1 || event.Member.RoleIDs[0] != "role-1" {
		t.Fatalf("member metadata = %#v", event.Member)
	}
}

func TestEventFromMessageMarksDefaultUserAvatar(t *testing.T) {
	event := eventFromMessage(events.TypeMessageCreate, &dgo.Message{
		ID:        "message-1",
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		Author:    &dgo.User{ID: "113779359301998592", Username: "Yoru", Discriminator: "0"},
	}, nil)
	if event.AvatarURL == "" || !event.AvatarIsDefault {
		t.Fatalf("event = %#v", event)
	}
}

func TestEventFromMessageUpdateIncludesCachedOldContent(t *testing.T) {
	event := eventFromMessageUpdate(&dgo.MessageUpdate{
		Message:      &dgo.Message{ID: "message-1", GuildID: "guild-1", ChannelID: "channel-1", Content: "new", Author: &dgo.User{ID: "wrong-user", Username: "Wrong", Bot: true}},
		BeforeUpdate: &dgo.Message{Content: "old", Author: &dgo.User{ID: "user-1", Username: "Yoru", Avatar: "old-avatar"}},
	}, nil)
	if event.Type != events.TypeMessageUpdate || event.Content != "new" || event.OldContent != "old" || !event.HasOldContent || event.UserID != "user-1" || event.Username != "Yoru" || event.IsBot || event.AvatarIsDefault || !strings.Contains(event.AvatarURL, "old-avatar") {
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

func TestEventFromChannelUpdateIncludesCachedBeforePayload(t *testing.T) {
	event := eventFromChannelUpdate(&dgo.ChannelUpdate{
		Channel: &dgo.Channel{
			ID:      "channel-1",
			GuildID: "guild-1",
			Topic:   "new topic",
			PermissionOverwrites: []*dgo.PermissionOverwrite{{
				ID:    "role-1",
				Type:  dgo.PermissionOverwriteTypeRole,
				Allow: dgo.PermissionManageMessages,
			}},
		},
		BeforeUpdate: &dgo.Channel{
			ID:      "channel-1",
			GuildID: "guild-1",
			Topic:   "old topic",
			PermissionOverwrites: []*dgo.PermissionOverwrite{{
				ID:   "role-1",
				Type: dgo.PermissionOverwriteTypeRole,
				Deny: dgo.PermissionVoiceConnect,
			}},
		},
	}, &dgo.User{ID: "bot-1", Username: "MHCAT", Avatar: "bot-avatar"})
	if event.Type != events.TypeChannelUpdate || event.ChannelID != "channel-1" || event.GuildID != "guild-1" || event.BotUserID != "bot-1" || event.BotAvatarURL == "" {
		t.Fatalf("event = %#v", event)
	}
	if event.ChannelUpdate == nil || !event.ChannelUpdate.HasOldChannel || event.ChannelUpdate.OldTopic != "old topic" || event.ChannelUpdate.NewTopic != "new topic" {
		t.Fatalf("channel update = %#v", event.ChannelUpdate)
	}
	if len(event.ChannelUpdate.OldPermissionOverwrites) != 1 || event.ChannelUpdate.OldPermissionOverwrites[0].Deny != dgo.PermissionVoiceConnect {
		t.Fatalf("old overwrites = %#v", event.ChannelUpdate.OldPermissionOverwrites)
	}
	if len(event.ChannelUpdate.NewPermissionOverwrites) != 1 || event.ChannelUpdate.NewPermissionOverwrites[0].Allow != dgo.PermissionManageMessages {
		t.Fatalf("new overwrites = %#v", event.ChannelUpdate.NewPermissionOverwrites)
	}
}

func TestEventFromChannelUpdateRetainsTopicNullState(t *testing.T) {
	event := eventFromChannelUpdate(&dgo.ChannelUpdate{
		Channel:      &dgo.Channel{ID: "channel-1", GuildID: "guild-1", Topic: "null"},
		BeforeUpdate: &dgo.Channel{ID: "channel-1", GuildID: "guild-1"},
	}, nil)
	if event.ChannelUpdate == nil || !event.ChannelUpdate.OldTopicNull || event.ChannelUpdate.NewTopicNull || event.ChannelUpdate.NewTopic != "null" {
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
	bot := &dgo.User{ID: "bot-1", Username: "MHCAT", Avatar: "bot-avatar"}
	if err := state.GuildAdd(&dgo.Guild{
		ID:   "guild-1",
		Name: "測試伺服器",
		Members: []*dgo.Member{{
			GuildID: "guild-1",
			User:    bot,
			Avatar:  "guild-bot-avatar",
		}},
	}); err != nil {
		t.Fatalf("guild add: %v", err)
	}
	state.User = bot
	member := &dgo.Member{
		GuildID: "guild-1",
		User:    &dgo.User{ID: "113779359301998592", Username: "Yoru", Discriminator: "0", Avatar: "user-avatar"},
		Avatar:  "guild-user-avatar",
	}
	event := eventFromMember(events.TypeMemberAdd, member, guildFromState(&dgo.Session{State: state}, member), botFromState(&dgo.Session{State: state}))
	if event.GuildID != "guild-1" || event.GuildName != "測試伺服器" || event.BotUserID != "bot-1" || event.BotAvatarURL == "" || event.Member == nil || event.Member.AccountCreatedAt.IsZero() || event.Member.Discriminator != "0" {
		t.Fatalf("event = %#v", event)
	}
	if !strings.Contains(event.BotAvatarURL, "/guilds/guild-1/users/bot-1/avatars/guild-bot-avatar") ||
		!strings.Contains(event.Member.AvatarURL, "/guilds/guild-1/users/113779359301998592/avatars/guild-user-avatar") {
		t.Fatalf("guild avatars were not preserved: %#v", event)
	}
}

func TestEventFromVoiceState(t *testing.T) {
	state := dgo.NewState()
	state.User = &dgo.User{ID: "bot-1", Username: "MHCAT", Avatar: "bot-avatar"}
	if err := state.GuildAdd(&dgo.Guild{
		ID: "guild-1",
		Channels: []*dgo.Channel{
			{ID: "old", GuildID: "guild-1", Name: "Old Voice"},
			{ID: "new", GuildID: "guild-1", Name: "New Voice"},
		},
	}); err != nil {
		t.Fatalf("guild add: %v", err)
	}
	event := eventFromVoiceState(&dgo.VoiceState{
		GuildID:   "guild-1",
		UserID:    "user-1",
		ChannelID: "new",
		Member:    &dgo.Member{User: &dgo.User{ID: "user-1", Username: "Bot", Avatar: "user-avatar", Bot: true}},
	}, &dgo.VoiceState{GuildID: "guild-1", UserID: "user-1", ChannelID: "old"}, &dgo.Session{State: state})
	if event.Type != events.TypeVoiceState || event.VoiceState == nil || event.VoiceState.ChannelID != "new" || event.VoiceState.BeforeChannel != "old" {
		t.Fatalf("event = %#v", event)
	}
	if event.VoiceState.ChannelName != "New Voice" || event.VoiceState.BeforeChannelName != "Old Voice" {
		t.Fatalf("voice channel names = %#v", event.VoiceState)
	}
	if !event.IsBot || event.Member == nil || !event.Member.IsBot || event.UserID != "user-1" || event.Username != "Bot" || event.AvatarURL == "" || event.BotAvatarURL == "" {
		t.Fatalf("event = %#v", event)
	}
}

func TestEventFromVoiceStateFallsBackToCachedMemberOnLeave(t *testing.T) {
	event := eventFromVoiceState(
		&dgo.VoiceState{GuildID: "guild-1", UserID: "user-1", Member: &dgo.Member{User: &dgo.User{ID: "user-1", Username: "New", Avatar: "new-avatar"}}},
		&dgo.VoiceState{
			GuildID:   "guild-1",
			UserID:    "user-1",
			ChannelID: "old",
			Member:    &dgo.Member{User: &dgo.User{ID: "user-1", Username: "Yoru", Avatar: "user-avatar"}},
		},
		nil,
	)
	if event.Username != "Yoru" || event.AvatarURL == "" || !strings.Contains(event.AvatarURL, "user-avatar") || event.Member == nil || event.VoiceState.BeforeChannel != "old" {
		t.Fatalf("event = %#v", event)
	}
}
