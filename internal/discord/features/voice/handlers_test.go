package voice

import (
	"context"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestSetHandlerSavesVoiceRoomConfig(t *testing.T) {
	repo := fakemongo.NewVoiceRoomConfigRepository()
	module := NewModule(repo, nil)
	interaction := voiceSetInteraction()
	responder := fakediscord.NewResponder()
	if err := module.SetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("set handler: %v", err)
	}
	saved, ok := repo.Last()
	if !ok {
		t.Fatal("expected saved config")
	}
	if saved.GuildID != "guild-1" || saved.TriggerChannelID != "voice-1" || saved.ParentID != "category-1" || saved.Name != "{name} 的包廂" || saved.Limit != 12 || !saved.Lock {
		t.Fatalf("saved config = %#v", saved)
	}
	assertEmbed(t, responder, legacyDoneEmoji+" | 成功進行設定", legacyVoiceEmoji+" 你成功對語音包廂進行`設定`")
}

func TestSetHandlerRejectsInvalidLimit(t *testing.T) {
	repo := fakemongo.NewVoiceRoomConfigRepository()
	module := NewModule(repo, nil)
	interaction := voiceSetInteraction()
	interaction.CommandOptions[optionUserLimit] = interactions.CommandOptionValue{
		Type: interactions.CommandOptionInteger,
		Int:  100,
	}
	responder := fakediscord.NewResponder()
	if err := module.SetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("set handler: %v", err)
	}
	if _, ok := repo.Last(); ok {
		t.Fatal("invalid limit should not save")
	}
	assertEmbedContains(t, responder, "必須為1-99的整數!")
}

func TestSetHandlerRequiresManageMessages(t *testing.T) {
	module := NewModule(fakemongo.NewVoiceRoomConfigRepository(), nil)
	interaction := voiceSetInteraction()
	interaction.Actor.PermissionBits = 0
	responder := fakediscord.NewResponder()
	if err := module.SetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("set handler: %v", err)
	}
	assertEmbedContains(t, responder, "你需要有`訊息管理`才能使用此指令")
}

func TestDeleteHandlerDeletesTriggerConfig(t *testing.T) {
	repo := fakemongo.NewVoiceRoomConfigRepository()
	if err := repo.SaveVoiceRoomConfig(context.Background(), domain.VoiceRoomConfig{
		GuildID:          "guild-1",
		TriggerChannelID: "voice-1",
		Name:             "{name}",
	}); err != nil {
		t.Fatalf("seed repo: %v", err)
	}
	module := NewModule(repo, nil)
	interaction := fakediscord.SlashInteraction(VoiceRoomDeleteCommandName)
	interaction.Actor.PermissionBits = permissionManageMessages
	interaction.CommandOptions = map[string]interactions.CommandOptionValue{
		optionChannelOrGroup: {
			Type:        interactions.CommandOptionChannel,
			String:      "voice-1",
			ChannelType: discordChannelTypeVoice,
		},
	}
	responder := fakediscord.NewResponder()
	if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("delete handler: %v", err)
	}
	if _, ok := repo.Configs["guild-1\x00voice-1"]; ok {
		t.Fatalf("config not deleted: %#v", repo.Configs)
	}
	assertEmbed(t, responder, legacyDoneEmoji+"成功進行刪除", legacyDeleteEmoji+"你成功對這個設定刪除")
}

func TestDeleteHandlerDeletesParentConfigs(t *testing.T) {
	repo := fakemongo.NewVoiceRoomConfigRepository()
	for _, config := range []domain.VoiceRoomConfig{
		{GuildID: "guild-1", TriggerChannelID: "voice-1", ParentID: "category-1", Name: "{name}"},
		{GuildID: "guild-1", TriggerChannelID: "voice-2", ParentID: "category-1", Name: "{name}"},
	} {
		if err := repo.SaveVoiceRoomConfig(context.Background(), config); err != nil {
			t.Fatalf("seed repo: %v", err)
		}
	}
	module := NewModule(repo, nil)
	interaction := fakediscord.SlashInteraction(VoiceRoomDeleteCommandName)
	interaction.Actor.PermissionBits = permissionManageMessages
	interaction.CommandOptions = map[string]interactions.CommandOptionValue{
		optionChannelOrGroup: {
			Type:        interactions.CommandOptionChannel,
			String:      "category-1",
			ChannelType: 4,
		},
	}
	responder := fakediscord.NewResponder()
	if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("delete handler: %v", err)
	}
	if len(repo.Configs) != 0 {
		t.Fatalf("parent configs not deleted: %#v", repo.Configs)
	}
	assertEmbed(t, responder, "成功進行刪除", "你成功對這個設定刪除")
}

func TestDeleteHandlerMissingMessagesMatchLegacy(t *testing.T) {
	module := NewModule(fakemongo.NewVoiceRoomConfigRepository(), nil)
	for _, tc := range []struct {
		name        string
		channelType int
		want        string
	}{
		{name: "trigger", channelType: discordChannelTypeVoice, want: "你沒有對這個頻道做出設定過喔!"},
		{name: "parent", channelType: 4, want: "你沒有對這個類別沒有設定喔!"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			interaction := fakediscord.SlashInteraction(VoiceRoomDeleteCommandName)
			interaction.Actor.PermissionBits = permissionManageMessages
			interaction.CommandOptions = map[string]interactions.CommandOptionValue{
				optionChannelOrGroup: {
					Type:        interactions.CommandOptionChannel,
					String:      "missing",
					ChannelType: tc.channelType,
				},
			}
			responder := fakediscord.NewResponder()
			if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
				t.Fatalf("delete handler: %v", err)
			}
			assertEmbedContains(t, responder, tc.want)
		})
	}
}

func voiceSetInteraction() interactions.Interaction {
	interaction := fakediscord.SlashInteraction(VoiceRoomSetCommandName)
	interaction.Actor.PermissionBits = permissionManageMessages
	interaction.CommandOptions = map[string]interactions.CommandOptionValue{
		optionTriggerChannel: {
			Type:            interactions.CommandOptionChannel,
			String:          "voice-1",
			ChannelType:     discordChannelTypeVoice,
			ChannelParentID: "category-1",
		},
		optionRoomName: {
			Type:   interactions.CommandOptionString,
			String: "{name} 的包廂",
		},
		optionOwnerLock: {
			Type: interactions.CommandOptionBoolean,
			Bool: true,
		},
		optionUserLimit: {
			Type: interactions.CommandOptionInteger,
			Int:  12,
		},
	}
	return interaction
}

func assertEmbed(t *testing.T, responder *fakediscord.Responder, title string, description string) {
	t.Helper()
	if len(responder.Defers) != 1 {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != title || embed.Description != description {
		t.Fatalf("embed = %#v", embed)
	}
}

func assertEmbedContains(t *testing.T, responder *fakediscord.Responder, text string) {
	t.Helper()
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if !strings.Contains(embed.Title, text) && !strings.Contains(embed.Description, text) {
		t.Fatalf("embed = %#v, want text %q", embed, text)
	}
}
