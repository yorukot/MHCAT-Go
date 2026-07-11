package discordgo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	dgo "github.com/bwmarrin/discordgo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

func TestSideEffectClientMessageRESTLifecycle(t *testing.T) {
	type recordedRequest struct {
		method string
		path   string
		body   string
	}
	var requests []recordedRequest
	client := newSideEffectRESTClient(t, responderRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("read message request: %v", err)
		}
		requests = append(requests, recordedRequest{method: request.Method, path: request.URL.Path, body: string(body)})
		switch {
		case request.Method == http.MethodPost && strings.HasSuffix(request.URL.Path, "/users/@me/channels"):
			return responderHTTPResponse(request, http.StatusOK, `{"id":"dm-1","type":1}`), nil
		case request.Method == http.MethodPost && strings.HasSuffix(request.URL.Path, "/channels/dm-1/messages"):
			return responderHTTPResponse(request, http.StatusOK, `{"id":"dm-message","channel_id":"dm-1"}`), nil
		case request.Method == http.MethodPost && strings.HasSuffix(request.URL.Path, "/channels/channel-1/typing"):
			return responderHTTPResponse(request, http.StatusNoContent, ""), nil
		case request.Method == http.MethodPatch && strings.HasSuffix(request.URL.Path, "/channels/channel-1/messages/message-1"):
			return responderHTTPResponse(request, http.StatusOK, `{"id":"message-1","channel_id":"channel-1"}`), nil
		case request.Method == http.MethodDelete && strings.HasSuffix(request.URL.Path, "/channels/channel-1/messages/message-1"):
			return responderHTTPResponse(request, http.StatusNoContent, ""), nil
		case request.Method == http.MethodPut && strings.Contains(request.URL.Path, "/channels/channel-1/messages/message-1/reactions/") && strings.HasSuffix(request.URL.Path, "/@me"):
			return responderHTTPResponse(request, http.StatusNoContent, ""), nil
		default:
			t.Fatalf("unexpected message request: %s %s", request.Method, request.URL.String())
			return nil, nil
		}
	}))

	direct, err := client.SendDirectMessage(context.Background(), "user-1", ports.OutboundMessage{Content: "hello"})
	if err != nil || direct.ChannelID != "dm-1" || direct.MessageID != "dm-message" {
		t.Fatalf("direct message = %#v err=%v", direct, err)
	}
	if err := client.SendTyping(context.Background(), "channel-1"); err != nil {
		t.Fatalf("send typing: %v", err)
	}
	ref := ports.MessageRef{ChannelID: "channel-1", MessageID: "message-1"}
	if err := client.EditMessage(context.Background(), ref, ports.OutboundMessage{Content: "updated"}); err != nil {
		t.Fatalf("edit message: %v", err)
	}
	if err := client.DeleteMessage(context.Background(), ref); err != nil {
		t.Fatalf("delete message: %v", err)
	}
	if err := client.AddReaction(context.Background(), "channel-1", "message-1", "wave"); err != nil {
		t.Fatalf("add reaction: %v", err)
	}
	if len(requests) != 6 {
		t.Fatalf("requests = %#v", requests)
	}
	if !strings.Contains(requests[0].body, `"recipient_id":"user-1"`) || !strings.Contains(requests[1].body, `"content":"hello"`) || !strings.Contains(requests[3].body, `"content":"updated"`) {
		t.Fatalf("request payloads = %#v", requests)
	}
}

func TestSideEffectClientChannelRESTLifecycle(t *testing.T) {
	var createPayload struct {
		Name                 string `json:"name"`
		Type                 int    `json:"type"`
		ParentID             string `json:"parent_id"`
		UserLimit            int    `json:"user_limit"`
		PermissionOverwrites []struct {
			ID    string `json:"id"`
			Type  int    `json:"type"`
			Allow string `json:"allow"`
			Deny  string `json:"deny"`
		} `json:"permission_overwrites"`
	}
	client := newSideEffectRESTClient(t, responderRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		switch {
		case request.Method == http.MethodGet && strings.HasSuffix(request.URL.Path, "/guilds/guild-1/channels"):
			return responderHTTPResponse(request, http.StatusOK, `[{"id":"text-1","guild_id":"guild-1","name":"general","type":0},{"id":"voice-1","guild_id":"guild-1","name":"target","type":2,"parent_id":"category-1"}]`), nil
		case request.Method == http.MethodGet && strings.HasSuffix(request.URL.Path, "/channels/voice-1"):
			return responderHTTPResponse(request, http.StatusOK, `{"id":"voice-1","guild_id":"guild-1","name":"target","type":2,"parent_id":"category-1"}`), nil
		case request.Method == http.MethodPost && strings.HasSuffix(request.URL.Path, "/guilds/guild-1/channels"):
			if err := json.NewDecoder(request.Body).Decode(&createPayload); err != nil {
				t.Fatalf("decode create channel payload: %v", err)
			}
			return responderHTTPResponse(request, http.StatusOK, `{"id":"voice-2","guild_id":"guild-1","name":"created","type":2,"parent_id":"category-1","permission_overwrites":[{"id":"guild-1","type":0,"allow":"1024","deny":"2048"}]}`), nil
		case request.Method == http.MethodPatch && strings.HasSuffix(request.URL.Path, "/channels/voice-2"):
			return responderHTTPResponse(request, http.StatusOK, `{"id":"voice-2","guild_id":"guild-1","name":"renamed","type":2,"parent_id":"category-1"}`), nil
		default:
			t.Fatalf("unexpected channel request: %s %s", request.Method, request.URL.String())
			return nil, nil
		}
	}))

	found, err := client.FindChannelByName(context.Background(), "guild-1", "target", int(dgo.ChannelTypeGuildVoice))
	if err != nil || found.ChannelID != "voice-1" || found.ParentID != "category-1" {
		t.Fatalf("find channel by name = %#v err=%v", found, err)
	}
	loaded, err := client.FindChannelByID(context.Background(), "guild-1", "voice-1")
	if err != nil || loaded.Name != "target" {
		t.Fatalf("find channel by id = %#v err=%v", loaded, err)
	}
	created, err := client.CreateChannel(context.Background(), ports.ChannelCreateRequest{
		GuildID:   "guild-1",
		Name:      "created",
		Type:      int(dgo.ChannelTypeGuildVoice),
		ParentID:  "category-1",
		UserLimit: 4,
		PermissionOverwrites: []ports.PermissionOverwrite{{
			ID: "guild-1", Type: int(dgo.PermissionOverwriteTypeRole), Allow: 1024, Deny: 2048,
		}},
	})
	if err != nil || created.ChannelID != "voice-2" || len(created.PermissionOverwrites) != 1 {
		t.Fatalf("create channel = %#v err=%v", created, err)
	}
	if createPayload.Name != "created" || createPayload.Type != int(dgo.ChannelTypeGuildVoice) || createPayload.ParentID != "category-1" || createPayload.UserLimit != 4 || len(createPayload.PermissionOverwrites) != 1 || createPayload.PermissionOverwrites[0].Allow != "1024" {
		t.Fatalf("create payload = %#v", createPayload)
	}
	renamed, err := client.RenameChannel(context.Background(), "guild-1", "voice-2", "renamed")
	if err != nil || renamed.Name != "renamed" {
		t.Fatalf("rename channel = %#v err=%v", renamed, err)
	}

	state := dgo.NewState()
	if err := state.GuildAdd(&dgo.Guild{ID: "guild-1", VoiceStates: []*dgo.VoiceState{{UserID: "user-1", ChannelID: "voice-1"}, nil, {UserID: "user-2", ChannelID: "voice-1"}, {UserID: "user-3", ChannelID: "other"}}}); err != nil {
		t.Fatalf("seed voice state: %v", err)
	}
	client.Session.session.State = state
	count, err := client.VoiceChannelMemberCount(context.Background(), "guild-1", "voice-1")
	if err != nil || count != 2 {
		t.Fatalf("voice member count=%d err=%v", count, err)
	}
}

func TestSideEffectClientMemberAndModerationRESTLifecycle(t *testing.T) {
	members := make([]*dgo.Member, 0, 21)
	userIDs := make([]string, 0, 21)
	for index := 1; index <= 21; index++ {
		userID := fmt.Sprintf("user-%02d", index)
		userIDs = append(userIDs, userID)
		members = append(members, &dgo.Member{User: &dgo.User{ID: userID, Username: "User" + fmt.Sprint(index), Bot: index == 21}})
	}
	membersBody, err := json.Marshal(members)
	if err != nil {
		t.Fatalf("marshal guild members: %v", err)
	}
	type moderationRequest struct {
		method string
		path   string
		query  string
		body   string
	}
	var moderationRequests []moderationRequest
	client := newSideEffectRESTClient(t, responderRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("read member request: %v", err)
		}
		recorded := moderationRequest{method: request.Method, path: request.URL.Path, query: request.URL.RawQuery, body: string(body)}
		switch {
		case request.Method == http.MethodGet && strings.HasSuffix(request.URL.Path, "/guilds/guild-1/members"):
			return responderHTTPResponse(request, http.StatusOK, string(membersBody)), nil
		case request.Method == http.MethodGet && strings.HasSuffix(request.URL.Path, "/guilds/guild-1/members/user-01"):
			return responderHTTPResponse(request, http.StatusOK, `{"user":{"id":"user-01","username":"Legacy","discriminator":"1234"}}`), nil
		case request.Method == http.MethodGet && strings.HasSuffix(request.URL.Path, "/guilds/guild-1/members/missing"):
			return responderHTTPResponse(request, http.StatusNotFound, `{"message":"Unknown Member","code":10007}`), nil
		case request.Method == http.MethodGet && strings.HasSuffix(request.URL.Path, "/guilds/guild-1/roles"):
			return responderHTTPResponse(request, http.StatusOK, `[{"id":"guild-1","name":"@everyone","position":0},{"id":"role-target","name":"Target","position":1},{"id":"role-bot","name":"Bot","position":2}]`), nil
		case request.Method == http.MethodGet && strings.HasSuffix(request.URL.Path, "/guilds/guild-1/members/bot-1"):
			return responderHTTPResponse(request, http.StatusOK, `{"user":{"id":"bot-1","username":"Bot","bot":true},"roles":["role-bot"]}`), nil
		case request.Method == http.MethodPatch && strings.HasSuffix(request.URL.Path, "/guilds/guild-1/members/user-01"):
			moderationRequests = append(moderationRequests, recorded)
			return responderHTTPResponse(request, http.StatusOK, `{}`), nil
		case request.Method == http.MethodDelete && strings.HasSuffix(request.URL.Path, "/guilds/guild-1/members/user-01"):
			moderationRequests = append(moderationRequests, recorded)
			return responderHTTPResponse(request, http.StatusNoContent, ""), nil
		case request.Method == http.MethodPut && strings.HasSuffix(request.URL.Path, "/guilds/guild-1/bans/user-01"):
			moderationRequests = append(moderationRequests, recorded)
			return responderHTTPResponse(request, http.StatusNoContent, ""), nil
		case request.Method == http.MethodGet && strings.HasSuffix(request.URL.Path, "/guilds/guild-1/audit-logs"):
			return responderHTTPResponse(request, http.StatusOK, `{"audit_log_entries":[{"id":"audit-1","user_id":"actor-1","target_id":"user-01","action_type":20,"reason":"cleanup","options":{"channel_id":"channel-1"}}],"users":[{"id":"actor-1","username":"Actor","discriminator":"4321"}]}`), nil
		default:
			t.Fatalf("unexpected member request: %s %s", request.Method, request.URL.String())
			return nil, nil
		}
	}))
	client.Session.session.State = dgo.NewState()
	client.Session.session.State.User = &dgo.User{ID: "bot-1", Username: "Bot", Bot: true}

	count, err := client.CountNonBotMembers(context.Background(), "guild-1")
	if err != nil || count != 20 {
		t.Fatalf("non-bot members=%d err=%v", count, err)
	}
	tag, ok, err := client.MemberTag(context.Background(), "guild-1", "user-01")
	if err != nil || !ok || tag != "Legacy#1234" {
		t.Fatalf("member tag=%q ok=%v err=%v", tag, ok, err)
	}
	if _, ok, err := client.MemberTag(context.Background(), "guild-1", "missing"); err != nil || ok {
		t.Fatalf("missing member ok=%v err=%v", ok, err)
	}
	tags, err := client.MemberTags(context.Background(), "guild-1", userIDs)
	if err != nil || len(tags) != 21 || tags["user-02"] != "User2" {
		t.Fatalf("member tags=%#v err=%v", tags, err)
	}
	assignable, err := client.CanAssignRole(context.Background(), "guild-1", "role-target")
	if err != nil || !assignable {
		t.Fatalf("assignable=%v err=%v", assignable, err)
	}
	if _, err := client.CanAssignRole(context.Background(), "guild-1", "missing-role"); !errors.Is(err, ports.ErrDiscordRoleMissing) {
		t.Fatalf("missing role error = %v", err)
	}

	voiceID := "voice-1"
	if err := client.MoveMember(context.Background(), "guild-1", "user-01", &voiceID); err != nil {
		t.Fatalf("move member: %v", err)
	}
	if err := client.SetNickname(context.Background(), "guild-1", "user-01", "New Name"); err != nil {
		t.Fatalf("set nickname: %v", err)
	}
	if err := client.KickMember(context.Background(), "guild-1", "user-01", "cleanup"); err != nil {
		t.Fatalf("kick member: %v", err)
	}
	if err := client.BanMember(context.Background(), "guild-1", "user-01", "abuse", 3); err != nil {
		t.Fatalf("ban member: %v", err)
	}
	if len(moderationRequests) != 4 || !strings.Contains(moderationRequests[0].body, `"channel_id":"voice-1"`) || !strings.Contains(moderationRequests[1].body, `"nick":"New Name"`) || moderationRequests[2].query != "reason=cleanup" || moderationRequests[3].query != "delete_message_days=3&reason=abuse" {
		t.Fatalf("moderation requests = %#v", moderationRequests)
	}

	auditEntries, err := client.AuditLog(context.Background(), ports.AuditLogQuery{GuildID: "guild-1", UserID: "actor-1", Before: "audit-2", Action: 20, Limit: 5})
	if err != nil || len(auditEntries) != 1 {
		t.Fatalf("audit entries=%#v err=%v", auditEntries, err)
	}
	entry := auditEntries[0]
	if entry.ID != "audit-1" || entry.UserTag != "Actor#4321" || entry.ChannelID != "channel-1" || entry.Action != 20 {
		t.Fatalf("audit entry = %#v", entry)
	}
}

func newSideEffectRESTClient(t *testing.T, transport http.RoundTripper) SideEffectClient {
	t.Helper()
	session, err := dgo.New("Bot test-token")
	if err != nil {
		t.Fatalf("new discord session: %v", err)
	}
	session.Client = &http.Client{Transport: transport}
	return SideEffectClient{Session: &Session{session: session}}
}
