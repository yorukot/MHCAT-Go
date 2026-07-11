package autochat

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestSetHandlerRequiresManageMessages(t *testing.T) {
	module := NewModule(fakemongo.NewAutoChatConfigRepository(), nil)
	responder := fakediscord.NewResponder()
	interaction := autoChatSetSlash()
	interaction.Actor.PermissionBits = 0

	if err := module.SetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if !reflect.DeepEqual(responder.Defers, []responses.DeferOptions{{}}) || len(responder.Edits) != 1 || !reflect.DeepEqual(responder.Edits[0], autoChatErrorMessage("你需要有`訊息管理`才能使用此指令")) {
		t.Fatalf("defers=%#v edits=%#v", responder.Defers, responder.Edits)
	}
}

func TestSetHandlerSavesAndRendersLegacySuccess(t *testing.T) {
	repo := fakemongo.NewAutoChatConfigRepository()
	usage := &fakeusage.Tracker{}
	module := NewModule(repo, usage)
	responder := fakediscord.NewResponder()

	if err := module.SetHandler()(context.Background(), autoChatSetSlash(), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	saved, ok := repo.Last()
	if !ok || saved.GuildID != "guild-1" || saved.ChannelID != "channel-1" {
		t.Fatalf("saved = %#v ok=%v", saved, ok)
	}
	if !reflect.DeepEqual(responder.Defers, []responses.DeferOptions{{}}) {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(responder.Edits) != 1 || !reflect.DeepEqual(responder.Edits[0], autoChatSetSuccessMessage("channel-1")) {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	if len(usage.Events) != 1 || usage.Events[0].Feature != "autochat-config" || usage.Events[0].CommandName != AutoChatSetCommandName {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestDeleteHandlerDeletesAndRendersLegacySuccess(t *testing.T) {
	repo := fakemongo.NewAutoChatConfigRepository()
	repo.Configs["guild-1"] = structConfig("guild-1", "channel-1")
	module := NewModule(repo, nil)
	responder := fakediscord.NewResponder()

	interaction := fakediscord.SlashInteraction(AutoChatDeleteCommandName)
	interaction.Actor.PermissionBits = permissionManageMessages
	if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if _, ok := repo.Configs["guild-1"]; ok {
		t.Fatalf("config was not deleted: %#v", repo.Configs)
	}
	if !reflect.DeepEqual(responder.Defers, []responses.DeferOptions{{}}) || len(responder.Edits) != 1 || !reflect.DeepEqual(responder.Edits[0], autoChatDeleteSuccessMessage()) {
		t.Fatalf("defers=%#v edits=%#v", responder.Defers, responder.Edits)
	}
}

func TestDeleteHandlerMissingConfigUsesLegacyError(t *testing.T) {
	module := NewModule(fakemongo.NewAutoChatConfigRepository(), nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction(AutoChatDeleteCommandName)
	interaction.Actor.PermissionBits = permissionManageMessages

	if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if !reflect.DeepEqual(responder.Defers, []responses.DeferOptions{{}}) || len(responder.Edits) != 1 || !reflect.DeepEqual(responder.Edits[0], autoChatErrorMessage("你沒有設定過，我不知道要刪除甚麼!")) {
		t.Fatalf("defers=%#v edits=%#v", responder.Defers, responder.Edits)
	}
}

func TestConfigHandlersReturnControlledBackendErrors(t *testing.T) {
	for _, commandName := range []string{AutoChatSetCommandName, AutoChatDeleteCommandName} {
		t.Run(commandName, func(t *testing.T) {
			repo := fakemongo.NewAutoChatConfigRepository()
			repo.Err = errors.New("mongo credential secret")
			module := NewModule(repo, nil)
			interaction := fakediscord.SlashInteraction(commandName)
			interaction.Actor.PermissionBits = permissionManageMessages
			handler := module.DeleteHandler()
			if commandName == AutoChatSetCommandName {
				interaction = autoChatSetSlash()
				handler = module.SetHandler()
			}
			responder := fakediscord.NewResponder()
			if err := handler(context.Background(), interaction, responder); err != nil {
				t.Fatalf("handler: %v", err)
			}
			want := autoChatErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
			if !reflect.DeepEqual(responder.Defers, []responses.DeferOptions{{}}) || len(responder.Edits) != 1 || !reflect.DeepEqual(responder.Edits[0], want) || strings.Contains(responder.Edits[0].Embeds[0].Title, "credential") {
				t.Fatalf("defers=%#v edits=%#v", responder.Defers, responder.Edits)
			}
		})
	}
}

func autoChatSetSlash() interactions.Interaction {
	interaction := fakediscord.SlashInteractionWithOptions(AutoChatSetCommandName, "", map[string]string{
		optionChannel: "channel-1",
	})
	interaction.Actor.PermissionBits = permissionManageMessages
	return interaction
}

func structConfig(guildID string, channelID string) domain.AutoChatConfig {
	return domain.AutoChatConfig{GuildID: guildID, ChannelID: channelID}
}
