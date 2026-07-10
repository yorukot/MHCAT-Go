package roles

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/roles"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestRoleSelectionDefinitionsMatchLegacyVisibleContract(t *testing.T) {
	want := []commands.Definition{
		{
			Type:        commands.CommandTypeChatInput,
			Name:        RoleButtonCommandName,
			Description: "設定領取身分組的消息(點按鈕自動增加身分組)",
			Ownership:   commands.ManagedOwnership("role-selection", commands.ScopeGuild),
			Options: []commands.Option{{
				Type:        commands.OptionTypeRole,
				Name:        "身分組",
				Description: "輸入身分組!",
				Required:    true,
			}},
		},
		{
			Type:        commands.CommandTypeChatInput,
			Name:        RoleReactionSetCommandName,
			Description: "設定領取身分組的消息-點按鈕自動增加身分組(如要更改某個表情符號所給予的身分組，請一樣打這個指令)",
			Ownership:   commands.ManagedOwnership("role-selection", commands.ScopeGuild),
			Options: []commands.Option{
				{Type: commands.OptionTypeString, Name: "訊息url", Description: "輸入訊息的url(對訊息點右鍵按複製訊息連結)!", Required: true},
				{Type: commands.OptionTypeRole, Name: "身分組", Description: "輸入要給的身分組!", Required: true},
				{Type: commands.OptionTypeString, Name: "表情符號", Description: "請輸入要放在訊息下面的表情符號", Required: true},
			},
		},
		{
			Type:        commands.CommandTypeChatInput,
			Name:        RoleReactionDeleteCommandName,
			Description: "選取身分組刪除-表情符號版(進行刪除)",
			Ownership:   commands.ManagedOwnership("role-selection", commands.ScopeGuild),
			Options: []commands.Option{
				{Type: commands.OptionTypeString, Name: "訊息url", Description: "輸入訊息的url(對訊息點右鍵按複製訊息連結)!", Required: true},
				{Type: commands.OptionTypeString, Name: "表情符號", Description: "請輸入要放在訊息下面的表情符號", Required: true},
			},
		},
	}
	if got := Definitions(); !reflect.DeepEqual(got, want) {
		t.Fatalf("definitions = %#v, want %#v", got, want)
	}
}

func TestRoleSelectionMessagesMatchLegacyVisibleContract(t *testing.T) {
	for _, test := range []struct {
		name string
		got  responses.Message
		want responses.Message
	}{
		{
			name: "permission error",
			got:  roleSelectionErrorMessage("你需要有`訊息管理`才能使用此指令"),
			want: responses.Message{
				Embeds:          []responses.Embed{{Title: roleSelectionErrorPrefix + "你需要有`訊息管理`才能使用此指令", Color: roleSelectionErrorColor}},
				AllowedMentions: &responses.AllowedMentions{},
			},
		},
		{
			name: "reaction setup success",
			got:  roleSelectionSuccessTitle(roleSelectionDoneEmoji + " | 表情符號選取身分組成功設定"),
			want: responses.Message{
				Embeds:          []responses.Embed{{Title: roleSelectionDoneEmoji + " | 表情符號選取身分組成功設定", Color: roleSelectionSuccessColor}},
				AllowedMentions: &responses.AllowedMentions{},
			},
		},
		{
			name: "reaction delete success",
			got:  roleSelectionSuccessTitle("表情符號選取身分組成功刪除"),
			want: responses.Message{
				Embeds:          []responses.Embed{{Title: "表情符號選取身分組成功刪除", Color: roleSelectionSuccessColor}},
				AllowedMentions: &responses.AllowedMentions{},
			},
		},
		{
			name: "button add success",
			got:  roleSelectionSuccessTitle(roleSelectionDoneEmoji + " | 成功增加身分組!"),
			want: responses.Message{
				Embeds:          []responses.Embed{{Title: roleSelectionDoneEmoji + " | 成功增加身分組!", Color: roleSelectionSuccessColor}},
				AllowedMentions: &responses.AllowedMentions{},
			},
		},
		{
			name: "button remove success",
			got:  roleSelectionSuccessTitle(roleSelectionDoneEmoji + " | 成功刪除身分組!"),
			want: responses.Message{
				Embeds:          []responses.Embed{{Title: roleSelectionDoneEmoji + " | 成功刪除身分組!", Color: roleSelectionSuccessColor}},
				AllowedMentions: &responses.AllowedMentions{},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if !reflect.DeepEqual(test.got, test.want) {
				t.Fatalf("message = %#v, want %#v", test.got, test.want)
			}
		})
	}
}

func TestRoleSelectionButtonErrorsMatchLegacyVisibleContract(t *testing.T) {
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

func TestRoleSelectionModalPanelAndReactionErrorsMatchLegacyVisibleContract(t *testing.T) {
	baseID := "2026071101341234567890.1234567"
	wantModal := responses.Modal{
		CustomID: "nal",
		Title:    "領取身分系統!",
		Rows: []responses.ModalRow{{Inputs: []responses.TextInput{{
			CustomID: "roleaddcontent" + baseID,
			Label:    "請輸入身分訊息內文",
			Style:    responses.TextInputStyleParagraph,
		}}}},
	}
	if got := roleSelectionModal(baseID); !reflect.DeepEqual(got, wantModal) {
		t.Fatalf("modal = %#v, want %#v", got, wantModal)
	}

	wantPanel := ports.OutboundMessage{
		Embeds: []ports.OutboundEmbed{{Title: "選取身分組", Description: "點按鈕領身分", Color: 0x123456}},
		Components: []ports.OutboundComponentRow{{Components: []ports.OutboundComponent{
			{Type: "button", CustomID: baseID + "add", Label: "✅點我增加身分組!", Style: "primary"},
			{Type: "button", CustomID: baseID + "delete", Label: "❎點我刪除身分組!", Style: "danger"},
		}}},
		AllowedMentions: ports.AllowedMentions{},
	}
	if got := roleSelectionButtonPanelOutbound(baseID, "點按鈕領身分", 0x123456); !reflect.DeepEqual(got, wantPanel) {
		t.Fatalf("panel = %#v, want %#v", got, wantPanel)
	}

	wantAddError := ports.OutboundMessage{
		Embeds:          []ports.OutboundEmbed{{Title: roleSelectionActionPrefix + "我沒有權限給大家這個身分組或是身分組被刪除了(請把我的身分組調高)!", Color: roleSelectionErrorColor}},
		AllowedMentions: ports.AllowedMentions{},
	}
	if got := roleSelectionRoleErrorOutbound(false); !reflect.DeepEqual(got, wantAddError) {
		t.Fatalf("reaction add error = %#v, want %#v", got, wantAddError)
	}
	wantRemoveError := ports.OutboundMessage{
		Embeds:          []ports.OutboundEmbed{{Title: roleSelectionErrorPrefix + "我沒有權限給大家這個身分組或是身分組被刪除了(請把我的身分組調高)!", Color: roleSelectionErrorColor}},
		AllowedMentions: ports.AllowedMentions{},
	}
	if got := roleSelectionRoleErrorOutbound(true); !reflect.DeepEqual(got, wantRemoveError) {
		t.Fatalf("reaction remove error = %#v, want %#v", got, wantRemoveError)
	}
}

func TestRoleSelectionResponseVisibilityMatchesLegacyContract(t *testing.T) {
	discord := fakediscord.NewSideEffects()
	module := NewModule(fakemongo.NewRoleSelectionRepository(), discord, discord, discord, discord, discord)

	for _, test := range []struct {
		name    string
		handler interactions.Handler
	}{
		{name: "reaction setup", handler: module.ReactionSetHandler()},
		{name: "reaction delete", handler: module.ReactionDeleteHandler()},
	} {
		t.Run(test.name, func(t *testing.T) {
			responder := fakediscord.NewResponder()
			if err := test.handler(context.Background(), fakediscord.SlashInteraction("denied"), responder); err != nil {
				t.Fatalf("handler: %v", err)
			}
			if len(responder.Defers) != 1 || responder.Defers[0].Ephemeral || len(responder.Edits) != 1 || len(responder.Replies) != 0 {
				t.Fatalf("response = defers %#v, edits %#v, replies %#v", responder.Defers, responder.Edits, responder.Replies)
			}
		})
	}

	responder := fakediscord.NewResponder()
	if err := module.ButtonSetupHandler()(context.Background(), fakediscord.SlashInteraction(RoleButtonCommandName), responder); err != nil {
		t.Fatalf("button setup: %v", err)
	}
	if len(responder.Replies) != 1 || !responder.Replies[0].Ephemeral || len(responder.Defers) != 0 {
		t.Fatalf("button setup response = replies %#v, defers %#v", responder.Replies, responder.Defers)
	}

	responder = fakediscord.NewResponder()
	if err := module.ButtonApplyHandler(false)(context.Background(), fakediscord.ComponentInteractionFromID("2026071101341234567890.1234567add"), responder); err != nil {
		t.Fatalf("button apply: %v", err)
	}
	if len(responder.Defers) != 1 || !responder.Defers[0].Ephemeral || len(responder.Edits) != 1 || responder.Edits[0].Content != "很抱歉，出現了錯誤!" {
		t.Fatalf("button apply response = defers %#v, edits %#v", responder.Defers, responder.Edits)
	}
}

func TestRoleSelectionLegacyRouterWorkflow(t *testing.T) {
	const baseID = "2026071101341234567890.1234567"
	repo := fakemongo.NewRoleSelectionRepository()
	discord := fakediscord.NewSideEffects()
	discord.AssignableRoles["guild-1/role-1"] = true
	module := NewModuleWithIDGenerator(repo, discord, discord, discord, discord, discord, func() string { return baseID })
	router := interactions.NewRouter()
	router.SetCustomIDParser(interactions.DefaultCustomIDParser{})
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register routes: %v", err)
	}

	setup := fakediscord.SlashInteractionWithOptions(RoleButtonCommandName, "", map[string]string{"身分組": "role-1"})
	setup.Actor.PermissionBits = permissionManageMessages
	responder := fakediscord.NewResponder()
	if err := router.Handle(context.Background(), setup, responder); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if !reflect.DeepEqual(responder.Modals, []responses.Modal{roleSelectionModal(baseID)}) {
		t.Fatalf("setup modals = %#v", responder.Modals)
	}

	modal := fakediscord.ModalInteraction(interactions.ModalKey{})
	modal.CustomID = roleSelectionModalID
	modal.ChannelID = "channel-1"
	modal.BotDisplayColor = 0x123456
	modal.ModalFields = []customid.ModalField{{CustomID: roleSelectionFieldPrefix + baseID, Value: "點按鈕領身分"}}
	responder = fakediscord.NewResponder()
	if err := router.Handle(context.Background(), modal, responder); err != nil {
		t.Fatalf("modal: %v", err)
	}
	if len(discord.Sent) != 1 || !reflect.DeepEqual(discord.Sent[0].Message, roleSelectionButtonPanelOutbound(baseID, "點按鈕領身分", 0x123456)) {
		t.Fatalf("sent panels = %#v", discord.Sent)
	}
	if len(responder.Defers) != 1 || responder.Defers[0].Ephemeral || len(responder.Edits) != 1 {
		t.Fatalf("modal response = defers %#v, edits %#v", responder.Defers, responder.Edits)
	}

	add := fakediscord.ComponentInteractionFromID(baseID + "add")
	responder = fakediscord.NewResponder()
	if err := router.Handle(context.Background(), add, responder); err != nil {
		t.Fatalf("add: %v", err)
	}
	if len(discord.AddedRoles) != 1 || len(responder.Defers) != 1 || !responder.Defers[0].Ephemeral || !reflect.DeepEqual(responder.Edits, []responses.Message{roleSelectionSuccessTitle(roleSelectionDoneEmoji + " | 成功增加身分組!")}) {
		t.Fatalf("add result = roles %#v, defers %#v, edits %#v", discord.AddedRoles, responder.Defers, responder.Edits)
	}

	remove := fakediscord.ComponentInteractionFromID(baseID + "delete")
	remove.Actor.RoleIDs = []string{"role-1"}
	responder = fakediscord.NewResponder()
	if err := router.Handle(context.Background(), remove, responder); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if len(discord.RemovedRoles) != 1 || len(responder.Defers) != 1 || !responder.Defers[0].Ephemeral || !reflect.DeepEqual(responder.Edits, []responses.Message{roleSelectionSuccessTitle(roleSelectionDoneEmoji + " | 成功刪除身分組!")}) {
		t.Fatalf("remove result = roles %#v, defers %#v, edits %#v", discord.RemovedRoles, responder.Defers, responder.Edits)
	}
}
