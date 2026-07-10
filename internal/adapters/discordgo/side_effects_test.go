package discordgo

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	dgo "github.com/bwmarrin/discordgo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

func TestSideEffectClientRequiresSession(t *testing.T) {
	client := SideEffectClient{}
	if _, err := client.SendMessage(context.Background(), "channel-1", ports.OutboundMessage{Content: "hello"}); err == nil {
		t.Fatal("expected send error")
	}
	if err := client.SendTyping(context.Background(), "channel-1"); err == nil {
		t.Fatal("expected typing error")
	}
	if err := client.DeleteChannel(context.Background(), "channel-1"); err == nil {
		t.Fatal("expected delete channel error")
	}
	if err := client.AddRole(context.Background(), "guild-1", "user-1", "role-1"); err == nil {
		t.Fatal("expected add role error")
	}
}

func TestCachedEmojiExistsUsesGlobalDiscordState(t *testing.T) {
	state := dgo.NewState()
	if err := state.GuildAdd(&dgo.Guild{ID: "guild-1"}); err != nil {
		t.Fatalf("seed first guild: %v", err)
	}
	if err := state.GuildAdd(&dgo.Guild{ID: "guild-2", Emojis: []*dgo.Emoji{{ID: "emoji-2"}}}); err != nil {
		t.Fatalf("seed second guild: %v", err)
	}
	client := SideEffectClient{Session: &Session{session: &dgo.Session{State: state}}}

	exists, err := client.CachedEmojiExists(context.Background(), " emoji-2 ")
	if err != nil || !exists {
		t.Fatalf("cached emoji: exists=%t err=%v", exists, err)
	}
	exists, err = client.CachedEmojiExists(context.Background(), "missing")
	if err != nil || exists {
		t.Fatalf("missing emoji: exists=%t err=%v", exists, err)
	}
}

func TestFetchMessageUsesDiscordRESTAndMapsNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/channels/channel-1/messages/message-1":
			_, _ = writer.Write([]byte(`{"id":"message-1","channel_id":"channel-1"}`))
		case "/channels/channel-1/messages/missing":
			writer.WriteHeader(http.StatusNotFound)
			_, _ = writer.Write([]byte(`{"message":"Unknown Message","code":10008}`))
		default:
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	previousChannels := dgo.EndpointChannels
	dgo.EndpointChannels = server.URL + "/channels/"
	defer func() { dgo.EndpointChannels = previousChannels }()
	session, err := dgo.New("Bot test-token")
	if err != nil {
		t.Fatalf("new discord session: %v", err)
	}
	client := SideEffectClient{Session: &Session{session: session}}

	message, err := client.FetchMessage(context.Background(), "channel-1", "message-1")
	if err != nil {
		t.Fatalf("fetch message: %v", err)
	}
	if message.ChannelID != "channel-1" || message.MessageID != "message-1" {
		t.Fatalf("message = %#v", message)
	}
	_, err = client.FetchMessage(context.Background(), "channel-1", "missing")
	if !errors.Is(err, ports.ErrDiscordMessageNotFound) {
		t.Fatalf("missing message error = %v", err)
	}
}

func TestRoleWritesUseRESTWithoutCachedMember(t *testing.T) {
	requests := make([]string, 0, 2)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requests = append(requests, request.Method+" "+request.URL.Path)
		writer.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	previousGuilds := dgo.EndpointGuilds
	dgo.EndpointGuilds = server.URL + "/guilds/"
	defer func() { dgo.EndpointGuilds = previousGuilds }()
	session, err := dgo.New("Bot test-token")
	if err != nil {
		t.Fatalf("new discord session: %v", err)
	}
	session.State = dgo.NewState()
	client := SideEffectClient{Session: &Session{session: session}}

	if err := client.AddRole(context.Background(), "guild-1", "uncached-user", "role-1"); err != nil {
		t.Fatalf("add role: %v", err)
	}
	if err := client.RemoveRole(context.Background(), "guild-1", "uncached-user", "role-1"); err != nil {
		t.Fatalf("remove role: %v", err)
	}
	want := []string{
		"PUT /guilds/guild-1/members/uncached-user/roles/role-1",
		"DELETE /guilds/guild-1/members/uncached-user/roles/role-1",
	}
	if len(requests) != len(want) {
		t.Fatalf("requests = %#v", requests)
	}
	for i := range want {
		if requests[i] != want[i] {
			t.Fatalf("request[%d] = %q, want %q", i, requests[i], want[i])
		}
	}
}

func TestOutboundMessageSendIncludesReplyReference(t *testing.T) {
	send := outboundMessageSend(" channel-1 ", ports.OutboundMessage{
		Content:          "hello",
		ReplyToMessageID: " message-1 ",
		AllowedMentions:  ports.AllowedMentions{},
	})
	if send.Reference == nil || send.Reference.ChannelID != "channel-1" || send.Reference.MessageID != "message-1" {
		t.Fatalf("reference = %#v", send.Reference)
	}
	if send.Reference.FailIfNotExists == nil || !*send.Reference.FailIfNotExists {
		t.Fatalf("fail if not exists = %#v", send.Reference.FailIfNotExists)
	}
	if send.AllowedMentions == nil || len(send.AllowedMentions.Parse) != 0 || send.AllowedMentions.RepliedUser {
		t.Fatalf("allowed mentions = %#v", send.AllowedMentions)
	}
}

func TestCoreAllowedMentionsSuppressesByDefault(t *testing.T) {
	allowed := coreAllowedMentions(ports.AllowedMentions{})
	if allowed == nil || len(allowed.Parse) != 0 || len(allowed.Users) != 0 || len(allowed.Roles) != 0 {
		t.Fatalf("allowed mentions = %#v", allowed)
	}
}

func TestLegacyStatsChannelCountsPreserveV14StringComparisons(t *testing.T) {
	total, textChannels, voiceChannels := legacyStatsChannelCounts([]*dgo.Channel{
		{ID: "text", Type: dgo.ChannelTypeGuildText},
		{ID: "voice", Type: dgo.ChannelTypeGuildVoice},
		{ID: "category", Type: dgo.ChannelTypeGuildCategory},
		nil,
	})
	if total != 3 || textChannels != 0 || voiceChannels != 0 {
		t.Fatalf("counts = (%d, %d, %d)", total, textChannels, voiceChannels)
	}
}

func TestLegacyStatsMemberCountsUseCachedBots(t *testing.T) {
	memberCount, userCount, botCount := legacyStatsMemberCounts(&dgo.Guild{
		MemberCount: 10,
		Members: []*dgo.Member{
			{User: &dgo.User{ID: "user-1"}},
			{User: &dgo.User{ID: "bot-1", Bot: true}},
			{User: &dgo.User{ID: "bot-2", Bot: true}},
			nil,
			{},
		},
	})
	if memberCount != 10 || userCount != 8 || botCount != 2 {
		t.Fatalf("counts = (%d, %d, %d)", memberCount, userCount, botCount)
	}
}

func TestGuildStatsUsesLegacyGuildCaches(t *testing.T) {
	state := dgo.NewState()
	if err := state.GuildAdd(&dgo.Guild{
		ID:          "guild-1",
		MemberCount: 10,
		Members: []*dgo.Member{
			{User: &dgo.User{ID: "user-1"}},
			{User: &dgo.User{ID: "bot-1", Bot: true}},
		},
		Channels: []*dgo.Channel{
			{ID: "text-1", GuildID: "guild-1", Type: dgo.ChannelTypeGuildText},
			{ID: "voice-1", GuildID: "guild-1", Type: dgo.ChannelTypeGuildVoice},
			{ID: "category-1", GuildID: "guild-1", Type: dgo.ChannelTypeGuildCategory},
		},
	}); err != nil {
		t.Fatalf("seed guild state: %v", err)
	}
	client := SideEffectClient{Session: &Session{session: &dgo.Session{State: state}}}

	snapshot, err := client.GuildStats(context.Background(), "guild-1")
	if err != nil {
		t.Fatalf("guild stats: %v", err)
	}
	if snapshot.MemberCount != 10 || snapshot.UserCount != 9 || snapshot.BotCount != 1 {
		t.Fatalf("member counts = %#v", snapshot)
	}
	if snapshot.ChannelCount != 3 || snapshot.TextChannelCount != 0 || snapshot.VoiceChannelCount != 0 {
		t.Fatalf("channel counts = %#v", snapshot)
	}
}

func TestGuildStatsMissingCachedGuildDoesNotFallBackToREST(t *testing.T) {
	client := SideEffectClient{Session: &Session{session: &dgo.Session{State: dgo.NewState()}}}
	if _, err := client.GuildStats(context.Background(), "missing-guild"); err == nil {
		t.Fatal("expected missing cached guild error")
	}
}

func TestFindCachedChannelByIDDoesNotFallBackToREST(t *testing.T) {
	state := dgo.NewState()
	if err := state.GuildAdd(&dgo.Guild{
		ID: "guild-1",
		Channels: []*dgo.Channel{
			{ID: "channel-1", GuildID: "guild-1", Name: "stats", Type: dgo.ChannelTypeGuildCategory},
		},
	}); err != nil {
		t.Fatalf("seed guild state: %v", err)
	}
	client := SideEffectClient{Session: &Session{session: &dgo.Session{State: state}}}

	channel, err := client.FindCachedChannelByID(context.Background(), "guild-1", "channel-1")
	if err != nil {
		t.Fatalf("find cached channel: %v", err)
	}
	if channel.ChannelID != "channel-1" || channel.GuildID != "guild-1" || channel.Type != int(dgo.ChannelTypeGuildCategory) {
		t.Fatalf("channel = %#v", channel)
	}
	_, err = client.FindCachedChannelByID(context.Background(), "guild-1", "missing")
	if !errors.Is(err, ports.ErrChannelNotFound) {
		t.Fatalf("missing channel error = %v", err)
	}
}

func TestRoleStatsUsesLegacyGuildMemberCache(t *testing.T) {
	state := dgo.NewState()
	if err := state.GuildAdd(&dgo.Guild{
		ID: "guild-1",
		Roles: []*dgo.Role{
			{ID: "guild-1", Name: "@everyone"},
			{ID: "role-1", Name: "VIP"},
		},
		Members: []*dgo.Member{
			{User: &dgo.User{ID: "user-1"}, Roles: []string{"role-1"}},
			{User: &dgo.User{ID: "user-2"}},
			{User: &dgo.User{ID: "user-3"}, Roles: []string{"role-1"}},
		},
	}); err != nil {
		t.Fatalf("seed guild state: %v", err)
	}
	client := SideEffectClient{Session: &Session{session: &dgo.Session{State: state}}}

	role, err := client.RoleStats(context.Background(), "guild-1", "role-1")
	if err != nil {
		t.Fatalf("role stats: %v", err)
	}
	if role.RoleID != "role-1" || role.RoleName != "VIP" || role.MemberCount != 2 {
		t.Fatalf("role stats = %#v", role)
	}
	everyone, err := client.RoleStats(context.Background(), "guild-1", "guild-1")
	if err != nil {
		t.Fatalf("everyone stats: %v", err)
	}
	if everyone.RoleName != "@everyone" || everyone.MemberCount != 3 {
		t.Fatalf("everyone stats = %#v", everyone)
	}
	missing, err := client.RoleStats(context.Background(), "guild-1", "missing-role")
	if err != nil {
		t.Fatalf("missing role stats: %v", err)
	}
	if missing.RoleID != "missing-role" || missing.RoleName != "" || missing.MemberCount != 0 {
		t.Fatalf("missing role stats = %#v", missing)
	}
}

func TestCachedRoleExistsUsesGuildRoleCache(t *testing.T) {
	state := dgo.NewState()
	if err := state.GuildAdd(&dgo.Guild{
		ID:    "guild-1",
		Roles: []*dgo.Role{{ID: "guild-1", Name: "@everyone"}, {ID: "role-1", Name: "VIP"}},
	}); err != nil {
		t.Fatalf("seed guild state: %v", err)
	}
	client := SideEffectClient{Session: &Session{session: &dgo.Session{State: state}}}

	exists, err := client.CachedRoleExists(context.Background(), "guild-1", "role-1")
	if err != nil || !exists {
		t.Fatalf("cached role = %t, err = %v", exists, err)
	}
	exists, err = client.CachedRoleExists(context.Background(), "guild-1", "missing-role")
	if err != nil || exists {
		t.Fatalf("missing cached role = %t, err = %v", exists, err)
	}
	if _, err := client.CachedRoleExists(context.Background(), "missing-guild", "role-1"); err == nil {
		t.Fatal("expected missing cached guild error")
	}
}

func TestAuditLogEntriesFromDiscordIncludesActorIdentity(t *testing.T) {
	action := dgo.AuditLogActionChannelUpdate
	entries := auditLogEntriesFromDiscord(&dgo.GuildAuditLog{
		Users: []*dgo.User{{
			ID:            "moderator-1",
			Username:      "Mio",
			Discriminator: "1234",
			Avatar:        "avatar-hash",
		}},
		AuditLogEntries: []*dgo.AuditLogEntry{nil, {
			ID:         "audit-1",
			UserID:     "moderator-1",
			TargetID:   "channel-1",
			ActionType: &action,
			Options:    &dgo.AuditLogOptions{ChannelID: "source-channel"},
		}},
	})

	if len(entries) != 1 {
		t.Fatalf("entries = %#v", entries)
	}
	entry := entries[0]
	if entry.ID != "audit-1" || entry.UserID != "moderator-1" || entry.Username != "Mio" || entry.UserTag != "Mio#1234" {
		t.Fatalf("entry = %#v", entry)
	}
	if entry.AvatarURL == "" || entry.TargetID != "channel-1" || entry.ChannelID != "source-channel" || entry.Action != int(action) {
		t.Fatalf("entry metadata = %#v", entry)
	}
}

func TestAuditLogEntriesFromDiscordLeavesDefaultAvatarEmpty(t *testing.T) {
	action := dgo.AuditLogActionChannelUpdate
	entries := auditLogEntriesFromDiscord(&dgo.GuildAuditLog{
		Users: []*dgo.User{{ID: "moderator-1", Username: "Mio"}},
		AuditLogEntries: []*dgo.AuditLogEntry{{
			ID:         "audit-1",
			UserID:     "moderator-1",
			ActionType: &action,
		}},
	})
	if len(entries) != 1 || entries[0].AvatarURL != "" {
		t.Fatalf("entries = %#v", entries)
	}
}

func TestOutboundMessageConversionIncludesEmbedsAndButtons(t *testing.T) {
	embeds := outboundEmbeds([]ports.OutboundEmbed{{
		Title:         "__**私人頻道**__",
		Description:   "你開啟了一個私人頻道，請等待客服人員的回復!",
		Color:         0x00DB00,
		FooterText:    "來自tester的公告",
		FooterIconURL: "https://example.invalid/avatar.png",
		ThumbnailURL:  "https://example.invalid/thumb.png",
		ImageURL:      "https://example.invalid/image.png",
		Timestamp:     time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC),
	}})
	if len(embeds) != 1 || embeds[0].Title != "__**私人頻道**__" || embeds[0].Color != 0x00DB00 {
		t.Fatalf("embeds = %#v", embeds)
	}
	if embeds[0].Footer == nil || embeds[0].Footer.Text != "來自tester的公告" {
		t.Fatalf("footer = %#v", embeds[0].Footer)
	}
	if embeds[0].Thumbnail == nil || embeds[0].Thumbnail.URL != "https://example.invalid/thumb.png" {
		t.Fatalf("thumbnail = %#v", embeds[0].Thumbnail)
	}
	if embeds[0].Image == nil || embeds[0].Image.URL != "https://example.invalid/image.png" {
		t.Fatalf("image = %#v", embeds[0].Image)
	}
	if embeds[0].Timestamp != "2026-07-04T00:00:00Z" {
		t.Fatalf("timestamp = %q", embeds[0].Timestamp)
	}

	components := outboundComponents([]ports.OutboundComponentRow{{Components: []ports.OutboundComponent{{
		Type:     "button",
		CustomID: "del",
		Label:    "🗑️ 刪除!",
		Style:    "danger",
	}}}})
	if len(components) != 1 {
		t.Fatalf("components = %#v", components)
	}
	row, ok := components[0].(dgo.ActionsRow)
	if !ok || len(row.Components) != 1 {
		t.Fatalf("row = %#v", components[0])
	}
	button, ok := row.Components[0].(dgo.Button)
	if !ok {
		t.Fatalf("button = %#v", row.Components[0])
	}
	if button.CustomID != "del" || button.Label != "🗑️ 刪除!" || button.Style != dgo.DangerButton {
		t.Fatalf("button = %#v", button)
	}
}

func TestOutboundMessageConversionIncludesSelectMenusAndEmojis(t *testing.T) {
	components := outboundComponents([]ports.OutboundComponentRow{{Components: []ports.OutboundComponent{{
		Type:        "select",
		CustomID:    "mhcat:v1:poll:owner_menu:",
		Placeholder: "🔧投票發起人操作",
		Options: []ports.OutboundSelectOption{{
			Label:       "公開投票結果",
			Description: "讓所有成員都可以查看該投票結果",
			Value:       "poll_public_result",
			Emoji:       "<:publicrelation:1023972880385585212>",
		}},
	}}}})
	if len(components) != 1 {
		t.Fatalf("components = %#v", components)
	}
	row, ok := components[0].(dgo.ActionsRow)
	if !ok || len(row.Components) != 1 {
		t.Fatalf("row = %#v", components[0])
	}
	menu, ok := row.Components[0].(dgo.SelectMenu)
	if !ok {
		t.Fatalf("menu = %#v", row.Components[0])
	}
	if menu.CustomID != "mhcat:v1:poll:owner_menu:" || menu.Placeholder != "🔧投票發起人操作" || len(menu.Options) != 1 {
		t.Fatalf("menu = %#v", menu)
	}
	if menu.Options[0].Emoji.Name != "publicrelation" || menu.Options[0].Emoji.ID != "1023972880385585212" {
		t.Fatalf("option emoji = %#v", menu.Options[0].Emoji)
	}
}
