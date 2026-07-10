package ticket

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestTicketDefinitionsMatchLegacyVisibleContract(t *testing.T) {
	permissions := manageMessagesPermission
	want := []commands.Definition{
		{
			Type:                     commands.CommandTypeChatInput,
			Name:                     "私人頻道設置",
			Description:              "設置私人頻道",
			DefaultMemberPermissions: &permissions,
			Ownership:                commands.ManagedOwnership("ticket-foundation", commands.ScopeGuild),
			Options: []commands.Option{
				{
					Type:         commands.OptionTypeChannel,
					Name:         "類別",
					Description:  "輸入私人頻道要在哪個類別開啟!",
					Required:     true,
					ChannelTypes: []int{4},
				},
				{
					Type:        commands.OptionTypeRole,
					Name:        "管理員身分組",
					Description: "輸入管理員身分組(有這個身分組的能夠管理私人頻道)!",
					Required:    true,
				},
			},
		},
		{
			Type:                     commands.CommandTypeChatInput,
			Name:                     "私人頻道刪除",
			Description:              "刪除之前設置的私人頻道",
			DefaultMemberPermissions: &permissions,
			Ownership:                commands.ManagedOwnership("ticket-foundation", commands.ScopeGuild),
		},
	}
	if got := Definitions(); !reflect.DeepEqual(got, want) {
		t.Fatalf("definitions = %#v, want %#v", got, want)
	}
}

func TestTicketModalAndPanelMatchLegacyVisibleContract(t *testing.T) {
	wantModal := responses.Modal{
		CustomID: "setup-id",
		Title:    "私人頻道系統!",
		Rows: []responses.ModalRow{
			{Inputs: []responses.TextInput{{
				CustomID: "ticketcolor",
				Label:    "請輸入嵌入顏色",
				Style:    responses.TextInputStyleShort,
				Required: true,
			}}},
			{Inputs: []responses.TextInput{{
				CustomID: "tickettitle",
				Label:    "請輸入標題",
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
	}
	if got := ticketSetupModal("setup-id"); !reflect.DeepEqual(got, wantModal) {
		t.Fatalf("modal = %#v, want %#v", got, wantModal)
	}

	wantPanel := ports.OutboundMessage{
		Embeds: []ports.OutboundEmbed{{
			Title:       "  私人頻道  ",
			Description: "  請按按鈕  ",
			Color:       0x12AB34,
		}},
		Components: []ports.OutboundComponentRow{{Components: []ports.OutboundComponent{{
			Type:     "button",
			CustomID: "tic",
			Label:    "🎫 點我創建客服頻道!",
			Style:    "primary",
		}}}},
	}
	if got := ticketPanelOutboundMessage("  私人頻道  ", "  請按按鈕  ", 0x12AB34); !reflect.DeepEqual(got, wantPanel) {
		t.Fatalf("panel = %#v, want %#v", got, wantPanel)
	}
}

func TestTicketMessagesMatchAuditedLegacyContract(t *testing.T) {
	tests := []struct {
		name string
		got  responses.Message
		want responses.Message
	}{
		{
			name: "setup permission denied",
			got:  ticketPermissionDeniedMessage("訊息管理"),
			want: responses.Message{
				Ephemeral: true,
				Embeds: []responses.Embed{{
					Title: "<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`訊息管理`才能使用此指令",
					Color: legacyDiscordNamedRed,
				}},
			},
		},
		{
			name: "duplicate config",
			got:  ticketDuplicateConfigMessage(),
			want: responses.Message{
				Ephemeral: true,
				Embeds: []responses.Embed{{
					Title:       "__**錯誤**__",
					Description: "您已經註冊一個私人頻道了，如果需要更改，請打\n`<>h 刪除私人頻道`\n將會告訴您如何刪除",
					Color:       legacyDiscordNamedRed,
				}},
			},
		},
		{
			name: "setup success",
			got:  ticketSetupSuccessMessage(),
			want: responses.Message{Embeds: []responses.Embed{{
				Title: "<a:green_tick:994529015652163614> | 成功創建私人頻道",
				Color: legacyDiscordNamedGreen,
			}}},
		},
		{
			name: "ticket already open",
			got:  ticketAlreadyOpenMessage(),
			want: responses.Message{
				Ephemeral: true,
				Embeds: []responses.Embed{{
					Title:       "__**客服頻道**__",
					Description: ":warning: 你已經有一個客服頻道了!",
					Color:       legacyDiscordNamedRed,
				}},
			},
		},
		{
			name: "deleted config",
			got:  ticketDeletedConfigMessage(),
			want: responses.Message{Content: ":x: 這個創建私人頻道的設置已經被刪除了喔，請麻煩管理員重新創建!"},
		},
		{
			name: "ticket open success",
			got:  ticketOpenSuccessMessage(),
			want: responses.Message{
				Ephemeral: true,
				Embeds: []responses.Embed{{
					Title:       "__**頻道**__",
					Description: ":white_check_mark: 你成功開啟了頻道!",
					Color:       legacyTicketOpenGreen,
				}},
			},
		},
		{
			name: "ticket close denied",
			got:  ticketCloseDeniedMessage(),
			want: responses.Message{Embeds: []responses.Embed{{
				Title:       "__**私人頻道**__",
				Description: "你開啟了一個私人頻道，請靜候客服人員的回復!",
				Color:       legacyDiscordNamedRed,
			}}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !reflect.DeepEqual(test.got, test.want) {
				t.Fatalf("message = %#v, want %#v", test.got, test.want)
			}
		})
	}

	wantWelcome := ports.OutboundMessage{
		Content:         "||@everyone||",
		AllowedMentions: ports.AllowedMentions{ParseEveryone: false},
		Embeds: []ports.OutboundEmbed{{
			Title:       "__**私人頻道**__",
			Description: "你開啟了一個私人頻道，請等待客服人員的回復!",
			Color:       legacyDiscordNamedGreen,
		}},
		Components: []ports.OutboundComponentRow{{Components: []ports.OutboundComponent{{
			Type:     "button",
			CustomID: "del",
			Label:    "🗑️ 刪除!",
			Style:    "danger",
		}}}},
	}
	if got := ticketWelcomeMessage(); !reflect.DeepEqual(got, wantWelcome) {
		t.Fatalf("welcome = %#v, want %#v", got, wantWelcome)
	}
}

func TestTicketLegacyRouterWorkflowMatchesAuditedContract(t *testing.T) {
	repo := fakemongo.NewTicketConfigRepository()
	discord := fakediscord.NewSideEffects()
	module := NewModuleWithSideEffects(repo, discord, discord, "fallback-bot")
	router := interactions.NewRouter()
	router.SetCustomIDParser(interactions.DefaultCustomIDParser{})
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register routes: %v", err)
	}

	setup := fakediscord.SlashInteractionWithOptions("私人頻道設置", "", map[string]string{
		"類別":     testCategoryID,
		"管理員身分組": testAdminRole,
	})
	setup.Actor.GuildID = testGuildID
	setup.Actor.PermissionBits = permissionManageMessages
	setupResponder := fakediscord.NewResponder()
	if err := router.Handle(context.Background(), setup, setupResponder); err != nil {
		t.Fatalf("setup route: %v", err)
	}
	if len(setupResponder.Modals) != 1 {
		t.Fatalf("setup modals = %#v", setupResponder.Modals)
	}

	modal := interactions.Interaction{
		Type:      interactions.TypeModal,
		CustomID:  setupResponder.Modals[0].CustomID,
		ChannelID: "panel-channel",
		Actor:     interactions.Actor{GuildID: testGuildID, UserID: "user-1"},
		ModalFields: []customid.ModalField{
			{CustomID: "ticketcolor", Value: "#12ab34"},
			{CustomID: "tickettitle", Value: "  私人頻道  "},
			{CustomID: "ticketcontent", Value: "  請按按鈕  "},
		},
	}
	modalResponder := fakediscord.NewResponder()
	if err := router.Handle(context.Background(), modal, modalResponder); err != nil {
		t.Fatalf("modal route: %v", err)
	}
	if len(modalResponder.Defers) != 1 || !reflect.DeepEqual(modalResponder.Edits, []responses.Message{ticketSetupSuccessMessage()}) {
		t.Fatalf("modal response = defers %#v, edits %#v", modalResponder.Defers, modalResponder.Edits)
	}
	if len(discord.Sent) != 1 || discord.Sent[0].ChannelID != "panel-channel" || !reflect.DeepEqual(discord.Sent[0].Message, ticketPanelOutboundMessage("  私人頻道  ", "  請按按鈕  ", 0x12AB34)) {
		t.Fatalf("ticket panels = %#v", discord.Sent)
	}
	wantConfig := domain.TicketConfig{
		GuildID:        testGuildID,
		CategoryID:     testCategoryID,
		AdminRoleID:    testAdminRole,
		EveryoneRoleID: testGuildID,
	}
	if got, err := repo.GetTicketConfig(context.Background(), testGuildID); err != nil || got != wantConfig {
		t.Fatalf("ticket config = %#v, error = %v", got, err)
	}

	open := ticketButtonInteraction("tic")
	open.ApplicationID = "application-bot"
	openResponder := fakediscord.NewResponder()
	if err := router.Handle(context.Background(), open, openResponder); err != nil {
		t.Fatalf("open route: %v", err)
	}
	allow := int64(permissionViewChannel | permissionSendMessages | permissionReadMessageHistory)
	wantCreate := ports.ChannelCreateRequest{
		GuildID:  testGuildID,
		ParentID: testCategoryID,
		Name:     "user-1",
		Type:     discordChannelTypeGuildText,
		PermissionOverwrites: []ports.PermissionOverwrite{
			{ID: testAdminRole, Type: permissionOverwriteRole, Allow: allow, Deny: permissionCreateInstantInvite},
			{ID: testGuildID, Type: permissionOverwriteRole, Deny: permissionViewChannel},
			{ID: "user-1", Type: permissionOverwriteMember, Allow: allow, Deny: permissionCreateInstantInvite},
			{ID: "application-bot", Type: permissionOverwriteMember, Allow: allow, Deny: permissionCreateInstantInvite},
		},
	}
	if !reflect.DeepEqual(discord.Created, []ports.ChannelCreateRequest{wantCreate}) {
		t.Fatalf("created channels = %#v, want %#v", discord.Created, wantCreate)
	}
	if !reflect.DeepEqual(openResponder.Replies, []responses.Message{ticketOpenSuccessMessage()}) {
		t.Fatalf("open replies = %#v", openResponder.Replies)
	}
	if len(discord.Sent) != 2 || discord.Sent[1].ChannelID != "created-channel-1" || !reflect.DeepEqual(discord.Sent[1].Message, ticketWelcomeMessage()) {
		t.Fatalf("welcome messages = %#v", discord.Sent)
	}

	closeInteraction := interactions.Interaction{
		Type:      interactions.TypeComponent,
		CustomID:  "del",
		ChannelID: "created-channel-1",
		Actor:     interactions.Actor{GuildID: testGuildID, UserID: "user-1"},
	}
	closeResponder := fakediscord.NewResponder()
	if err := router.Handle(context.Background(), closeInteraction, closeResponder); err != nil {
		t.Fatalf("close route: %v", err)
	}
	if !reflect.DeepEqual(discord.Deleted, []string{"created-channel-1"}) || len(closeResponder.Replies) != 0 {
		t.Fatalf("close result = deleted %#v, replies %#v", discord.Deleted, closeResponder.Replies)
	}
}

func TestTicketConcurrentSetupSubmissionsCreateOneConfigAndPanel(t *testing.T) {
	const attempts = 16

	repo := fakemongo.NewTicketConfigRepository()
	discord := fakediscord.NewSideEffects()
	module := NewModuleWithSideEffects(repo, nil, discord, "")
	handler := module.SetupModalHandler()
	interactionsByAttempt := make([]interactions.Interaction, attempts)
	responders := make([]*fakediscord.Responder, attempts)
	for index := range attempts {
		interactionsByAttempt[index] = ticketModalInteraction(t, "#12ab34", fmt.Sprintf("Panel %d", index), "Open a ticket")
		responders[index] = fakediscord.NewResponder()
	}

	start := make(chan struct{})
	errs := make([]error, attempts)
	var wait sync.WaitGroup
	wait.Add(attempts)
	for index := range attempts {
		go func(index int) {
			defer wait.Done()
			<-start
			errs[index] = handler(context.Background(), interactionsByAttempt[index], responders[index])
		}(index)
	}
	close(start)
	wait.Wait()

	successes := 0
	duplicates := 0
	for index, err := range errs {
		if err != nil {
			t.Fatalf("attempt %d: %v", index, err)
		}
		responder := responders[index]
		if len(responder.Defers) != 1 || len(responder.Edits) != 1 {
			t.Fatalf("attempt %d response = defers %#v, edits %#v", index, responder.Defers, responder.Edits)
		}
		switch responder.Edits[0].Embeds[0].Title {
		case "<a:green_tick:994529015652163614> | 成功創建私人頻道":
			successes++
		case "__**錯誤**__":
			duplicates++
		default:
			t.Fatalf("attempt %d unexpected edit = %#v", index, responder.Edits[0])
		}
	}
	if successes != 1 || duplicates != attempts-1 {
		t.Fatalf("successes = %d, duplicates = %d", successes, duplicates)
	}
	if len(discord.Sent) != 1 {
		t.Fatalf("sent panels = %#v", discord.Sent)
	}
	if _, err := repo.GetTicketConfig(context.Background(), testGuildID); err != nil {
		t.Fatalf("get winning config: %v", err)
	}
}
