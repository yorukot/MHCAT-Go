package voice

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
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

func TestSetHandlerPreservesLegacyRoomNameText(t *testing.T) {
	repo := fakemongo.NewVoiceRoomConfigRepository()
	module := NewModule(repo, nil)
	interaction := voiceSetInteraction()
	interaction.CommandOptions[optionRoomName] = interactions.CommandOptionValue{
		Type:   interactions.CommandOptionString,
		String: "  {name}-{name}  ",
	}
	responder := fakediscord.NewResponder()
	if err := module.SetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("set handler: %v", err)
	}
	saved, ok := repo.Last()
	if !ok || saved.Name != "  {name}-{name}  " {
		t.Fatalf("saved config = %#v, ok = %t", saved, ok)
	}
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

func TestSetHandlerAcceptsExplicitZeroLimitLikeLegacy(t *testing.T) {
	repo := fakemongo.NewVoiceRoomConfigRepository()
	module := NewModule(repo, nil)
	interaction := voiceSetInteraction()
	interaction.CommandOptions[optionUserLimit] = interactions.CommandOptionValue{
		Type: interactions.CommandOptionInteger,
		Int:  0,
	}
	responder := fakediscord.NewResponder()
	if err := module.SetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("set handler: %v", err)
	}
	saved, ok := repo.Last()
	if !ok || saved.Limit != 0 {
		t.Fatalf("saved config = %#v, ok = %t", saved, ok)
	}
	assertEmbed(t, responder, legacyDoneEmoji+" | 成功進行設定", legacyVoiceEmoji+" 你成功對語音包廂進行`設定`")
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
	interaction := voiceLockInteraction(" secret ")
	responder := fakediscord.NewResponder()
	if err := module.LockHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("lock handler: %v", err)
	}
	saved, ok := repo.Last()
	if !ok {
		t.Fatal("expected saved lock")
	}
	if saved.GuildID != "guild-1" || saved.ChannelID != "voice-1" || saved.OwnerID != "user-1" || saved.TextChannelID != "text-1" || saved.Password != " secret " || !saved.PasswordPresent {
		t.Fatalf("saved lock = %#v", saved)
	}
	if len(saved.AllowedUserIDs) != 0 {
		t.Fatalf("allowed users should reset, got %#v", saved.AllowedUserIDs)
	}
	assertLockEmbed(t, responder, legacyDoneEmoji+" | 成功進行設定", legacyVoiceEmoji+" 你成功對語音包廂密碼進行設定為: secret ")
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
	if saved.Password != "" || saved.PasswordPresent {
		t.Fatalf("expected absent password, got %#v", saved)
	}
	assertLockEmbed(t, responder, legacyDoneEmoji+" | 成功進行設定", legacyVoiceEmoji+" 你成功對語音包廂密碼進行設定為:null")
}

func TestLockAnswerHandlerAllowsUserAndReturnsLegacySuccess(t *testing.T) {
	repo := fakemongo.NewVoiceRoomLockRepository()
	repo.Locks["guild-1\x00123456789012345678"] = domain.VoiceRoomLock{
		GuildID:         "guild-1",
		ChannelID:       "123456789012345678",
		Password:        " secret ",
		PasswordPresent: true,
		OwnerID:         "owner-1",
		TextChannelID:   "text-1",
	}
	module := NewLockModule(repo, nil)
	responder := fakediscord.NewResponder()
	if err := module.AnswerHandler()(context.Background(), voiceLockAnswerInteraction(" secret "), responder); err != nil {
		t.Fatalf("answer handler: %v", err)
	}
	lock := repo.Locks["guild-1\x00123456789012345678"]
	if len(lock.AllowedUserIDs) != 1 || lock.AllowedUserIDs[0] != "user-1" {
		t.Fatalf("allowed users = %#v", lock.AllowedUserIDs)
	}
	assertLockAnswerMessage(t, responder, legacyUnlockEmoji+" | 您成功輸入正確密碼\n可以重新加入語音頻道囉!")
	if responder.Edits[0].Embeds[0].Color != legacyVoiceLockColor {
		t.Fatalf("success color = %#x", responder.Edits[0].Embeds[0].Color)
	}
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
	if responder.Edits[0].Embeds[0].Color != legacyVoiceLockErrorColor {
		t.Fatalf("error color = %#x", responder.Edits[0].Embeds[0].Color)
	}
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

func TestLockPromptHandlerShowsLegacyAnswerModal(t *testing.T) {
	now := time.UnixMilli(1_700_000_000_000)
	module := NewLockModuleWithClock(fakemongo.NewVoiceRoomLockRepository(), nil, voiceFixedClock{now: now})
	interaction := fakediscord.ComponentInteractionFromID(voiceLockPromptButtonID("123456789012345678", "user-1", now.Add(voiceLockPromptTTL)))
	responder := fakediscord.NewResponder()
	if err := module.PromptHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("prompt handler: %v", err)
	}
	if len(responder.Modals) != 1 {
		t.Fatalf("modals = %#v", responder.Modals)
	}
	modal := responder.Modals[0]
	if modal.CustomID != "123456789012345678anser" || modal.Title != "請輸入密碼!" {
		t.Fatalf("modal = %#v", modal)
	}
	if len(modal.Rows) != 1 || len(modal.Rows[0].Inputs) != 1 || modal.Rows[0].Inputs[0].CustomID != voiceLockAnswerInputID || modal.Rows[0].Inputs[0].Label != "請輸入包廂密碼!" {
		t.Fatalf("modal inputs = %#v", modal.Rows)
	}
}

func TestLockModuleRoutesVersionedPromptButton(t *testing.T) {
	router := interactions.NewRouter()
	router.SetCustomIDParser(interactions.DefaultCustomIDParser{})
	now := time.UnixMilli(1_700_000_000_000)
	module := NewLockModuleWithClock(fakemongo.NewVoiceRoomLockRepository(), nil, voiceFixedClock{now: now})
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	responder := fakediscord.NewResponder()
	interaction := fakediscord.ComponentInteractionFromID(voiceLockPromptButtonID("123456789012345678", "user-1", now.Add(voiceLockPromptTTL)))
	if err := router.Handle(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handle prompt: %v", err)
	}
	if len(responder.Modals) != 1 || responder.Modals[0].CustomID != "123456789012345678anser" {
		t.Fatalf("modals = %#v", responder.Modals)
	}
}

func TestLockPromptHandlerExpiresAtLegacyDeadline(t *testing.T) {
	deadline := time.UnixMilli(1_700_000_060_000)
	module := NewLockModuleWithClock(fakemongo.NewVoiceRoomLockRepository(), nil, voiceFixedClock{now: deadline})
	interaction := fakediscord.ComponentInteractionFromID(voiceLockPromptButtonID("123456789012345678", "user-1", deadline))
	responder := fakediscord.NewResponder()
	if err := module.PromptHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("prompt handler: %v", err)
	}
	if len(responder.Modals) != 0 || len(responder.Replies) != 1 || !responder.Replies[0].Ephemeral || !strings.Contains(responder.Replies[0].Embeds[0].Title, "請重新加入語音頻道後再試一次!") {
		t.Fatalf("modals=%#v replies=%#v", responder.Modals, responder.Replies)
	}
}

func TestVoiceLockPromptButtonIDFitsDiscordLimit(t *testing.T) {
	deadline := time.UnixMilli(1_700_000_060_000)
	customID := voiceLockPromptButtonID("1234567890123456789", "9876543210987654321", deadline)
	if customID == "lock_start" || len(customID) > customid.MaxCustomIDLength {
		t.Fatalf("custom id length=%d value=%q", len(customID), customID)
	}
	channelID, userID, expiresAt, ok := voiceLockPromptPayload(customID)
	if !ok || channelID != "1234567890123456789" || userID != "9876543210987654321" || !expiresAt.Equal(deadline) {
		t.Fatalf("channel=%q user=%q expires=%v ok=%t", channelID, userID, expiresAt, ok)
	}
}

func TestLockModuleRoutesLegacyPromptButtonToRetryError(t *testing.T) {
	router := interactions.NewRouter()
	router.SetCustomIDParser(interactions.DefaultCustomIDParser{})
	module := NewLockModule(fakemongo.NewVoiceRoomLockRepository(), nil)
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	responder := fakediscord.NewResponder()
	if err := router.Handle(context.Background(), fakediscord.ComponentInteractionFromID("lock_start"), responder); err != nil {
		t.Fatalf("handle legacy prompt: %v", err)
	}
	if len(responder.Replies) != 1 || len(responder.Replies[0].Embeds) != 1 || !strings.Contains(responder.Replies[0].Embeds[0].Title, "請重新加入語音頻道後再試一次!") || !responder.Replies[0].Ephemeral {
		t.Fatalf("reply = %#v", responder.Replies)
	}
}

func TestLockEventHandlerPromptsDisconnectsAndDMsLockedJoin(t *testing.T) {
	repo := fakemongo.NewVoiceRoomLockRepository()
	repo.Locks["guild-1\x00123456789012345678"] = domain.VoiceRoomLock{
		GuildID:       "guild-1",
		ChannelID:     "123456789012345678",
		Password:      "secret",
		OwnerID:       "owner-1",
		TextChannelID: "text-1",
	}
	sideEffects := fakediscord.NewSideEffects()
	now := time.UnixMilli(1_700_000_000_000)
	module := NewLockEventModuleWithClock(repo, sideEffects, sideEffects, sideEffects, voiceFixedClock{now: now})
	module.color = func() int { return 0x123456 }
	err := module.VoiceStateHandler()(context.Background(), events.Event{
		Type:    events.TypeVoiceState,
		GuildID: "guild-1",
		UserID:  "user-1",
		VoiceState: &events.VoiceState{
			GuildID:   "guild-1",
			UserID:    "user-1",
			ChannelID: "123456789012345678",
		},
	})
	if err != nil {
		t.Fatalf("voice state handler: %v", err)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].ChannelID != "text-1" {
		t.Fatalf("sent messages = %#v", sideEffects.Sent)
	}
	prompt := sideEffects.Sent[0].Message
	if prompt.Content != "<@user-1>" || len(prompt.Embeds) != 1 || prompt.Embeds[0].Color != legacyVoiceLockColor || !strings.Contains(prompt.Embeds[0].Description, "<#123456789012345678>") {
		t.Fatalf("prompt = %#v", prompt)
	}
	if len(prompt.Components) != 1 || len(prompt.Components[0].Components) != 1 {
		t.Fatalf("prompt components = %#v", prompt.Components)
	}
	button := prompt.Components[0].Components[0]
	if button.Label != "點我輸入密碼!" || button.Emoji != legacyArrowPinkEmoji || button.Style != "success" {
		t.Fatalf("prompt button = %#v", button)
	}
	channelID, userID, expiresAt, ok := voiceLockPromptPayload(button.CustomID)
	if !ok || channelID != "123456789012345678" || userID != "user-1" || !expiresAt.Equal(now.Add(voiceLockPromptTTL)) || len(button.CustomID) > customid.MaxCustomIDLength {
		t.Fatalf("prompt custom id parsed as channel=%q user=%q expires=%v ok=%t raw=%q", channelID, userID, expiresAt, ok, button.CustomID)
	}
	if len(sideEffects.MovedMembers) != 1 || sideEffects.MovedMembers[0].GuildID != "guild-1" || sideEffects.MovedMembers[0].UserID != "user-1" || sideEffects.MovedMembers[0].ChannelID != nil {
		t.Fatalf("moved members = %#v", sideEffects.MovedMembers)
	}
	if len(sideEffects.DirectMessages) != 1 || sideEffects.DirectMessages[0].UserID != "user-1" || sideEffects.DirectMessages[0].Message.Embeds[0].Color != 0x123456 || !strings.Contains(sideEffects.DirectMessages[0].Message.Embeds[0].Description, "<#text-1>") {
		t.Fatalf("direct messages = %#v", sideEffects.DirectMessages)
	}
}

func TestLockEventHandlerSkipsAllowedBotAndUnchangedVoiceState(t *testing.T) {
	repo := fakemongo.NewVoiceRoomLockRepository()
	repo.Locks["guild-1\x00123456789012345678"] = domain.VoiceRoomLock{
		GuildID:        "guild-1",
		ChannelID:      "123456789012345678",
		Password:       "secret",
		OwnerID:        "owner-1",
		TextChannelID:  "text-1",
		AllowedUserIDs: []string{"allowed-user"},
	}
	sideEffects := fakediscord.NewSideEffects()
	module := NewLockEventModule(repo, sideEffects, sideEffects, sideEffects)
	for _, event := range []events.Event{
		{Type: events.TypeVoiceState, GuildID: "guild-1", IsBot: true, VoiceState: &events.VoiceState{GuildID: "guild-1", UserID: "bot-1", ChannelID: "123456789012345678"}},
		{Type: events.TypeVoiceState, GuildID: "guild-1", VoiceState: &events.VoiceState{GuildID: "guild-1", UserID: "allowed-user", ChannelID: "123456789012345678"}},
		{Type: events.TypeVoiceState, GuildID: "guild-1", VoiceState: &events.VoiceState{GuildID: "guild-1", UserID: "user-1", ChannelID: "123456789012345678", BeforeChannel: "123456789012345678"}},
	} {
		if err := module.VoiceStateHandler()(context.Background(), event); err != nil {
			t.Fatalf("voice state handler: %v", err)
		}
	}
	if len(sideEffects.Sent) != 0 || len(sideEffects.MovedMembers) != 0 || len(sideEffects.DirectMessages) != 0 {
		t.Fatalf("side effects should be empty: sent=%#v moved=%#v dm=%#v", sideEffects.Sent, sideEffects.MovedMembers, sideEffects.DirectMessages)
	}
}

func TestRoomEventHandlerCreatesTracksMovesAndDMsLockableRoom(t *testing.T) {
	configs := fakemongo.NewVoiceRoomConfigRepository()
	configs.Configs["guild-1\x00trigger-1"] = domain.VoiceRoomConfig{
		GuildID:          "guild-1",
		TriggerChannelID: "trigger-1",
		ParentID:         "parent-1",
		Name:             "{name} 的包廂",
		Limit:            7,
		Lock:             true,
	}
	states := fakemongo.NewVoiceRoomStateRepository()
	locks := fakemongo.NewVoiceRoomLockRepository()
	sideEffects := fakediscord.NewSideEffects()
	sideEffects.Channels = append(sideEffects.Channels, portsChannel("guild-1", "parent-1", "", "parent", 4, []ports.PermissionOverwrite{
		{ID: "guild-1", Type: 0, Deny: 1024},
		{ID: "user-1", Type: permissionOverwriteMember, Deny: permissionManageChannels},
	}))
	module := NewRoomEventModule(configs, states, locks, sideEffects, sideEffects, sideEffects)
	module.color = func() int { return 0x654321 }
	err := module.VoiceStateHandler()(context.Background(), events.Event{
		Type:    events.TypeVoiceState,
		GuildID: "guild-1",
		Member:  &events.Member{UserID: "user-1", Username: "Yoru"},
		VoiceState: &events.VoiceState{
			GuildID:   "guild-1",
			UserID:    "user-1",
			ChannelID: "trigger-1",
		},
	})
	if err != nil {
		t.Fatalf("voice state handler: %v", err)
	}
	if len(sideEffects.Created) != 1 {
		t.Fatalf("created channels = %#v", sideEffects.Created)
	}
	createdReq := sideEffects.Created[0]
	if createdReq.GuildID != "guild-1" || createdReq.ParentID != "parent-1" || createdReq.Name != "Yoru 的包廂" || createdReq.Type != discordChannelTypeVoice || createdReq.UserLimit != 7 {
		t.Fatalf("created request = %#v", createdReq)
	}
	if !hasOverwrite(createdReq.PermissionOverwrites, "guild-1", 0, 0, 1024) || !hasOverwrite(createdReq.PermissionOverwrites, "user-1", permissionOverwriteMember, permissionManageChannels|permissionManageRoles, 0) {
		t.Fatalf("permission overwrites = %#v", createdReq.PermissionOverwrites)
	}
	ownerOverwrite, ok := findOverwrite(createdReq.PermissionOverwrites, "user-1", permissionOverwriteMember)
	if !ok || ownerOverwrite.Deny&(permissionManageChannels|permissionManageRoles) != 0 {
		t.Fatalf("owner overwrite = %#v ok=%t", ownerOverwrite, ok)
	}
	if _, ok := states.States["guild-1\x00created-channel-1"]; !ok || len(states.Saved) != 1 {
		t.Fatalf("states=%#v saved=%#v", states.States, states.Saved)
	}
	lock := locks.Locks["guild-1\x00created-channel-1"]
	if lock.OwnerID != "user-1" || lock.Password != "" || lock.TextChannelID != "" {
		t.Fatalf("seed lock = %#v", lock)
	}
	if len(sideEffects.MovedMembers) != 1 || sideEffects.MovedMembers[0].ChannelID == nil || *sideEffects.MovedMembers[0].ChannelID != "created-channel-1" {
		t.Fatalf("moved members = %#v", sideEffects.MovedMembers)
	}
	if len(sideEffects.DirectMessages) != 1 || sideEffects.DirectMessages[0].Message.Embeds[0].Color != 0x654321 || !strings.Contains(sideEffects.DirectMessages[0].Message.Embeds[0].Title, "你開啟了一個可上鎖的語音頻道") {
		t.Fatalf("direct messages = %#v", sideEffects.DirectMessages)
	}
}

func TestRoomEventHandlerCreatesDynamicRoomForBotLikeLegacy(t *testing.T) {
	configs := fakemongo.NewVoiceRoomConfigRepository()
	configs.Configs["guild-1\x00trigger-1"] = domain.VoiceRoomConfig{
		GuildID:          "guild-1",
		TriggerChannelID: "trigger-1",
		ParentID:         "parent-1",
		Name:             "{name} room",
	}
	states := fakemongo.NewVoiceRoomStateRepository()
	sideEffects := fakediscord.NewSideEffects()
	sideEffects.Channels = append(sideEffects.Channels, portsChannel("guild-1", "parent-1", "", "parent", 4, nil))
	module := NewRoomEventModule(configs, states, fakemongo.NewVoiceRoomLockRepository(), sideEffects, sideEffects, sideEffects)
	err := module.VoiceStateHandler()(context.Background(), events.Event{
		Type:    events.TypeVoiceState,
		GuildID: "guild-1",
		IsBot:   true,
		Member:  &events.Member{UserID: "bot-1", Username: "Helper", IsBot: true},
		VoiceState: &events.VoiceState{
			GuildID:   "guild-1",
			UserID:    "bot-1",
			ChannelID: "trigger-1",
		},
	})
	if err != nil {
		t.Fatalf("voice state handler: %v", err)
	}
	if len(sideEffects.Created) != 1 || sideEffects.Created[0].Name != "Helper room" {
		t.Fatalf("created channels = %#v", sideEffects.Created)
	}
	if len(sideEffects.MovedMembers) != 1 || sideEffects.MovedMembers[0].UserID != "bot-1" {
		t.Fatalf("moved members = %#v", sideEffects.MovedMembers)
	}
}

func TestRoomEventHandlerDeletesEmptyTrackedRoom(t *testing.T) {
	configs := fakemongo.NewVoiceRoomConfigRepository()
	states := fakemongo.NewVoiceRoomStateRepository()
	states.States["guild-1\x00dynamic-1"] = domain.VoiceRoomState{GuildID: "guild-1", ChannelID: "dynamic-1"}
	locks := fakemongo.NewVoiceRoomLockRepository()
	locks.Locks["guild-1\x00dynamic-1"] = domain.VoiceRoomLock{GuildID: "guild-1", ChannelID: "dynamic-1", OwnerID: "owner-1"}
	sideEffects := fakediscord.NewSideEffects()
	sideEffects.Channels = append(sideEffects.Channels, portsChannel("guild-1", "dynamic-1", "parent-1", "dynamic", discordChannelTypeVoice, nil))
	sideEffects.VoiceMembers["guild-1/dynamic-1"] = 0
	module := NewRoomEventModule(configs, states, locks, sideEffects, sideEffects, sideEffects)
	err := module.VoiceStateHandler()(context.Background(), events.Event{
		Type:    events.TypeVoiceState,
		GuildID: "guild-1",
		UserID:  "user-1",
		VoiceState: &events.VoiceState{
			GuildID:       "guild-1",
			UserID:        "user-1",
			BeforeChannel: "dynamic-1",
		},
	})
	if err != nil {
		t.Fatalf("voice state handler: %v", err)
	}
	if len(sideEffects.Deleted) != 1 || sideEffects.Deleted[0] != "dynamic-1" {
		t.Fatalf("deleted channels = %#v", sideEffects.Deleted)
	}
	_, stateExists := states.States["guild-1\x00dynamic-1"]
	_, lockExists := locks.Locks["guild-1\x00dynamic-1"]
	if stateExists || lockExists {
		t.Fatalf("states=%#v locks=%#v", states.States, locks.Locks)
	}
}

func TestRoomEventHandlerKeepsOccupiedTrackedRoom(t *testing.T) {
	states := fakemongo.NewVoiceRoomStateRepository()
	states.States["guild-1\x00dynamic-1"] = domain.VoiceRoomState{GuildID: "guild-1", ChannelID: "dynamic-1"}
	locks := fakemongo.NewVoiceRoomLockRepository()
	sideEffects := fakediscord.NewSideEffects()
	sideEffects.VoiceMembers["guild-1/dynamic-1"] = 1
	module := NewRoomEventModule(fakemongo.NewVoiceRoomConfigRepository(), states, locks, sideEffects, sideEffects, sideEffects)
	err := module.VoiceStateHandler()(context.Background(), events.Event{
		Type:    events.TypeVoiceState,
		GuildID: "guild-1",
		UserID:  "user-1",
		VoiceState: &events.VoiceState{
			GuildID:       "guild-1",
			UserID:        "user-1",
			BeforeChannel: "dynamic-1",
		},
	})
	if err != nil {
		t.Fatalf("voice state handler: %v", err)
	}
	if len(sideEffects.Deleted) != 0 || len(states.Deleted) != 0 {
		t.Fatalf("deleted channels=%#v states=%#v", sideEffects.Deleted, states.Deleted)
	}
}

func TestVoiceRoomDynamicNameMatchesLegacyFirstReplacement(t *testing.T) {
	if got := voiceRoomDynamicName("  {name}-{name}  ", "Yoru", "user-1"); got != "  Yoru-{name}  " {
		t.Fatalf("dynamic room name = %q", got)
	}
}

func portsChannel(guildID string, channelID string, parentID string, name string, channelType int, overwrites []ports.PermissionOverwrite) ports.ChannelRef {
	return ports.ChannelRef{
		GuildID:              guildID,
		ChannelID:            channelID,
		ParentID:             parentID,
		Name:                 name,
		Type:                 channelType,
		PermissionOverwrites: append([]ports.PermissionOverwrite(nil), overwrites...),
	}
}

func hasOverwrite(overwrites []ports.PermissionOverwrite, id string, overwriteType int, allow int64, deny int64) bool {
	for _, overwrite := range overwrites {
		if overwrite.ID == id && overwrite.Type == overwriteType && overwrite.Allow&allow == allow && overwrite.Deny&deny == deny {
			return true
		}
	}
	return false
}

func findOverwrite(overwrites []ports.PermissionOverwrite, id string, overwriteType int) (ports.PermissionOverwrite, bool) {
	for _, overwrite := range overwrites {
		if overwrite.ID == id && overwrite.Type == overwriteType {
			return overwrite, true
		}
	}
	return ports.PermissionOverwrite{}, false
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

type voiceFixedClock struct {
	now time.Time
}

func (c voiceFixedClock) Now() time.Time {
	return c.now
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
