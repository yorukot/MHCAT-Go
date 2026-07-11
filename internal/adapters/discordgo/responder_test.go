package discordgo

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	dgo "github.com/bwmarrin/discordgo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	discordruntime "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/runtime"
)

type recordedResponderRequest struct {
	method string
	url    string
	body   string
}

type responderRoundTripFunc func(*http.Request) (*http.Response, error)

func (f responderRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestInteractionResponderCreatesAndEditsFollowUpByMessageID(t *testing.T) {
	var requests []recordedResponderRequest
	session, err := dgo.New("")
	if err != nil {
		t.Fatalf("new discord session: %v", err)
	}
	session.Client = &http.Client{Transport: responderRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		body, readErr := io.ReadAll(request.Body)
		if readErr != nil {
			t.Fatalf("read request body: %v", readErr)
		}
		requests = append(requests, recordedResponderRequest{method: request.Method, url: request.URL.String(), body: string(body)})
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"id":"loading-1"}`)),
			Request:    request,
		}, nil
	})}
	responder := NewInteractionResponder(session, &dgo.Interaction{AppID: "app-1", Token: "token-1"})
	if err := responder.state.MarkDefer(context.Background(), responses.DeferOptions{}); err != nil {
		t.Fatalf("mark deferred response: %v", err)
	}

	messageID, err := responder.CreateFollowUp(context.Background(), responses.Message{Content: "loading"})
	if err != nil {
		t.Fatalf("create follow-up: %v", err)
	}
	if messageID != "loading-1" {
		t.Fatalf("message ID = %q", messageID)
	}
	if err := responder.EditFollowUp(context.Background(), messageID, responses.Message{Embeds: []responses.Embed{{Title: "result"}}}); err != nil {
		t.Fatalf("edit follow-up: %v", err)
	}
	if err := responder.DeleteFollowUp(context.Background(), messageID); err != nil {
		t.Fatalf("delete follow-up: %v", err)
	}

	if len(requests) != 3 {
		t.Fatalf("requests = %#v", requests)
	}
	if requests[0].method != http.MethodPost || !strings.Contains(requests[0].url, "/webhooks/app-1/token-1?wait=true") || !strings.Contains(requests[0].body, `"content":"loading"`) {
		t.Fatalf("create request = %#v", requests[0])
	}
	if requests[1].method != http.MethodPatch || !strings.Contains(requests[1].url, "/webhooks/app-1/token-1/messages/loading-1") || !strings.Contains(requests[1].body, `"title":"result"`) {
		t.Fatalf("edit request = %#v", requests[1])
	}
	if requests[2].method != http.MethodDelete || !strings.Contains(requests[2].url, "/webhooks/app-1/token-1/messages/loading-1") {
		t.Fatalf("delete request = %#v", requests[2])
	}
}

func TestInteractionResponderRollsBackFailedInitialResponses(t *testing.T) {
	transportErr := errors.New("discord transport failed")
	tests := []struct {
		name string
		call func(*InteractionResponder) error
	}{
		{name: "reply", call: func(r *InteractionResponder) error { return r.Reply(context.Background(), responses.Message{}) }},
		{name: "defer", call: func(r *InteractionResponder) error {
			return r.Defer(context.Background(), responses.DeferOptions{Ephemeral: true})
		}},
		{name: "defer update", call: func(r *InteractionResponder) error { return r.DeferUpdate(context.Background()) }},
		{name: "modal", call: func(r *InteractionResponder) error {
			return r.ShowModal(context.Background(), responses.Modal{CustomID: "modal", Title: "Modal"})
		}},
		{name: "update", call: func(r *InteractionResponder) error { return r.UpdateMessage(context.Background(), responses.Message{}) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			session := newResponderTestSession(t, responderRoundTripFunc(func(*http.Request) (*http.Response, error) {
				return nil, transportErr
			}))
			responder := NewInteractionResponder(session, &dgo.Interaction{ID: "interaction", AppID: "app", Token: "token"})
			if err := test.call(responder); !errors.Is(err, transportErr) {
				t.Fatalf("response error = %v", err)
			}
			if responder.state.Status() != responses.StatusInitial || responder.state.Ephemeral() || !responder.state.Deadline().IsZero() {
				t.Fatalf("state was not rolled back: status=%s ephemeral=%v deadline=%s", responder.state.Status(), responder.state.Ephemeral(), responder.state.Deadline())
			}
		})
	}
}

func TestInteractionResponderSerializesConcurrentInitialResponses(t *testing.T) {
	var requests atomic.Int32
	session := newResponderTestSession(t, responderRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		return responderHTTPResponse(request, http.StatusNoContent, ""), nil
	}))
	responder := NewInteractionResponder(session, &dgo.Interaction{ID: "interaction", AppID: "app", Token: "token"})
	results := make(chan error, 2)
	var wait sync.WaitGroup
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			results <- responder.Reply(context.Background(), responses.Message{Content: "ok"})
		}()
	}
	wait.Wait()
	close(results)
	var succeeded, rejected int
	for err := range results {
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, responses.ErrAlreadyResponded):
			rejected++
		default:
			t.Fatalf("unexpected response error: %v", err)
		}
	}
	if succeeded != 1 || rejected != 1 || requests.Load() != 1 {
		t.Fatalf("succeeded=%d rejected=%d requests=%d", succeeded, rejected, requests.Load())
	}
}

func TestInteractionResponderCanRetryAfterTransportFailure(t *testing.T) {
	transportErr := errors.New("temporary failure")
	var attempts atomic.Int32
	session := newResponderTestSession(t, responderRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		if attempts.Add(1) == 1 {
			return nil, transportErr
		}
		return responderHTTPResponse(request, http.StatusNoContent, ""), nil
	}))
	responder := NewInteractionResponder(session, &dgo.Interaction{ID: "interaction", AppID: "app", Token: "token"})
	if err := responder.Reply(context.Background(), responses.Message{Content: "first"}); !errors.Is(err, transportErr) {
		t.Fatalf("first response error = %v", err)
	}
	if err := responder.Reply(context.Background(), responses.Message{Content: "second"}); err != nil {
		t.Fatalf("retry response: %v", err)
	}
	if responder.state.Status() != responses.StatusReplied || attempts.Load() != 2 {
		t.Fatalf("status=%s attempts=%d", responder.state.Status(), attempts.Load())
	}
}

func TestRuntimeRetriesInitialErrorResponseAfterReplyTransportFailure(t *testing.T) {
	transportErr := errors.New("temporary failure")
	var requestPaths []string
	session := newResponderTestSession(t, responderRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		requestPaths = append(requestPaths, request.URL.Path)
		if len(requestPaths) == 1 {
			return nil, transportErr
		}
		return responderHTTPResponse(request, http.StatusNoContent, ""), nil
	}))
	router := interactions.NewRouter()
	if err := router.RegisterSlash("fail", func(ctx context.Context, _ interactions.Interaction, responder responses.Responder) error {
		return responder.Reply(ctx, responses.Message{Content: "first"})
	}); err != nil {
		t.Fatalf("register route: %v", err)
	}
	dispatcher, err := discordruntime.NewDispatcher(router, nil)
	if err != nil {
		t.Fatalf("new dispatcher: %v", err)
	}
	responder := NewInteractionResponder(session, &dgo.Interaction{ID: "interaction", AppID: "app", Token: "token"})
	err = dispatcher.Dispatch(context.Background(), interactions.Interaction{Type: interactions.TypeSlash, CommandName: "fail"}, responder)
	if !errors.Is(err, transportErr) {
		t.Fatalf("dispatch error = %v", err)
	}
	if len(requestPaths) != 2 || !strings.HasSuffix(requestPaths[0], "/callback") || !strings.HasSuffix(requestPaths[1], "/callback") {
		t.Fatalf("request paths = %#v", requestPaths)
	}
}

func TestInteractionResponderDeferredLifecycleUsesOriginalAndFollowUps(t *testing.T) {
	var requests []recordedResponderRequest
	session := newResponderTestSession(t, responderRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("read request: %v", err)
		}
		requests = append(requests, recordedResponderRequest{method: request.Method, url: request.URL.String(), body: string(body)})
		if request.Method == http.MethodPost && strings.Contains(request.URL.Path, "/callback") {
			return responderHTTPResponse(request, http.StatusNoContent, ""), nil
		}
		return responderHTTPResponse(request, http.StatusOK, `{"id":"message-1"}`), nil
	}))
	responder := NewInteractionResponder(session, &dgo.Interaction{ID: "interaction", AppID: "app", Token: "token"})
	if err := responder.Defer(context.Background(), responses.DeferOptions{Ephemeral: true}); err != nil {
		t.Fatalf("defer: %v", err)
	}
	if err := responder.EditOriginal(context.Background(), responses.Message{Content: "done", Ephemeral: true}); err != nil {
		t.Fatalf("edit original: %v", err)
	}
	if err := responder.FollowUp(context.Background(), responses.Message{Content: "follow"}); err != nil {
		t.Fatalf("follow up: %v", err)
	}
	if err := responder.Error(context.Background(), errors.New("internal detail")); err != nil {
		t.Fatalf("error follow up: %v", err)
	}
	if len(requests) != 4 {
		t.Fatalf("requests = %#v", requests)
	}
	if requests[1].method != http.MethodPatch || !strings.Contains(requests[1].url, "/messages/@original") || !strings.Contains(requests[1].body, `"content":"done"`) {
		t.Fatalf("edit request = %#v", requests[1])
	}
	if requests[2].method != http.MethodPost || requests[3].method != http.MethodPost || !strings.Contains(requests[3].body, `"flags":64`) {
		t.Fatalf("follow-up requests = %#v", requests[2:])
	}
}

func TestInteractionResponderNilReceiverFailsClosed(t *testing.T) {
	var responder *InteractionResponder
	if err := responder.Reply(context.Background(), responses.Message{}); !errors.Is(err, errInteractionResponderNotConfigured) {
		t.Fatalf("reply error = %v", err)
	}
	if err := responder.Error(context.Background(), errors.New("boom")); !errors.Is(err, errInteractionResponderNotConfigured) {
		t.Fatalf("error response = %v", err)
	}
	zeroValue := &InteractionResponder{}
	if err := zeroValue.Reply(context.Background(), responses.Message{}); !errors.Is(err, errInteractionResponderNotConfigured) {
		t.Fatalf("zero-value reply error = %v", err)
	}
	if err := zeroValue.Error(context.Background(), errors.New("boom")); !errors.Is(err, errInteractionResponderNotConfigured) {
		t.Fatalf("zero-value error response = %v", err)
	}
}

func newResponderTestSession(t *testing.T, transport http.RoundTripper) *dgo.Session {
	t.Helper()
	session, err := dgo.New("")
	if err != nil {
		t.Fatalf("new discord session: %v", err)
	}
	session.Client = &http.Client{Transport: transport}
	return session
}

func responderHTTPResponse(request *http.Request, status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    request,
	}
}

func TestInteractionResponseDataConvertsEmbedsAndComponents(t *testing.T) {
	timestamp := time.Date(2026, 7, 4, 1, 2, 3, 0, time.UTC)
	data := interactionResponseData(responses.Message{
		Content: "hello @everyone",
		Files: []responses.File{{
			Name:        "rank.png",
			ContentType: "image/png",
			Data:        []byte("png"),
		}},
		Embeds: []responses.Embed{{
			Title:       "MHCAT",
			Description: "legacy help",
			Color:       0x5865F2,
			Timestamp:   timestamp,
			Author:      &responses.EmbedAuthor{Name: "MHCAT", IconURL: "https://i.imgur.com/AQAodBA.png"},
			Footer:      &responses.EmbedFooter{Text: "tester的查詢"},
			Fields:      []responses.EmbedField{{Name: "field", Value: "value", Inline: true}},
		}},
		Components: []responses.ComponentRow{
			{Components: []responses.Component{{
				Type:        responses.ComponentTypeSelect,
				CustomID:    "helphelphelphelpmenu",
				Placeholder: "📜 選擇命令類別",
				Disabled:    true,
				Options:     []responses.SelectOption{{Label: "實用工具", Value: "實用工具", Emoji: "<:bestpractice:986070549115596950>", Default: true}},
			}}},
			{Components: []responses.Component{{
				Type:  responses.ComponentTypeButton,
				Style: responses.ButtonStyleLink,
				Label: "邀請我",
				URL:   "https://dsc.gg/mhcat",
				Emoji: "<a:catjump:984807173529931837>",
			}}},
		},
	})

	if len(data.Embeds) != 1 || data.Embeds[0].Title != "MHCAT" || len(data.Embeds[0].Fields) != 1 {
		t.Fatalf("embeds = %#v", data.Embeds)
	}
	if len(data.Files) != 1 || data.Files[0].Name != "rank.png" {
		t.Fatalf("files = %#v", data.Files)
	}
	if data.AllowedMentions == nil || len(data.AllowedMentions.Parse) != 0 {
		t.Fatalf("allowed mentions should suppress mentions by default: %#v", data.AllowedMentions)
	}
	if data.Embeds[0].Timestamp != timestamp.Format(time.RFC3339) {
		t.Fatalf("timestamp = %q", data.Embeds[0].Timestamp)
	}
	if len(data.Components) != 2 {
		t.Fatalf("components = %#v", data.Components)
	}
	row, ok := data.Components[0].(dgo.ActionsRow)
	if !ok || len(row.Components) != 1 {
		t.Fatalf("first row = %#v", data.Components[0])
	}
	selectMenu, ok := row.Components[0].(dgo.SelectMenu)
	if !ok {
		t.Fatalf("select component = %#v", row.Components[0])
	}
	if selectMenu.CustomID != "helphelphelphelpmenu" || !selectMenu.Disabled || len(selectMenu.Options) != 1 {
		t.Fatalf("select menu = %#v", selectMenu)
	}
	if !selectMenu.Options[0].Default {
		t.Fatalf("select option default not mapped: %#v", selectMenu.Options[0])
	}
	if selectMenu.Options[0].Emoji == nil || selectMenu.Options[0].Emoji.Name != "bestpractice" || selectMenu.Options[0].Emoji.ID != "986070549115596950" {
		t.Fatalf("select emoji = %#v", selectMenu.Options[0].Emoji)
	}
	buttonRow, ok := data.Components[1].(dgo.ActionsRow)
	if !ok || len(buttonRow.Components) != 1 {
		t.Fatalf("button row = %#v", data.Components[1])
	}
	button, ok := buttonRow.Components[0].(dgo.Button)
	if !ok {
		t.Fatalf("button component = %#v", buttonRow.Components[0])
	}
	if button.Style != dgo.LinkButton || button.URL != "https://dsc.gg/mhcat" || button.Emoji == nil || !button.Emoji.Animated {
		t.Fatalf("button = %#v", button)
	}
}

func TestInteractionResponseDataAllowsExplicitMentions(t *testing.T) {
	data := interactionResponseData(responses.Message{
		Content: "hello <@123> <@&456>",
		AllowedMentions: &responses.AllowedMentions{
			UserIDs:     []string{"123"},
			RoleIDs:     []string{"456"},
			RepliedUser: true,
		},
	})
	if data.AllowedMentions == nil || len(data.AllowedMentions.Parse) != 0 {
		t.Fatalf("allowed mentions = %#v", data.AllowedMentions)
	}
	if len(data.AllowedMentions.Users) != 1 || data.AllowedMentions.Users[0] != "123" || len(data.AllowedMentions.Roles) != 1 || data.AllowedMentions.Roles[0] != "456" || !data.AllowedMentions.RepliedUser {
		t.Fatalf("allowed mentions = %#v", data.AllowedMentions)
	}
}

func TestDiscordButtonStyleConvertsSuccess(t *testing.T) {
	data := interactionResponseData(responses.Message{
		Components: []responses.ComponentRow{{Components: []responses.Component{{
			Type:     responses.ComponentTypeButton,
			CustomID: "botinfoupdate",
			Label:    "更新",
			Style:    responses.ButtonStyleSuccess,
			Emoji:    "<:update:1020532095212335235>",
		}}}},
	})
	row, ok := data.Components[0].(dgo.ActionsRow)
	if !ok || len(row.Components) != 1 {
		t.Fatalf("row = %#v", data.Components)
	}
	button, ok := row.Components[0].(dgo.Button)
	if !ok {
		t.Fatalf("button = %#v", row.Components[0])
	}
	if button.Style != dgo.SuccessButton || button.CustomID != "botinfoupdate" || button.Emoji == nil || button.Emoji.Name != "update" {
		t.Fatalf("button = %#v", button)
	}
}

func TestInteractionResponseDataCanExplicitlyClearAttachments(t *testing.T) {
	data := interactionResponseData(responses.Message{ClearAttachments: true})
	if data.Attachments == nil || len(*data.Attachments) != 0 {
		t.Fatalf("attachments = %#v", data.Attachments)
	}
	edit := webhookEdit(responses.Message{ClearAttachments: true})
	if edit.Attachments == nil || len(*edit.Attachments) != 0 {
		t.Fatalf("edit attachments = %#v", edit.Attachments)
	}
}

func TestModalResponseDataConvertsTextInputs(t *testing.T) {
	data := modalResponseData(responses.Modal{
		CustomID: "mhcat:v1:ticket:setup:c=222222222222222222,r=333333333333333333",
		Title:    "私人頻道系統!",
		Rows: []responses.ModalRow{
			{Inputs: []responses.TextInput{{
				CustomID: "ticketcolor",
				Label:    "請輸入嵌入顏色",
				Style:    responses.TextInputStyleShort,
				Required: true,
			}}},
			{Inputs: []responses.TextInput{{
				CustomID: "ticketcontent",
				Label:    "請輸入內文",
				Style:    responses.TextInputStyleParagraph,
				Required: true,
			}}},
		},
	})

	if data.CustomID == "" || data.Title != "私人頻道系統!" || len(data.Components) != 2 {
		t.Fatalf("modal response data = %#v", data)
	}
	firstRow, ok := data.Components[0].(dgo.ActionsRow)
	if !ok || len(firstRow.Components) != 1 {
		t.Fatalf("first row = %#v", data.Components[0])
	}
	firstInput, ok := firstRow.Components[0].(dgo.TextInput)
	if !ok {
		t.Fatalf("first input = %#v", firstRow.Components[0])
	}
	if firstInput.CustomID != "ticketcolor" || firstInput.Style != dgo.TextInputShort || !firstInput.Required {
		t.Fatalf("first input = %#v", firstInput)
	}
	secondRow, ok := data.Components[1].(dgo.ActionsRow)
	if !ok || len(secondRow.Components) != 1 {
		t.Fatalf("second row = %#v", data.Components[1])
	}
	secondInput, ok := secondRow.Components[0].(dgo.TextInput)
	if !ok {
		t.Fatalf("second input = %#v", secondRow.Components[0])
	}
	if secondInput.CustomID != "ticketcontent" || secondInput.Style != dgo.TextInputParagraph {
		t.Fatalf("second input = %#v", secondInput)
	}
}
