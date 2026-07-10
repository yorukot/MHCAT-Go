package voice

import (
	"context"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
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

func TestLockHandlerSavesPassword(t *testing.T) {
	repo := fakemongo.NewVoiceRoomLockRepository()
	repo.Locks["guild-1\x00voice-1"] = domain.VoiceRoomLock{
		GuildID:       "guild-1",
		ChannelID:     "voice-1",
		OwnerID:       "user-1",
		TextChannelID: "old-text",
	}
	module := NewLockModule(repo, nil)
	interaction := voiceLockInteraction("secret")
	responder := fakediscord.NewResponder()
	if err := module.LockHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("lock handler: %v", err)
	}
	saved, ok := repo.Last()
	if !ok {
		t.Fatal("expected saved lock")
	}
	if saved.GuildID != "guild-1" || saved.ChannelID != "voice-1" || saved.OwnerID != "user-1" || saved.TextChannelID != "text-1" || saved.Password != "secret" {
		t.Fatalf("saved lock = %#v", saved)
	}
	if len(saved.AllowedUserIDs) != 0 {
		t.Fatalf("allowed users should reset, got %#v", saved.AllowedUserIDs)
	}
	assertLockEmbed(t, responder, legacyDoneEmoji+" | 成功進行設定", legacyVoiceEmoji+" 你成功對語音包廂密碼進行設定為:secret")
}

func TestLockHandlerRequiresVoiceChannel(t *testing.T) {
	module := NewLockModule(fakemongo.NewVoiceRoomLockRepository(), nil)
	interaction := voiceLockInteraction("secret")
	interaction.Actor.VoiceChannelID = ""
	responder := fakediscord.NewResponder()
	if err := module.LockHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("lock handler: %v", err)
	}
	assertLockEmbedContains(t, responder, "你不在一個語音包廂!")
}

func TestLockHandlerMissingLockMessageMatchesLegacy(t *testing.T) {
	module := NewLockModule(fakemongo.NewVoiceRoomLockRepository(), nil)
	responder := fakediscord.NewResponder()
	if err := module.LockHandler()(context.Background(), voiceLockInteraction("secret"), responder); err != nil {
		t.Fatalf("lock handler: %v", err)
	}
	assertLockEmbedContains(t, responder, "你不在語音包廂或該語音包廂不支援設密碼!")
}

func TestLockHandlerRejectsNonOwner(t *testing.T) {
	repo := fakemongo.NewVoiceRoomLockRepository()
	repo.Locks["guild-1\x00voice-1"] = domain.VoiceRoomLock{
		GuildID:       "guild-1",
		ChannelID:     "voice-1",
		OwnerID:       "other-user",
		TextChannelID: "text-1",
	}
	module := NewLockModule(repo, nil)
	responder := fakediscord.NewResponder()
	if err := module.LockHandler()(context.Background(), voiceLockInteraction("secret"), responder); err != nil {
		t.Fatalf("lock handler: %v", err)
	}
	assertLockEmbedContains(t, responder, "只有包廂房主可以設定!")
}

func TestLockHandlerEmptyPasswordStoresNullEquivalent(t *testing.T) {
	repo := fakemongo.NewVoiceRoomLockRepository()
	repo.Locks["guild-1\x00voice-1"] = domain.VoiceRoomLock{
		GuildID:       "guild-1",
		ChannelID:     "voice-1",
		Password:      "old",
		OwnerID:       "user-1",
		TextChannelID: "old-text",
	}
	module := NewLockModule(repo, nil)
	responder := fakediscord.NewResponder()
	if err := module.LockHandler()(context.Background(), voiceLockInteraction(""), responder); err != nil {
		t.Fatalf("lock handler: %v", err)
	}
	saved, ok := repo.Last()
	if !ok {
		t.Fatal("expected saved lock")
	}
	if saved.Password != "" {
		t.Fatalf("expected empty password, got %#v", saved)
	}
	assertLockEmbed(t, responder, legacyDoneEmoji+" | 成功進行設定", legacyVoiceEmoji+" 你成功對語音包廂密碼進行設定為:null")
}

func TestLockAnswerHandlerAllowsUserAndReturnsLegacySuccess(t *testing.T) {
	repo := fakemongo.NewVoiceRoomLockRepository()
	repo.Locks["guild-1\x00123456789012345678"] = domain.VoiceRoomLock{
		GuildID:       "guild-1",
		ChannelID:     "123456789012345678",
		Password:      "secret",
		OwnerID:       "owner-1",
		TextChannelID: "text-1",
	}
	module := NewLockModule(repo, nil)
	responder := fakediscord.NewResponder()
	if err := module.AnswerHandler()(context.Background(), voiceLockAnswerInteraction("secret"), responder); err != nil {
		t.Fatalf("answer handler: %v", err)
	}
	lock := repo.Locks["guild-1\x00123456789012345678"]
	if len(lock.AllowedUserIDs) != 1 || lock.AllowedUserIDs[0] != "user-1" {
		t.Fatalf("allowed users = %#v", lock.AllowedUserIDs)
	}
	assertLockAnswerMessage(t, responder, legacyUnlockEmoji+" | 您成功輸入正確密碼\n可以重新加入語音頻道囉!")
}

func TestLockAnswerHandlerWrongPasswordUsesLegacyError(t *testing.T) {
	repo := fakemongo.NewVoiceRoomLockRepository()
	repo.Locks["guild-1\x00123456789012345678"] = domain.VoiceRoomLock{
		GuildID:       "guild-1",
		ChannelID:     "123456789012345678",
		Password:      "secret",
		OwnerID:       "owner-1",
		TextChannelID: "text-1",
	}
	module := NewLockModule(repo, nil)
	responder := fakediscord.NewResponder()
	if err := module.AnswerHandler()(context.Background(), voiceLockAnswerInteraction("wrong"), responder); err != nil {
		t.Fatalf("answer handler: %v", err)
	}
	if got := repo.Locks["guild-1\x00123456789012345678"].AllowedUserIDs; len(got) != 0 {
		t.Fatalf("wrong password should not allow user: %#v", got)
	}
	assertLockAnswerMessage(t, responder, "<a:Discord_AnimatedNo:1015989839809757295> | 你的密碼輸入錯誤!請重新加入語音頻道後在試一次!")
}

func TestLockModuleRoutesLegacyAnswerModal(t *testing.T) {
	repo := fakemongo.NewVoiceRoomLockRepository()
	repo.Locks["guild-1\x00123456789012345678"] = domain.VoiceRoomLock{
		GuildID:       "guild-1",
		ChannelID:     "123456789012345678",
		Password:      "secret",
		OwnerID:       "owner-1",
		TextChannelID: "text-1",
	}
	router := interactions.NewRouter()
	router.SetCustomIDParser(interactions.DefaultCustomIDParser{})
	module := NewLockModule(repo, nil)
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	responder := fakediscord.NewResponder()
	if err := router.Handle(context.Background(), voiceLockAnswerInteraction("secret"), responder); err != nil {
		t.Fatalf("handle answer modal: %v", err)
	}
	assertLockAnswerMessage(t, responder, legacyUnlockEmoji+" | 您成功輸入正確密碼\n可以重新加入語音頻道囉!")
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

func voiceLockInteraction(password string) interactions.Interaction {
	interaction := fakediscord.SlashInteraction(VoiceRoomLockCommandName)
	interaction.ChannelID = "text-1"
	interaction.Actor.VoiceChannelID = "voice-1"
	interaction.CommandOptions = map[string]interactions.CommandOptionValue{}
	if password != "" {
		interaction.CommandOptions[optionLockPassword] = interactions.CommandOptionValue{
			Type:   interactions.CommandOptionString,
			String: password,
		}
	}
	return interaction
}

func voiceLockAnswerInteraction(password string) interactions.Interaction {
	return interactions.Interaction{
		Type:     interactions.TypeModal,
		CustomID: "123456789012345678anser",
		Actor:    interactions.Actor{UserID: "user-1", Username: "User", UserTag: "User#0001", GuildID: "guild-1"},
		ModalFields: []customid.ModalField{{
			CustomID: voiceLockAnswerInputID,
			Value:    password,
		}},
	}
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

func assertLockAnswerMessage(t *testing.T, responder *fakediscord.Responder, title string) {
	t.Helper()
	if len(responder.Defers) != 1 {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	if got := responder.Edits[0].Embeds[0].Title; got != title {
		t.Fatalf("title = %q, want %q", got, title)
	}
	rows := responder.Edits[0].Components
	if len(rows) != 1 || len(rows[0].Components) != 1 {
		t.Fatalf("components = %#v", rows)
	}
	button := rows[0].Components[0]
	if button.Type != responses.ComponentTypeButton || button.Style != responses.ButtonStyleLink || button.URL != "https://discord.com/channels/guild-1/123456789012345678" {
		t.Fatalf("button = %#v", button)
	}
}

func assertLockEmbed(t *testing.T, responder *fakediscord.Responder, title string, description string) {
	t.Helper()
	assertEmbed(t, responder, title, description)
	if !responder.Defers[0].Ephemeral {
		t.Fatalf("lock response should defer ephemerally: %#v", responder.Defers)
	}
	if !responder.Edits[0].Ephemeral {
		t.Fatalf("lock edit should be marked ephemeral: %#v", responder.Edits[0])
	}
}

func assertLockEmbedContains(t *testing.T, responder *fakediscord.Responder, text string) {
	t.Helper()
	assertEmbedContains(t, responder, text)
	if len(responder.Defers) != 1 || !responder.Defers[0].Ephemeral {
		t.Fatalf("lock response should defer ephemerally: %#v", responder.Defers)
	}
	if len(responder.Edits) != 1 || !responder.Edits[0].Ephemeral {
		t.Fatalf("lock edit should be marked ephemeral: %#v", responder.Edits)
	}
}
