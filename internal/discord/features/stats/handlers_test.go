package stats

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestQueryHandlerRendersLegacyStaticEmbed(t *testing.T) {
	usage := &fakeusage.Tracker{}
	module := NewModuleWithColor(usage, func() int { return 0x123456 })
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction(StatsQueryCommandName)

	if err := module.QueryHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Replies) != 1 {
		t.Fatalf("replies = %#v", responder.Replies)
	}
	reply := responder.Replies[0]
	if len(reply.Embeds) != 1 {
		t.Fatalf("embeds = %#v", reply.Embeds)
	}
	embed := reply.Embeds[0]
	if embed.Title != "統計系統查詢" || embed.Color != 0x123456 {
		t.Fatalf("embed = %#v", embed)
	}
	wantDescription := "\n" +
		"        我的統計系統是每**10分鐘更新一次**`(因為discord api每10分鐘才能更新一次)`\n" +
		"        輸入 /統計系統創建 [選擇要`文字頻道`或是`語音頻道`] [輸入想創建的統計名稱]\n" +
		"        \n" +
		"        **用戶查詢**\n" +
		"        ```\n" +
		"用戶總數 (伺服器的總人數)\n" +
		"使用者總數 (伺服器非機器人人數)\n" +
		"機器人數 (伺服器總共的機器人數量)```\n" +
		"        **伺服器頻道**\n" +
		"        ```\n" +
		"頻道數量 (頻道總數量)\n" +
		"文字頻道數量 (文字頻道總數)\n" +
		"語音頻道數量 (語音頻道總數)```\n" +
		"        "
	if embed.Description != wantDescription || reply.Ephemeral {
		t.Fatalf("reply = %#v", reply)
	}
	if reply.AllowedMentions == nil {
		t.Fatal("expected allowed mentions to be disabled explicitly")
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != StatsQueryCommandName || usage.Events[0].Feature != "stats-query" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestQueryHandlerDoesNotDeferLikeLegacy(t *testing.T) {
	module := NewModule(nil)
	responder := fakediscord.NewResponder()
	if err := module.QueryHandler()(context.Background(), fakediscord.SlashInteraction(StatsQueryCommandName), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 0 || len(responder.Edits) != 0 {
		t.Fatalf("unexpected defer/edit: defers=%#v edits=%#v", responder.Defers, responder.Edits)
	}
}

func TestDeleteHandlerRequiresManageMessages(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	repo.Put(domain.StatsConfig{GuildID: "guild-1", ParentID: "parent-1"})
	module := NewDeleteModule(repo, nil)
	responder := fakediscord.NewResponder()

	if err := module.DeleteHandler()(context.Background(), fakediscord.SlashInteraction(StatsDeleteCommandName), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "你需要有`訊息管理`") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	if _, ok := repo.Configs["guild-1"]; !ok {
		t.Fatal("permission failure should not delete stats config")
	}
}

func TestStatsDeleteMessagesMatchLegacyPayloads(t *testing.T) {
	tests := []struct {
		name      string
		message   responses.Message
		wantTitle string
		wantColor int
	}{
		{
			name:      "success",
			message:   statsDeleteSuccessMessage("parent-1"),
			wantTitle: "<a:greentick:980496858445135893> | 成功刪除，該類別以下的頻道我已經管不了囉!(類別id:parent-1)",
			wantColor: statsSuccessColor,
		},
		{
			name:      "missing config",
			message:   statsErrorMessage("你還沒有創建過統計數據，是要刪除甚麼啦!"),
			wantTitle: "<a:Discord_AnimatedNo:1015989839809757295> | 你還沒有創建過統計數據，是要刪除甚麼啦!",
			wantColor: statsErrorColor,
		},
		{
			name:      "unknown error",
			message:   statsErrorMessage("很抱歉，出現了未知的錯誤，請重試!"),
			wantTitle: "<a:Discord_AnimatedNo:1015989839809757295> | 很抱歉，出現了未知的錯誤，請重試!",
			wantColor: statsErrorColor,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			message := test.message
			if len(message.Embeds) != 1 || message.Embeds[0].Title != test.wantTitle || message.Embeds[0].Color != test.wantColor {
				t.Fatalf("message = %#v", message)
			}
			if message.Content != "" || message.Embeds[0].Description != "" || len(message.Components) != 0 || len(message.Files) != 0 || message.Ephemeral || message.AllowedMentions == nil {
				t.Fatalf("unexpected payload fields = %#v", message)
			}
		})
	}
}

func TestCreateHandlerRequiresManageMessages(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	discord := fakediscord.NewSideEffects()
	module := NewCreateModule(repo, discord, discord, nil, "bot-1")
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(StatsCreateCommandName, "", map[string]string{
		statsOptionChannelType: "文字頻道",
	})

	if err := module.CreateHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(responder.Follow) != 1 || !strings.Contains(responder.Follow[0].Embeds[0].Title, "你需要有`訊息管理`") {
		t.Fatalf("followups = %#v", responder.Follow)
	}
	if len(discord.Created) != 0 {
		t.Fatalf("created channels = %#v", discord.Created)
	}
}

func TestStatsCreateMessagesMatchLegacyPayloads(t *testing.T) {
	tests := []struct {
		name      string
		message   func() responses.Message
		wantTitle string
		wantColor int
	}{
		{
			name:      "loading",
			message:   statsCreateLoadingMessage,
			wantTitle: "<a:lodding:980493229592043581> | 正在進行設置中!",
			wantColor: statsSuccessColor,
		},
		{
			name:      "success",
			message:   statsCreateSuccessMessage,
			wantTitle: "<a:greentick:980496858445135893> | 成功創建!頻道(不要動到數字就沒問題)跟類別的名稱都能自行更改喔!",
			wantColor: statsSuccessColor,
		},
		{
			name:      "permission",
			message:   func() responses.Message { return statsErrorMessage("你需要有`訊息管理`才能使用此指令") },
			wantTitle: "<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`訊息管理`才能使用此指令",
			wantColor: statsErrorColor,
		},
		{
			name:      "invalid channel type",
			message:   func() responses.Message { return statsCreateErrorMessage(domain.ErrInvalidStatsChannelType) },
			wantTitle: "<a:Discord_AnimatedNo:1015989839809757295> | 你沒有進行設置要文字頻道還是語音頻道!或是你打錯了!",
			wantColor: statsErrorColor,
		},
		{
			name:      "missing option",
			message:   func() responses.Message { return statsCreateErrorMessage(domain.ErrStatsOptionRequired) },
			wantTitle: "<a:Discord_AnimatedNo:1015989839809757295> | 由於你已經創建過了，所以你必須說明你要創建的統計名稱，或是刪除現有的統計資料(使用統計資料刪除)!",
			wantColor: statsErrorColor,
		},
		{
			name:      "duplicate option",
			message:   func() responses.Message { return statsCreateErrorMessage(domain.ErrStatsChannelAlreadyExists) },
			wantTitle: "<a:Discord_AnimatedNo:1015989839809757295> | 這個統計你已經創建過了!",
			wantColor: statsErrorColor,
		},
		{
			name:      "invalid option",
			message:   func() responses.Message { return statsCreateErrorMessage(domain.ErrInvalidStatsOption) },
			wantTitle: "<a:Discord_AnimatedNo:1015989839809757295> | 沒有這項統計可以創建欸QQ",
			wantColor: statsErrorColor,
		},
		{
			name:      "unknown error",
			message:   func() responses.Message { return statsCreateErrorMessage(errors.New("database unavailable")) },
			wantTitle: "<a:Discord_AnimatedNo:1015989839809757295> | 很抱歉，出現了未知的錯誤，請重試!",
			wantColor: statsErrorColor,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			message := test.message()
			if len(message.Embeds) != 1 || message.Embeds[0].Title != test.wantTitle || message.Embeds[0].Color != test.wantColor {
				t.Fatalf("message = %#v", message)
			}
			if message.Content != "" || message.Embeds[0].Description != "" || len(message.Components) != 0 || len(message.Files) != 0 || message.Ephemeral {
				t.Fatalf("unexpected payload fields = %#v", message)
			}
			if message.AllowedMentions == nil {
				t.Fatal("allowed mentions must be disabled explicitly")
			}
		})
	}
}

func TestCreateHandlerCreatesStatsWithLegacySuccess(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	discord := fakediscord.NewSideEffects()
	discord.TotalMembers = 15
	discord.NonBotMembers = 12
	usage := &fakeusage.Tracker{}
	module := NewCreateModule(repo, discord, discord, usage, "")
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(StatsCreateCommandName, "", map[string]string{
		statsOptionChannelType: "文字頻道",
	})
	interaction.Actor.PermissionBits = permissionManageMessages
	interaction.ApplicationID = "bot-1"

	if err := module.CreateHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 || len(responder.Follow) != 1 || len(responder.FollowEdits) != 1 {
		t.Fatalf("defers=%#v followups=%#v followup edits=%#v", responder.Defers, responder.Follow, responder.FollowEdits)
	}
	if title := responder.Follow[0].Embeds[0].Title; title != "<a:lodding:980493229592043581> | 正在進行設置中!" {
		t.Fatalf("loading title = %q", title)
	}
	if responder.FollowEdits[0].MessageID != responder.FollowIDs[0] {
		t.Fatalf("follow-up edits = %#v ids=%#v", responder.FollowEdits, responder.FollowIDs)
	}
	if title := responder.FollowEdits[0].Message.Embeds[0].Title; title != "<a:greentick:980496858445135893> | 成功創建!頻道(不要動到數字就沒問題)跟類別的名稱都能自行更改喔!" {
		t.Fatalf("success title = %q", title)
	}
	if len(discord.Created) != 4 || discord.Created[1].Name != "總人數: 15" {
		t.Fatalf("created channels = %#v", discord.Created)
	}
	for _, request := range discord.Created[1:] {
		if len(request.PermissionOverwrites) != 2 || request.PermissionOverwrites[0].ID != "bot-1" || request.PermissionOverwrites[0].Type != 1 {
			t.Fatalf("permission overwrites = %#v", request.PermissionOverwrites)
		}
	}
	if saved := repo.Configs["guild-1"]; saved.MemberNumberName != "15" || saved.UserNumberName != "12" || saved.BotNumberName != "3" {
		t.Fatalf("saved = %#v", saved)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != StatsCreateCommandName || usage.Events[0].Feature != "stats-create" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestCreateHandlerExistingStatsRequiresOptionWithLegacyError(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	repo.Put(domain.StatsConfig{GuildID: "guild-1", ParentID: "parent-1"})
	discord := fakediscord.NewSideEffects()
	module := NewCreateModule(repo, discord, discord, nil, "")
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(StatsCreateCommandName, "", map[string]string{
		statsOptionChannelType: "文字頻道",
	})
	interaction.Actor.PermissionBits = permissionManageMessages

	if err := module.CreateHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Follow) != 1 || !strings.Contains(responder.Follow[0].Embeds[0].Title, "由於你已經創建過了") {
		t.Fatalf("followups = %#v", responder.Follow)
	}
}

func TestCreateHandlerAddsOptionalStatsChannel(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	repo.Put(domain.StatsConfig{GuildID: "guild-1", ParentID: "parent-1"})
	discord := fakediscord.NewSideEffects()
	discord.Channels = append(discord.Channels, ports.ChannelRef{GuildID: "guild-1", ChannelID: "parent-1", Name: "stats", Type: 4})
	discord.TextChannelCount = 9
	module := NewCreateModule(repo, discord, discord, nil, "")
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(StatsCreateCommandName, "", map[string]string{
		statsOptionChannelType: "文字頻道",
		statsOptionStat:        "文字頻道數量",
	})
	interaction.Actor.PermissionBits = permissionManageMessages

	if err := module.CreateHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(discord.Created) != 1 || discord.Created[0].Name != "總文字頻道數: 9" || discord.Created[0].ParentID != "parent-1" {
		t.Fatalf("created channels = %#v", discord.Created)
	}
	if saved := repo.Configs["guild-1"]; saved.TextNumberName != "9" || saved.TextNumberID == "" {
		t.Fatalf("saved = %#v", saved)
	}
}

func TestRoleHandlerRequiresManageMessages(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	repo.Put(domain.StatsConfig{GuildID: "guild-1", ParentID: "parent-1"})
	discord := fakediscord.NewSideEffects()
	discord.RoleNames["guild-1/role-1"] = "VIP"
	module := NewRoleModule(repo, repo, discord, discord, nil, "bot-1")
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(StatsRoleCommandName, "", map[string]string{
		statsOptionChannelType: "文字頻道",
		statsOptionRole:        "role-1",
	})

	if err := module.RoleHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(responder.Follow) != 1 || responder.Follow[0].Embeds[0].Title != "<a:lodding:980493229592043581> | 正在進行設置中!" {
		t.Fatalf("followups = %#v", responder.Follow)
	}
	if len(responder.FollowEdits) != 1 || responder.FollowEdits[0].MessageID != responder.FollowIDs[0] || !strings.Contains(responder.FollowEdits[0].Message.Embeds[0].Title, "你需要有`訊息管理`") {
		t.Fatalf("follow-up edits = %#v ids=%#v", responder.FollowEdits, responder.FollowIDs)
	}
	if len(responder.Edits) != 0 {
		t.Fatalf("original edits = %#v", responder.Edits)
	}
	if len(discord.Created) != 0 {
		t.Fatalf("created channels = %#v", discord.Created)
	}
}

func TestStatsRoleMessagesMatchLegacyPayloads(t *testing.T) {
	tests := []struct {
		name            string
		message         responses.Message
		wantTitle       string
		wantDescription string
		wantColor       int
	}{
		{
			name:            "success",
			message:         statsRoleSuccessMessage("channel-1"),
			wantTitle:       "統計特定身分組成功創建",
			wantDescription: "已成功為您創建統計特定身分組\n頻道:<#channel-1> 名字可以更改喔，不要動到數字就好awa",
			wantColor:       statsSuccessColor,
		},
		{
			name:      "invalid channel type",
			message:   statsRoleErrorMessage(domain.ErrInvalidStatsChannelType),
			wantTitle: "<a:Discord_AnimatedNo:1015989839809757295> | 你沒有進行設置要文字頻道還是語音頻道!或是你打錯了!",
			wantColor: statsErrorColor,
		},
		{
			name:      "missing base stats",
			message:   statsRoleErrorMessage(ports.ErrStatsConfigMissing),
			wantTitle: "<a:Discord_AnimatedNo:1015989839809757295> | 你還沒創建過統計頻道，請先使用`/統計系統創建`",
			wantColor: statsErrorColor,
		},
		{
			name:      "unknown error",
			message:   statsRoleErrorMessage(errors.New("database unavailable")),
			wantTitle: "<a:Discord_AnimatedNo:1015989839809757295> | 很抱歉，出現了未知的錯誤，請重試!",
			wantColor: statsErrorColor,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			message := test.message
			if len(message.Embeds) != 1 || message.Embeds[0].Title != test.wantTitle || message.Embeds[0].Description != test.wantDescription || message.Embeds[0].Color != test.wantColor {
				t.Fatalf("message = %#v", message)
			}
			if message.Content != "" || len(message.Components) != 0 || len(message.Files) != 0 || message.Ephemeral || message.AllowedMentions == nil {
				t.Fatalf("unexpected payload fields = %#v", message)
			}
		})
	}
}

func TestRoleHandlerMissingBaseStatsUsesLegacyError(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	discord := fakediscord.NewSideEffects()
	discord.RoleNames["guild-1/role-1"] = "VIP"
	module := NewRoleModule(repo, repo, discord, discord, nil, "")
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(StatsRoleCommandName, "", map[string]string{
		statsOptionChannelType: "文字頻道",
		statsOptionRole:        "role-1",
	})
	interaction.Actor.PermissionBits = permissionManageMessages

	if err := module.RoleHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Follow) != 1 || len(responder.FollowEdits) != 1 || !strings.Contains(responder.FollowEdits[0].Message.Embeds[0].Title, "你還沒創建過統計頻道") {
		t.Fatalf("followups=%#v follow-up edits=%#v", responder.Follow, responder.FollowEdits)
	}
	if responder.FollowEdits[0].MessageID != responder.FollowIDs[0] || len(responder.Edits) != 0 {
		t.Fatalf("follow-up ids=%#v edits=%#v", responder.FollowIDs, responder.Edits)
	}
	if len(discord.Created) != 0 {
		t.Fatalf("created channels = %#v", discord.Created)
	}
}

func TestRoleHandlerCreatesLegacyRoleStatsChannel(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	repo.Put(domain.StatsConfig{GuildID: "guild-1", ParentID: "parent-1"})
	discord := fakediscord.NewSideEffects()
	discord.Channels = append(discord.Channels, ports.ChannelRef{GuildID: "guild-1", ChannelID: "parent-1", Name: "stats", Type: 4})
	discord.RoleNames["guild-1/role-1"] = "VIP"
	discord.RoleMemberCounts["guild-1/role-1"] = 6
	usage := &fakeusage.Tracker{}
	module := NewRoleModule(repo, repo, discord, discord, usage, "")
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(StatsRoleCommandName, "", map[string]string{
		statsOptionChannelType: "文字頻道",
		statsOptionRole:        "role-1",
	})
	interaction.Actor.PermissionBits = permissionManageMessages
	interaction.ApplicationID = "bot-1"

	if err := module.RoleHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 || len(responder.Follow) != 1 || len(responder.FollowEdits) != 1 {
		t.Fatalf("defers=%#v followups=%#v follow-up edits=%#v", responder.Defers, responder.Follow, responder.FollowEdits)
	}
	if responder.Follow[0].Embeds[0].Title != "<a:lodding:980493229592043581> | 正在進行設置中!" || responder.FollowEdits[0].MessageID != responder.FollowIDs[0] || len(responder.Edits) != 0 {
		t.Fatalf("followups=%#v follow-up edits=%#v original edits=%#v", responder.Follow, responder.FollowEdits, responder.Edits)
	}
	embed := responder.FollowEdits[0].Message.Embeds[0]
	if embed.Title != "統計特定身分組成功創建" || !strings.Contains(embed.Description, "頻道:<#created-channel-1> 名字可以更改喔，不要動到數字就好awa") {
		t.Fatalf("embed = %#v", embed)
	}
	if len(discord.Created) != 1 || discord.Created[0].Name != "VIP: 6" || discord.Created[0].ParentID != "parent-1" {
		t.Fatalf("created channels = %#v", discord.Created)
	}
	if overwrites := discord.Created[0].PermissionOverwrites; len(overwrites) != 2 || overwrites[0].ID != "bot-1" || overwrites[0].Type != 1 {
		t.Fatalf("permission overwrites = %#v", overwrites)
	}
	if saved := repo.RoleConfigs["guild-1/role-1"]; saved.ChannelID != "created-channel-1" || saved.ChannelName != "6" || saved.RoleID != "role-1" {
		t.Fatalf("saved role config = %#v", saved)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != StatsRoleCommandName || usage.Events[0].Feature != "stats-role-count" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestDeleteHandlerDeletesStatsConfigWithLegacySuccess(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	repo.Put(domain.StatsConfig{GuildID: "guild-1", ParentID: "parent-1"})
	usage := &fakeusage.Tracker{}
	module := NewDeleteModule(repo, usage)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction(StatsDeleteCommandName)
	interaction.Actor.PermissionBits = permissionManageMessages

	if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 || len(responder.Edits) != 1 {
		t.Fatalf("defers=%#v edits=%#v", responder.Defers, responder.Edits)
	}
	if title := responder.Edits[0].Embeds[0].Title; title != "<a:greentick:980496858445135893> | 成功刪除，該類別以下的頻道我已經管不了囉!(類別id:parent-1)" {
		t.Fatalf("title = %q", title)
	}
	if _, ok := repo.Configs["guild-1"]; ok {
		t.Fatal("stats config should be deleted")
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != StatsDeleteCommandName || usage.Events[0].Feature != "stats-delete" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestDeleteHandlerMissingStatsConfigUsesLegacyError(t *testing.T) {
	module := NewDeleteModule(fakemongo.NewStatsConfigRepository(), nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction(StatsDeleteCommandName)
	interaction.Actor.PermissionBits = permissionManageMessages

	if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "你還沒有創建過統計數據") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestDeleteHandlerUnknownErrorUsesSafeLegacyStyleError(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	repo.Err = ports.ErrCoinLimitExceeded
	module := NewDeleteModule(repo, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction(StatsDeleteCommandName)
	interaction.Actor.PermissionBits = permissionManageMessages

	if err := module.DeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "未知的錯誤") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}
