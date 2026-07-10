package roles

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/roles"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestReactionSetHandlerStoresConfigAndAddsReaction(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	discord := fakediscord.NewSideEffects()
	discord.AssignableRoles["guild-1/role-1"] = true
	seedRoleReactionMessage(discord, "guild-1", "channel-1", "message-1")
	usage := &fakeusage.Tracker{}
	module := NewModule(repo, discord, discord, discord, discord, discord)
	router := interactions.NewRouter(interactions.Usage(usage))
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions(RoleReactionSetCommandName, "", map[string]string{
		"訊息url": "https://discord.com/channels/guild-1/channel-1/message-1",
		"身分組":   "role-1",
		"表情符號":  "✅",
	})
	interaction.Actor.PermissionBits = permissionManageMessages
	responder := fakediscord.NewResponder()

	if err := router.Handle(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != roleSelectionDoneEmoji+" | 表情符號選取身分組成功設定" {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	if _, ok := repo.Reactions["guild-1/message-1/✅"]; !ok {
		t.Fatalf("reaction config not saved: %#v", repo.Reactions)
	}
	if len(discord.Reactions) != 1 || discord.Reactions[0].ChannelID != "channel-1" {
		t.Fatalf("reactions = %#v", discord.Reactions)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != RoleReactionSetCommandName {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestReactionSetHandlerReportsRoleHierarchyBeforeInvalidURL(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	discord := fakediscord.NewSideEffects()
	module := NewModule(repo, discord, discord, discord, discord, discord)
	interaction := fakediscord.SlashInteractionWithOptions(RoleReactionSetCommandName, "", map[string]string{
		"訊息url": "not-a-message-url",
		"身分組":   "role-1",
		"表情符號":  "not-an-emoji",
	})
	interaction.Actor.PermissionBits = permissionManageMessages
	responder := fakediscord.NewResponder()

	if err := module.ReactionSetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	want := roleSelectionErrorPrefix + "我沒有權限給大家這個身分組(請把我的身分組調高)!"
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != want {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestReactionSetHandlerReportsLegacyInvalidEmoji(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	discord := fakediscord.NewSideEffects()
	discord.AssignableRoles["guild-1/role-1"] = true
	seedRoleReactionChannel(discord, "guild-1", "channel-1")
	module := NewModule(repo, discord, discord, discord, discord, discord)
	interaction := fakediscord.SlashInteractionWithOptions(RoleReactionSetCommandName, "", map[string]string{
		"訊息url": "https://discord.com/channels/guild-1/channel-1/message-1",
		"身分組":   "role-1",
		"表情符號":  "not-an-emoji",
	})
	interaction.Actor.PermissionBits = permissionManageMessages
	responder := fakediscord.NewResponder()

	if err := module.ReactionSetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	want := roleSelectionErrorPrefix + "你必須輸入正確的表情符號!(表情符號所在伺服器我必須在裡面!)"
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != want {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	if len(discord.Reactions) != 0 || len(repo.Reactions) != 0 {
		t.Fatalf("invalid emoji side effects: reactions=%#v configs=%#v", discord.Reactions, repo.Reactions)
	}
}

func TestReactionDeleteHandlerDeletesConfig(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	repo.Reactions["guild-1/message-1/✅"] = domain.RoleReactionConfig{GuildID: "guild-1", MessageID: "message-1", React: "✅", RoleID: "role-1"}
	discord := fakediscord.NewSideEffects()
	seedRoleReactionMessage(discord, "guild-1", "channel-1", "message-1")
	module := NewModule(repo, discord, discord, discord, discord, discord)
	interaction := fakediscord.SlashInteractionWithOptions(RoleReactionDeleteCommandName, "", map[string]string{
		"訊息url": "https://discord.com/channels/guild-1/channel-1/message-1",
		"表情符號":  "✅",
	})
	interaction.Actor.PermissionBits = permissionManageMessages
	responder := fakediscord.NewResponder()

	if err := module.ReactionDeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if _, ok := repo.Reactions["guild-1/message-1/✅"]; ok {
		t.Fatalf("reaction config should be deleted")
	}
	if got := responder.Edits[0].Embeds[0].Title; got != "表情符號選取身分組成功刪除" {
		t.Fatalf("title = %q", got)
	}
}

func TestReactionDeleteHandlerRejectsLegacyDiscordAppHost(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	discord := fakediscord.NewSideEffects()
	module := NewModule(repo, discord, discord, discord, discord, discord)
	interaction := fakediscord.SlashInteractionWithOptions(RoleReactionDeleteCommandName, "", map[string]string{
		"訊息url": "https://discordapp.com/channels/guild-1/channel-1/message-1",
		"表情符號":  "✅",
	})
	interaction.Actor.PermissionBits = permissionManageMessages
	responder := fakediscord.NewResponder()

	if err := module.ReactionDeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	want := roleSelectionErrorPrefix + "你輸入的不是一個訊息連結"
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != want {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	if len(discord.Reactions) != 0 {
		t.Fatalf("reactions = %#v", discord.Reactions)
	}
}

func TestReactionSetHandlerReportsMissingLegacyMessage(t *testing.T) {
	tests := []struct {
		name        string
		seedChannel bool
	}{
		{name: "missing cached channel"},
		{name: "missing fetched message", seedChannel: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := fakemongo.NewRoleSelectionRepository()
			discord := fakediscord.NewSideEffects()
			discord.AssignableRoles["guild-1/role-1"] = true
			if test.seedChannel {
				seedRoleReactionChannel(discord, "guild-1", "channel-1")
			}
			module := NewModule(repo, discord, discord, discord, discord, discord)
			interaction := fakediscord.SlashInteractionWithOptions(RoleReactionSetCommandName, "", map[string]string{
				"訊息url": "https://discord.com/channels/guild-1/channel-1/message-1",
				"身分組":   "role-1",
				"表情符號":  "✅",
			})
			interaction.Actor.PermissionBits = permissionManageMessages
			responder := fakediscord.NewResponder()

			if err := module.ReactionSetHandler()(context.Background(), interaction, responder); err != nil {
				t.Fatalf("handler: %v", err)
			}
			want := roleSelectionErrorPrefix + "很抱歉，找不到這個訊息"
			if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != want {
				t.Fatalf("edits = %#v", responder.Edits)
			}
			if len(discord.Reactions) != 0 || len(repo.Reactions) != 0 {
				t.Fatalf("missing target side effects: reactions=%#v configs=%#v", discord.Reactions, repo.Reactions)
			}
		})
	}
}

func TestButtonSetupShowsLegacyModalAndStoresButtonConfigs(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	discord := fakediscord.NewSideEffects()
	discord.AssignableRoles["guild-1/role-1"] = true
	module := NewModuleWithIDGenerator(repo, discord, discord, discord, discord, discord, func() string { return "2026070901011234.567" })
	interaction := fakediscord.SlashInteractionWithOptions(RoleButtonCommandName, "", map[string]string{"身分組": "role-1"})
	interaction.Actor.PermissionBits = permissionManageMessages
	responder := fakediscord.NewResponder()

	if err := module.ButtonSetupHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Modals) != 1 || responder.Modals[0].CustomID != "nal" || responder.Modals[0].Title != "領取身分系統!" {
		t.Fatalf("modals = %#v", responder.Modals)
	}
	input := responder.Modals[0].Rows[0].Inputs[0]
	if input.CustomID != "roleaddcontent2026070901011234.567" || input.Label != "請輸入身分訊息內文" || input.Style != responses.TextInputStyleParagraph || input.Required {
		t.Fatalf("modal input = %#v", input)
	}
	if _, ok := repo.Buttons["guild-1/2026070901011234.567add"]; !ok {
		t.Fatalf("add button config missing: %#v", repo.Buttons)
	}
	if _, ok := repo.Buttons["guild-1/2026070901011234.567delete"]; !ok {
		t.Fatalf("delete button config missing: %#v", repo.Buttons)
	}
}

func TestButtonModalSendsLegacyPanel(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	discord := fakediscord.NewSideEffects()
	module := NewModule(repo, discord, discord, discord, discord, discord)
	interaction := fakediscord.ModalInteraction(interactions.ModalKey{})
	interaction.ChannelID = "channel-1"
	interaction.BotDisplayColor = 0x123456
	interaction.ModalFields = []customid.ModalField{{CustomID: "roleaddcontent2026070901011234.567", Value: "點按鈕領身分"}}
	responder := fakediscord.NewResponder()

	if err := module.ButtonModalHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 || responder.Defers[0].Ephemeral {
		t.Fatalf("legacy role panel modal should defer publicly: %#v", responder.Defers)
	}
	if len(discord.Sent) != 1 {
		t.Fatalf("sent = %#v", discord.Sent)
	}
	panel := discord.Sent[0].Message
	if panel.Embeds[0].Title != "選取身分組" || panel.Embeds[0].Description != "點按鈕領身分" || panel.Embeds[0].Color != 0x123456 {
		t.Fatalf("panel = %#v", panel)
	}
	if panel.Components[0].Components[0].CustomID != "2026070901011234.567add" || panel.Components[0].Components[1].CustomID != "2026070901011234.567delete" {
		t.Fatalf("components = %#v", panel.Components)
	}
}

func TestButtonApplyAddsAndRemovesRole(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	repo.Buttons["guild-1/button-add"] = domain.RoleButtonConfig{GuildID: "guild-1", Number: "button-add", RoleID: "role-1"}
	repo.Buttons["guild-1/button-delete"] = domain.RoleButtonConfig{GuildID: "guild-1", Number: "button-delete", RoleID: "role-1"}
	discord := fakediscord.NewSideEffects()
	discord.AssignableRoles["guild-1/role-1"] = true
	module := NewModule(repo, discord, discord, discord, discord, discord)

	add := fakediscord.ComponentInteractionFromID("button-add")
	add.Actor.PermissionBits = 0
	responder := fakediscord.NewResponder()
	if err := module.ButtonApplyHandler(false)(context.Background(), add, responder); err != nil {
		t.Fatalf("add: %v", err)
	}
	if len(discord.AddedRoles) != 1 || discord.AddedRoles[0].RoleID != "role-1" {
		t.Fatalf("added roles = %#v", discord.AddedRoles)
	}
	remove := fakediscord.ComponentInteractionFromID("button-delete")
	remove.Actor.RoleIDs = []string{"role-1"}
	responder = fakediscord.NewResponder()
	if err := module.ButtonApplyHandler(true)(context.Background(), remove, responder); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if len(discord.RemovedRoles) != 1 || discord.RemovedRoles[0].RoleID != "role-1" {
		t.Fatalf("removed roles = %#v", discord.RemovedRoles)
	}
}

func TestReactionEventsApplyRolesAndIgnoreMissingConfig(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	repo.Reactions["guild-1/message-1/emoji-1"] = domain.RoleReactionConfig{GuildID: "guild-1", MessageID: "message-1", React: "emoji-1", RoleID: "role-1"}
	discord := fakediscord.NewSideEffects()
	module := NewModule(repo, discord, nil, discord, discord, discord)

	err := module.ReactionEventHandler(false)(context.Background(), events.Event{
		Type:      events.TypeReactionAdd,
		GuildID:   "guild-1",
		MessageID: "message-1",
		UserID:    "user-1",
		Reaction:  &events.Reaction{EmojiID: "emoji-1"},
	})
	if err != nil {
		t.Fatalf("add event: %v", err)
	}
	err = module.ReactionEventHandler(true)(context.Background(), events.Event{
		Type:      events.TypeReactionRemove,
		GuildID:   "guild-1",
		MessageID: "missing",
		UserID:    "user-1",
		Reaction:  &events.Reaction{EmojiName: "✅"},
	})
	if err != nil {
		t.Fatalf("missing config should be ignored: %v", err)
	}
	if len(discord.AddedRoles) != 1 || discord.AddedRoles[0].RoleID != "role-1" {
		t.Fatalf("added roles = %#v", discord.AddedRoles)
	}
}

func TestReactionEventsIgnoreBots(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	repo.Reactions["guild-1/message-1/emoji-1"] = domain.RoleReactionConfig{GuildID: "guild-1", MessageID: "message-1", React: "emoji-1", RoleID: "role-1"}
	discord := fakediscord.NewSideEffects()
	module := NewModule(repo, discord, discord, discord, discord, discord)

	for _, remove := range []bool{false, true} {
		err := module.ReactionEventHandler(remove)(context.Background(), events.Event{
			Type:      events.TypeReactionAdd,
			GuildID:   "guild-1",
			MessageID: "message-1",
			UserID:    "bot-1",
			IsBot:     true,
			Reaction:  &events.Reaction{EmojiID: "emoji-1"},
		})
		if err != nil {
			t.Fatalf("handler: %v", err)
		}
	}
	if len(discord.AddedRoles) != 0 || len(discord.RemovedRoles) != 0 || len(discord.DirectMessages) != 0 {
		t.Fatalf("bot reaction side effects: added=%#v removed=%#v dms=%#v", discord.AddedRoles, discord.RemovedRoles, discord.DirectMessages)
	}
}

func TestReactionEventFailureDMUsesLegacyDirectionIcon(t *testing.T) {
	for _, tc := range []struct {
		name       string
		remove     bool
		eventType  events.Type
		wantPrefix string
		missing    bool
	}{
		{name: "add hierarchy", eventType: events.TypeReactionAdd, wantPrefix: "<a:error:980086028113182730> | "},
		{name: "add missing role", eventType: events.TypeReactionAdd, wantPrefix: "<a:error:980086028113182730> | ", missing: true},
		{name: "remove hierarchy", remove: true, eventType: events.TypeReactionRemove, wantPrefix: roleSelectionErrorPrefix},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := fakemongo.NewRoleSelectionRepository()
			repo.Reactions["guild-1/message-1/emoji-1"] = domain.RoleReactionConfig{GuildID: "guild-1", MessageID: "message-1", React: "emoji-1", RoleID: "role-1"}
			discord := fakediscord.NewSideEffects()
			if tc.missing {
				discord.MissingRoles["guild-1/role-1"] = true
			}
			module := NewModule(repo, discord, discord, discord, discord, discord)

			err := module.ReactionEventHandler(tc.remove)(context.Background(), events.Event{
				Type:      tc.eventType,
				GuildID:   "guild-1",
				MessageID: "message-1",
				UserID:    "user-1",
				Reaction:  &events.Reaction{EmojiID: "emoji-1"},
			})
			if err != nil {
				t.Fatalf("handler: %v", err)
			}
			if len(discord.DirectMessages) != 1 || len(discord.DirectMessages[0].Message.Embeds) != 1 {
				t.Fatalf("direct messages = %#v", discord.DirectMessages)
			}
			want := tc.wantPrefix + "我沒有權限給大家這個身分組或是身分組被刪除了(請把我的身分組調高)!"
			if got := discord.DirectMessages[0].Message.Embeds[0].Title; got != want {
				t.Fatalf("title = %q, want %q", got, want)
			}
		})
	}
}

func TestReactionEventOperationalFailuresAreLoggedWithoutMisleadingDM(t *testing.T) {
	wantErr := errors.New("role event dependency failed")
	tests := []struct {
		name  string
		setup func(*fakemongo.RoleSelectionRepository, *fakediscord.SideEffects) ports.DiscordRoleInspector
	}{
		{
			name: "repository failure",
			setup: func(repo *fakemongo.RoleSelectionRepository, _ *fakediscord.SideEffects) ports.DiscordRoleInspector {
				repo.Err = wantErr
				return nil
			},
		},
		{
			name: "discord role API failure",
			setup: func(_ *fakemongo.RoleSelectionRepository, discord *fakediscord.SideEffects) ports.DiscordRoleInspector {
				discord.Err = wantErr
				return nil
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := fakemongo.NewRoleSelectionRepository()
			repo.Reactions["guild-1/message-1/emoji-1"] = domain.RoleReactionConfig{GuildID: "guild-1", MessageID: "message-1", React: "emoji-1", RoleID: "role-1"}
			discord := fakediscord.NewSideEffects()
			inspector := test.setup(repo, discord)
			module := NewModule(repo, discord, inspector, discord, discord, discord)

			err := module.ReactionEventHandler(false)(context.Background(), events.Event{
				Type:      events.TypeReactionAdd,
				GuildID:   "guild-1",
				MessageID: "message-1",
				UserID:    "user-1",
				Reaction:  &events.Reaction{EmojiID: "emoji-1"},
			})
			if !errors.Is(err, wantErr) {
				t.Fatalf("handler error = %v", err)
			}
			if len(discord.DirectMessages) != 0 {
				t.Fatalf("direct messages = %#v", discord.DirectMessages)
			}
		})
	}
}

func TestButtonApplyAlreadyAssignedUsesLegacyError(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	repo.Buttons["guild-1/button-add"] = domain.RoleButtonConfig{GuildID: "guild-1", Number: "button-add", RoleID: "role-1"}
	discord := fakediscord.NewSideEffects()
	discord.AssignableRoles["guild-1/role-1"] = true
	module := NewModule(repo, discord, discord, discord, discord, discord)
	interaction := fakediscord.ComponentInteractionFromID("button-add")
	interaction.Actor.RoleIDs = []string{"role-1"}
	responder := fakediscord.NewResponder()

	if err := module.ButtonApplyHandler(false)(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	want := roleSelectionErrorPrefix + "你已經擁有身分組了!"
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || responder.Edits[0].Embeds[0].Title != want {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestRoleSelectionButtonErrorsMatchLegacy(t *testing.T) {
	unknown := errors.New("button dependency failed")
	tests := []struct {
		name        string
		err         error
		remove      bool
		wantTitle   string
		wantContent string
	}{
		{name: "missing config reliability response", err: ports.ErrRoleButtonConfigMissing, wantContent: "很抱歉，出現了錯誤!"},
		{name: "already assigned", err: coreservice.ErrRoleAlreadyAssigned, wantTitle: roleSelectionErrorPrefix + "你已經擁有身分組了!"},
		{name: "not assigned", err: coreservice.ErrRoleNotAssigned, remove: true, wantTitle: roleSelectionErrorPrefix + " 你沒有這個身分組!"},
		{name: "missing role add", err: ports.ErrDiscordRoleMissing, wantTitle: roleSelectionActionPrefix + "請通知群主管裡員找不到這個身分組!"},
		{name: "missing role remove", err: ports.ErrDiscordRoleMissing, remove: true, wantTitle: roleSelectionActionPrefix + "找不到這個身分組!"},
		{name: "hierarchy add", err: ports.ErrDiscordRoleNotAssignable, wantTitle: roleSelectionActionPrefix + "請通知群主管裡員我沒有權限給你這個身分組(請把我的身分組調高)!"},
		{name: "hierarchy remove", err: ports.ErrDiscordRoleNotAssignable, remove: true, wantTitle: roleSelectionActionPrefix + "請通知群主管裡員我沒有權限給你這個身分組(請把我的身分組調高)!"},
		{name: "operational fallback", err: unknown, wantContent: "opps,出現了錯誤!\n有可能是你設定沒設定好\n或是我沒有權限喔(請確認我的權限比你要加的權限高，還需要管理身分組的權限)"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			message := roleSelectionButtonError(test.err, test.remove)
			if message.Content != test.wantContent {
				t.Fatalf("content = %q, want %q", message.Content, test.wantContent)
			}
			if test.wantTitle == "" {
				if len(message.Embeds) != 0 {
					t.Fatalf("embeds = %#v", message.Embeds)
				}
			} else if len(message.Embeds) != 1 || message.Embeds[0].Title != test.wantTitle || message.Embeds[0].Color != roleSelectionErrorColor {
				t.Fatalf("embeds = %#v", message.Embeds)
			}
			if message.AllowedMentions == nil {
				t.Fatal("allowed mentions must be explicit")
			}
		})
	}
}

func seedRoleReactionChannel(discord *fakediscord.SideEffects, guildID string, channelID string) {
	discord.Channels = append(discord.Channels, ports.ChannelRef{GuildID: guildID, ChannelID: channelID})
}

func seedRoleReactionMessage(discord *fakediscord.SideEffects, guildID string, channelID string, messageID string) {
	seedRoleReactionChannel(discord, guildID, channelID)
	discord.Messages[channelID+"/"+messageID] = ports.MessageRef{ChannelID: channelID, MessageID: messageID}
}
