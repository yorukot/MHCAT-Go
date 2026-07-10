package roles

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestConfigureReactionSavesLegacyConfigAndAddsReaction(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	discord := fakediscord.NewSideEffects()
	discord.AssignableRoles["guild-1/role-1"] = true
	discord.CachedEmojiIDs["123456789012345678"] = true
	seedReactionMessage(discord, "guild-1", "channel-1", "message-1")
	service := SelectionService{Repository: repo, RoleInspector: discord, Reactions: discord}

	config, err := service.ConfigureReaction(context.Background(), ReactionSetCommand{
		GuildID:    "guild-1",
		MessageURL: "https://discord.com/channels/guild-1/channel-1/message-1",
		RoleID:     "role-1",
		Emoji:      "<:mhcat:123456789012345678>",
	})
	if err != nil {
		t.Fatalf("ConfigureReaction: %v", err)
	}
	if config.React != "123456789012345678" || config.MessageID != "message-1" {
		t.Fatalf("config = %#v", config)
	}
	if len(discord.Reactions) != 1 || discord.Reactions[0].Emoji != "mhcat:123456789012345678" {
		t.Fatalf("reactions = %#v", discord.Reactions)
	}
}

func TestDeleteReactionRemovesConfig(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	repo.Reactions["guild-1/message-1/✅"] = domain.RoleReactionConfig{GuildID: "guild-1", MessageID: "message-1", React: "✅", RoleID: "role-1"}
	discord := fakediscord.NewSideEffects()
	seedReactionMessage(discord, "guild-1", "channel-1", "message-1")
	service := SelectionService{Repository: repo, Reactions: discord}

	err := service.DeleteReaction(context.Background(), ReactionDeleteCommand{
		GuildID:    "guild-1",
		MessageURL: "https://discord.com/channels/guild-1/channel-1/message-1",
		Emoji:      "✅",
	})
	if err != nil {
		t.Fatalf("DeleteReaction: %v", err)
	}
	if _, ok := repo.Reactions["guild-1/message-1/✅"]; ok {
		t.Fatalf("reaction config should be deleted")
	}
}

func TestDeleteReactionNormalizesCustomEmojiAsReliabilityFix(t *testing.T) {
	const emojiID = "123456789012345678"
	repo := fakemongo.NewRoleSelectionRepository()
	repo.Reactions["guild-1/message-1/"+emojiID] = domain.RoleReactionConfig{GuildID: "guild-1", MessageID: "message-1", React: emojiID, RoleID: "role-1"}
	discord := fakediscord.NewSideEffects()
	seedReactionMessage(discord, "guild-1", "channel-1", "message-1")
	service := SelectionService{Repository: repo, Reactions: discord}

	err := service.DeleteReaction(context.Background(), ReactionDeleteCommand{
		GuildID:    "guild-1",
		MessageURL: "https://discord.com/channels/guild-1/channel-1/message-1",
		Emoji:      "<:mhcat:" + emojiID + ">",
	})
	if err != nil {
		t.Fatalf("DeleteReaction: %v", err)
	}
	if _, ok := repo.Reactions["guild-1/message-1/"+emojiID]; ok {
		t.Fatal("custom reaction config should be deleted by its stored ID")
	}
	if len(discord.Reactions) != 1 || discord.Reactions[0].Emoji != "mhcat:"+emojiID {
		t.Fatalf("reactions = %#v", discord.Reactions)
	}
}

func TestApplyReactionAddsAndRemovesRole(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	repo.Reactions["guild-1/message-1/✅"] = domain.RoleReactionConfig{GuildID: "guild-1", MessageID: "message-1", React: "✅", RoleID: "role-1"}
	discord := fakediscord.NewSideEffects()
	service := SelectionService{Repository: repo, Roles: discord}

	if err := service.ApplyReaction(context.Background(), ReactionApplyCommand{GuildID: "guild-1", MessageID: "message-1", React: "✅", UserID: "user-1"}); err != nil {
		t.Fatalf("ApplyReaction add: %v", err)
	}
	if err := service.ApplyReaction(context.Background(), ReactionApplyCommand{GuildID: "guild-1", MessageID: "message-1", React: "✅", UserID: "user-1", Remove: true}); err != nil {
		t.Fatalf("ApplyReaction remove: %v", err)
	}
	if len(discord.AddedRoles) != 1 || discord.AddedRoles[0].RoleID != "role-1" {
		t.Fatalf("added roles = %#v", discord.AddedRoles)
	}
	if len(discord.RemovedRoles) != 1 || discord.RemovedRoles[0].RoleID != "role-1" {
		t.Fatalf("removed roles = %#v", discord.RemovedRoles)
	}
}

func TestPrepareAndApplyButton(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	discord := fakediscord.NewSideEffects()
	discord.AssignableRoles["guild-1/role-1"] = true
	service := SelectionService{Repository: repo, RoleInspector: discord, Roles: discord}

	prepared, err := service.PrepareButton(context.Background(), ButtonPrepareCommand{GuildID: "guild-1", RoleID: "role-1", BaseID: "2026070901011234"})
	if err != nil {
		t.Fatalf("PrepareButton: %v", err)
	}
	if prepared.AddID != "2026070901011234add" || prepared.RemoveID != "2026070901011234delete" {
		t.Fatalf("prepared = %#v", prepared)
	}
	if err := service.ApplyButton(context.Background(), ButtonApplyCommand{GuildID: "guild-1", UserID: "user-1", Number: prepared.AddID}); err != nil {
		t.Fatalf("ApplyButton add: %v", err)
	}
	if err := service.ApplyButton(context.Background(), ButtonApplyCommand{GuildID: "guild-1", UserID: "user-1", Number: prepared.RemoveID, Remove: true, ActorRoleIDs: []string{"role-1"}}); err != nil {
		t.Fatalf("ApplyButton remove: %v", err)
	}
}

func TestApplyButtonRejectsRoleStateMismatches(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	repo.Buttons["guild-1/number-add"] = domain.RoleButtonConfig{GuildID: "guild-1", Number: "number-add", RoleID: "role-1"}
	repo.Buttons["guild-1/number-delete"] = domain.RoleButtonConfig{GuildID: "guild-1", Number: "number-delete", RoleID: "role-1"}
	discord := fakediscord.NewSideEffects()
	discord.AssignableRoles["guild-1/role-1"] = true
	service := SelectionService{Repository: repo, RoleInspector: discord, Roles: discord}

	err := service.ApplyButton(context.Background(), ButtonApplyCommand{GuildID: "guild-1", UserID: "user-1", Number: "number-add", ActorRoleIDs: []string{"role-1"}})
	if !errors.Is(err, ErrRoleAlreadyAssigned) {
		t.Fatalf("expected already assigned, got %v", err)
	}
	err = service.ApplyButton(context.Background(), ButtonApplyCommand{GuildID: "guild-1", UserID: "user-1", Number: "number-delete", Remove: true})
	if !errors.Is(err, ErrRoleNotAssigned) {
		t.Fatalf("expected not assigned, got %v", err)
	}
}

func TestConfigureReactionRejectsUnassignableRole(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	discord := fakediscord.NewSideEffects()
	service := SelectionService{Repository: repo, RoleInspector: discord, Reactions: discord}
	_, err := service.ConfigureReaction(context.Background(), ReactionSetCommand{
		GuildID:    "guild-1",
		MessageURL: "https://discord.com/channels/guild-1/channel-1/message-1",
		RoleID:     "role-1",
		Emoji:      "✅",
	})
	if !errors.Is(err, ports.ErrDiscordRoleNotAssignable) {
		t.Fatalf("expected not assignable, got %v", err)
	}
}

func TestConfigureReactionChecksRoleBeforeMessageURL(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	discord := fakediscord.NewSideEffects()
	service := SelectionService{Repository: repo, RoleInspector: discord, Reactions: discord}

	_, err := service.ConfigureReaction(context.Background(), ReactionSetCommand{
		GuildID:    "guild-1",
		MessageURL: "not-a-message-url",
		RoleID:     "role-1",
		Emoji:      "not-an-emoji",
	})
	if !errors.Is(err, ports.ErrDiscordRoleNotAssignable) {
		t.Fatalf("expected role hierarchy error before URL validation, got %v", err)
	}
}

func TestConfigureReactionRejectsInvalidLegacyEmoji(t *testing.T) {
	tests := []struct {
		name  string
		emoji string
	}{
		{name: "plain text", emoji: "not-an-emoji"},
		{name: "uncached custom emoji", emoji: "<:mhcat:123456789012345678>"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := fakemongo.NewRoleSelectionRepository()
			discord := fakediscord.NewSideEffects()
			discord.AssignableRoles["guild-1/role-1"] = true
			seedReactionChannel(discord, "guild-1", "channel-1")
			service := SelectionService{Repository: repo, RoleInspector: discord, Reactions: discord}

			_, err := service.ConfigureReaction(context.Background(), ReactionSetCommand{
				GuildID:    "guild-1",
				MessageURL: "https://discord.com/channels/guild-1/channel-1/message-1",
				RoleID:     "role-1",
				Emoji:      test.emoji,
			})
			if !errors.Is(err, domain.ErrInvalidRoleSelectionEmoji) {
				t.Fatalf("expected invalid emoji, got %v", err)
			}
			if len(discord.MessageFetches) != 0 || len(discord.Reactions) != 0 || len(repo.Reactions) != 0 {
				t.Fatalf("invalid emoji side effects: fetches=%#v reactions=%#v configs=%#v", discord.MessageFetches, discord.Reactions, repo.Reactions)
			}
		})
	}
}

func TestConfigureReactionChecksCachedChannelBeforeEmoji(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	discord := fakediscord.NewSideEffects()
	discord.AssignableRoles["guild-1/role-1"] = true
	service := SelectionService{Repository: repo, RoleInspector: discord, Reactions: discord}

	_, err := service.ConfigureReaction(context.Background(), ReactionSetCommand{
		GuildID:    "guild-1",
		MessageURL: "https://discord.com/channels/guild-1/missing-channel/message-1",
		RoleID:     "role-1",
		Emoji:      "not-an-emoji",
	})
	if !errors.Is(err, ports.ErrChannelNotFound) {
		t.Fatalf("expected cached channel error before emoji validation, got %v", err)
	}
	if len(discord.MessageFetches) != 0 || len(discord.Reactions) != 0 {
		t.Fatalf("missing channel side effects: fetches=%#v reactions=%#v", discord.MessageFetches, discord.Reactions)
	}
}

func TestConfigureReactionChecksMessageAfterEmoji(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	discord := fakediscord.NewSideEffects()
	discord.AssignableRoles["guild-1/role-1"] = true
	seedReactionChannel(discord, "guild-1", "channel-1")
	service := SelectionService{Repository: repo, RoleInspector: discord, Reactions: discord}

	_, err := service.ConfigureReaction(context.Background(), ReactionSetCommand{
		GuildID:    "guild-1",
		MessageURL: "https://discord.com/channels/guild-1/channel-1/missing-message",
		RoleID:     "role-1",
		Emoji:      "✅",
	})
	if !errors.Is(err, ports.ErrDiscordMessageNotFound) {
		t.Fatalf("expected missing message, got %v", err)
	}
	if len(discord.MessageFetches) != 1 || len(discord.Reactions) != 0 || len(repo.Reactions) != 0 {
		t.Fatalf("missing message side effects: fetches=%#v reactions=%#v configs=%#v", discord.MessageFetches, discord.Reactions, repo.Reactions)
	}
}

func TestConfigureReactionIgnoresURLGuildSegmentAfterCachedChannelLookup(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	discord := fakediscord.NewSideEffects()
	discord.AssignableRoles["guild-1/role-1"] = true
	seedReactionMessage(discord, "guild-1", "channel-1", "message-1")
	service := SelectionService{Repository: repo, RoleInspector: discord, Reactions: discord}

	config, err := service.ConfigureReaction(context.Background(), ReactionSetCommand{
		GuildID:    "guild-1",
		MessageURL: "https://discord.com/channels/another-guild/channel-1/message-1",
		RoleID:     "role-1",
		Emoji:      "✅",
	})
	if err != nil {
		t.Fatalf("ConfigureReaction: %v", err)
	}
	if config.GuildID != "guild-1" || config.MessageID != "message-1" {
		t.Fatalf("config = %#v", config)
	}
}

func TestDeleteReactionRejectsLegacyDiscordAppHost(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	discord := fakediscord.NewSideEffects()
	service := SelectionService{Repository: repo, Reactions: discord}

	err := service.DeleteReaction(context.Background(), ReactionDeleteCommand{
		GuildID:    "guild-1",
		MessageURL: "https://discordapp.com/channels/guild-1/channel-1/message-1",
		Emoji:      "✅",
	})
	if !errors.Is(err, domain.ErrInvalidRoleSelectionConfig) {
		t.Fatalf("expected invalid URL, got %v", err)
	}
	if len(discord.Reactions) != 0 {
		t.Fatalf("reactions = %#v", discord.Reactions)
	}
}

func TestDeleteReactionIgnoresURLGuildSegmentAfterCachedChannelLookup(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	repo.Reactions["guild-1/message-1/✅"] = domain.RoleReactionConfig{GuildID: "guild-1", MessageID: "message-1", React: "✅", RoleID: "role-1"}
	discord := fakediscord.NewSideEffects()
	seedReactionMessage(discord, "guild-1", "channel-1", "message-1")
	service := SelectionService{Repository: repo, Reactions: discord}

	err := service.DeleteReaction(context.Background(), ReactionDeleteCommand{
		GuildID:    "guild-1",
		MessageURL: "https://discord.com/channels/another-guild/channel-1/message-1",
		Emoji:      "✅",
	})
	if err != nil {
		t.Fatalf("DeleteReaction: %v", err)
	}
	if _, ok := repo.Reactions["guild-1/message-1/✅"]; ok {
		t.Fatal("reaction config should be deleted")
	}
}

func seedReactionChannel(discord *fakediscord.SideEffects, guildID string, channelID string) {
	discord.Channels = append(discord.Channels, ports.ChannelRef{GuildID: guildID, ChannelID: channelID})
}

func seedReactionMessage(discord *fakediscord.SideEffects, guildID string, channelID string, messageID string) {
	seedReactionChannel(discord, guildID, channelID)
	discord.Messages[channelID+"/"+messageID] = ports.MessageRef{ChannelID: channelID, MessageID: messageID}
}
