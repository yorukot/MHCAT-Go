package voice

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestVoiceRoomMessagesMatchLegacyVisibleContract(t *testing.T) {
	tests := []struct {
		name string
		got  responses.Message
		want responses.Message
	}{
		{
			name: "set success",
			got:  voiceSetSuccessMessage(),
			want: responses.Message{
				Embeds:          []responses.Embed{{Title: legacyDoneEmoji + " | 成功進行設定", Description: legacyVoiceEmoji + " 你成功對語音包廂進行`設定`", Color: voiceSuccessColor}},
				AllowedMentions: &responses.AllowedMentions{},
			},
		},
		{
			name: "trigger delete success",
			got:  voiceDeleteTriggerSuccessMessage(),
			want: responses.Message{
				Embeds:          []responses.Embed{{Title: legacyDoneEmoji + "成功進行刪除", Description: legacyDeleteEmoji + "你成功對這個設定刪除", Color: voiceSuccessColor}},
				AllowedMentions: &responses.AllowedMentions{},
			},
		},
		{
			name: "category delete success",
			got:  voiceDeleteParentSuccessMessage(),
			want: responses.Message{
				Embeds:          []responses.Embed{{Title: "成功進行刪除", Description: "你成功對這個設定刪除", Color: voiceSuccessColor}},
				AllowedMentions: &responses.AllowedMentions{},
			},
		},
		{
			name: "lock success preserves password",
			got:  voiceLockSuccessMessage(" secret "),
			want: responses.Message{
				Embeds:          []responses.Embed{{Title: legacyDoneEmoji + " | 成功進行設定", Description: legacyVoiceEmoji + " 你成功對語音包廂密碼進行設定為: secret ", Color: voiceSuccessColor}},
				AllowedMentions: &responses.AllowedMentions{},
				Ephemeral:       true,
			},
		},
		{
			name: "missing lock answer",
			got:  voiceLockAnswerMissingMessage(),
			want: responses.Message{
				Embeds:          []responses.Embed{{Title: "很抱歉，該包廂可能已被刪除!", Color: voiceErrorColor}},
				AllowedMentions: &responses.AllowedMentions{},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !reflect.DeepEqual(test.got, test.want) {
				t.Fatalf("message = %#v, want %#v", test.got, test.want)
			}
		})
	}

	success := voiceLockAnswerSuccessMessage("guild-1", "voice-1")
	wrong := voiceLockAnswerWrongMessage("guild-1", "voice-1")
	if success.Embeds[0].Title != legacyUnlockEmoji+" | 您成功輸入正確密碼\n可以重新加入語音頻道囉!" || success.Embeds[0].Color != legacyVoiceLockColor {
		t.Fatalf("answer success = %#v", success)
	}
	if wrong.Embeds[0].Title != "<a:Discord_AnimatedNo:1015989839809757295> | 你的密碼輸入錯誤!請重新加入語音頻道後在試一次!" || wrong.Embeds[0].Color != legacyVoiceLockErrorColor {
		t.Fatalf("answer wrong = %#v", wrong)
	}
	for _, message := range []responses.Message{success, wrong} {
		if len(message.Components) != 1 || len(message.Components[0].Components) != 1 || message.Components[0].Components[0].URL != "https://discord.com/channels/guild-1/voice-1" {
			t.Fatalf("answer link = %#v", message.Components)
		}
	}
}

func TestVoiceRoomDeletePreservesLegacyStageCategoryBranch(t *testing.T) {
	repo := fakemongo.NewVoiceRoomConfigRepository()
	repo.Configs["guild-1\x00stage-1"] = domain.VoiceRoomConfig{
		GuildID:          "guild-1",
		TriggerChannelID: "stage-1",
		ParentID:         "category-1",
		Name:             "{name}",
	}
	module := NewModule(repo, nil)
	interaction := fakediscord.SlashInteraction(VoiceRoomDeleteCommandName)
	interaction.Actor.PermissionBits = permissionManageMessages
	interaction.CommandOptions = map[string]interactions.CommandOptionValue{
		optionChannelOrGroup: {
			Type:        interactions.CommandOptionChannel,
			String:      "stage-1",
			ChannelType: 13,
		},
	}
	responder := fakediscord.NewResponder()
	if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("delete handler: %v", err)
	}
	if _, ok := repo.Configs["guild-1\x00stage-1"]; !ok {
		t.Fatal("legacy stage branch should not delete the trigger row")
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "你沒有對這個類別沒有設定喔!") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestVoiceRoomParentlessTriggerUsesSafeGoBehavior(t *testing.T) {
	configs := fakemongo.NewVoiceRoomConfigRepository()
	configs.Configs["guild-1\x00trigger-1"] = domain.VoiceRoomConfig{
		GuildID:          "guild-1",
		TriggerChannelID: "trigger-1",
		ParentID:         "stale-parent",
		Name:             "{name} room",
	}
	sideEffects := fakediscord.NewSideEffects()
	sideEffects.Channels = append(sideEffects.Channels, ports.ChannelRef{
		GuildID:   "guild-1",
		ChannelID: "trigger-1",
		Name:      "Join to create",
		Type:      discordChannelTypeVoice,
	})
	module := NewRoomEventModule(configs, fakemongo.NewVoiceRoomStateRepository(), fakemongo.NewVoiceRoomLockRepository(), sideEffects, sideEffects, sideEffects)
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
	if len(sideEffects.Created) != 1 || sideEffects.Created[0].ParentID != "" || sideEffects.Created[0].Name != "Yoru room" {
		t.Fatalf("created channels = %#v", sideEffects.Created)
	}
}
