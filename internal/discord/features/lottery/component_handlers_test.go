package lottery

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakebotinfo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

const lotteryTestID = "1700000000000999lotter"

func TestLotteryEnterHandlerJoinsAndPreservesLegacyErrors(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	repo := fakemongo.NewLotteryRepository()
	repo.Lotteries["guild-1:"+lotteryTestID] = domain.Lottery{
		GuildID:        "guild-1",
		ID:             lotteryTestID,
		EndsAtUnix:     now.Add(time.Hour).Unix(),
		RequiredRoleID: "role-1",
	}
	module := NewComponentModule(repo, nil, nil, nil, lotteryFixedClock{now: now})
	interaction := fakediscord.ComponentInteractionFromID(lotteryTestID)
	interaction.Actor.RoleIDs = []string{"role-1"}
	responder := fakediscord.NewResponder()

	if err := module.EnterHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("enter: %v", err)
	}
	if len(responder.Defers) != 1 || !responder.Defers[0].Ephemeral || len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<a:green_tick:994529015652163614> | 成功參加抽獎!" {
		t.Fatalf("responses defers=%#v edits=%#v", responder.Defers, responder.Edits)
	}
	stored := repo.Lotteries["guild-1:"+lotteryTestID]
	if len(stored.Participants) != 1 || stored.Participants[0].UserID != "user-1" || stored.Participants[0].JoinedAtMillis != now.UnixMilli() {
		t.Fatalf("stored = %#v", stored)
	}

	duplicate := fakediscord.NewResponder()
	if err := module.EnterHandler()(context.Background(), interaction, duplicate); err != nil {
		t.Fatalf("duplicate: %v", err)
	}
	if len(duplicate.Edits) != 1 || !strings.Contains(duplicate.Edits[0].Content, "你無法重複參加") {
		t.Fatalf("duplicate edit = %#v", duplicate.Edits)
	}

	roleDeniedInteraction := interaction
	roleDeniedInteraction.Actor.UserID = "user-2"
	roleDeniedInteraction.Actor.RoleIDs = nil
	roleDenied := fakediscord.NewResponder()
	if err := module.EnterHandler()(context.Background(), roleDeniedInteraction, roleDenied); err != nil {
		t.Fatalf("role denied: %v", err)
	}
	if len(roleDenied.Edits) != 1 || !strings.Contains(roleDenied.Edits[0].Content, "創辦人設定你不能抽獎") {
		t.Fatalf("role denied edit = %#v", roleDenied.Edits)
	}
}

func TestLotterySearchHandlerRendersParticipantsExportAndOwnerControls(t *testing.T) {
	repo := fakemongo.NewLotteryRepository()
	repo.Lotteries["guild-1:"+lotteryTestID] = domain.Lottery{
		GuildID: "guild-1",
		ID:      lotteryTestID,
		OwnerID: "user-1",
		Participants: []domain.LotteryParticipant{
			{UserID: "user-1", JoinedAtMillis: 1_700_000_000_000},
			{UserID: "user-2", JoinedAtRaw: "legacy time"},
			{UserID: "user-3", JoinedAtRaw: "modern time"},
		},
	}
	info := &fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{ID: "guild-1", OwnerID: "guild-owner"}}
	members := fakediscord.NewSideEffects()
	members.MemberTagValues["user-1"] = "Owner#0001"
	members.MemberTagValues["user-3"] = "ModernUser"
	module := NewComponentModule(repo, info, members, nil, lotteryFixedClock{})
	module.color = func() int { return 0x123456 }
	interaction := fakediscord.ComponentInteractionFromID(lotteryTestID + "search")
	responder := fakediscord.NewResponder()

	if err := module.SearchHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(responder.Edits) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	message := responder.Edits[0]
	if message.Content == "" || len(message.Embeds) != 1 || message.Embeds[0].Title != "抽獎人數資訊" || message.Embeds[0].Color != 0x123456 {
		t.Fatalf("message = %#v", message)
	}
	if !strings.Contains(message.Embeds[0].Description, "`3`") || !strings.Contains(message.Embeds[0].Description, "`有`") || !strings.Contains(message.Embeds[0].Description, "Owner#0001") || !strings.Contains(message.Embeds[0].Description, "ModernUser#0") || !strings.Contains(message.Embeds[0].Description, "使用者已消失") {
		t.Fatalf("description = %q", message.Embeds[0].Description)
	}
	if len(message.Components) != 1 || message.Components[0].Components[0].CustomID != lotteryTestID+"restart" || message.Components[0].Components[1].CustomID != lotteryTestID+"stop" {
		t.Fatalf("components = %#v", message.Components)
	}
	if len(message.Files) != 1 || message.Files[0].Name != "discord.txt" {
		t.Fatalf("files = %#v", message.Files)
	}
	file := string(message.Files[0].Data)
	if !strings.Contains(file, "Owner#0001(id:user-1)|參加時間:2023/11/15\u200906:13:20 [台北標準時間]") || !strings.Contains(file, "使用者已退出伺服器!(id:user-2)|參加時間:legacy time") || !strings.Contains(file, "ModernUser#0(id:user-3)|參加時間:modern time") {
		t.Fatalf("file = %q", file)
	}
}

func TestLotterySearchHidesManagerControlsFromOtherUsers(t *testing.T) {
	repo := fakemongo.NewLotteryRepository()
	repo.Lotteries["guild-1:"+lotteryTestID] = domain.Lottery{GuildID: "guild-1", ID: lotteryTestID, OwnerID: "owner-1"}
	info := &fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{OwnerID: "guild-owner"}}
	module := NewComponentModule(repo, info, fakediscord.NewSideEffects(), nil, lotteryFixedClock{})
	interaction := fakediscord.ComponentInteractionFromID(lotteryTestID + "search")
	interaction.Actor.PermissionBits = permissionManageMessages
	responder := fakediscord.NewResponder()

	if err := module.SearchHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Content != "" || len(responder.Edits[0].Components) != 0 {
		t.Fatalf("edit = %#v", responder.Edits)
	}
}

func TestLegacyLotteryParticipantTimeMatchesNode20(t *testing.T) {
	tests := []struct {
		name   string
		millis int64
		want   string
	}{
		{name: "normal hour", millis: 1_700_000_000_000, want: "2023/11/15\u200906:13:20 [台北標準時間]"},
		{name: "midnight hour", millis: 1_699_977_601_000, want: "2023/11/15\u200924:00:01 [台北標準時間]"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := legacyLotteryParticipantTime(domain.LotteryParticipant{JoinedAtMillis: test.millis})
			if got != test.want {
				t.Fatalf("participant time = %q, want %q", got, test.want)
			}
		})
	}
}

func TestLotteryRerollSendsOneLegacyWinnerMessageAndEndsLottery(t *testing.T) {
	repo := fakemongo.NewLotteryRepository()
	repo.Lotteries["guild-1:"+lotteryTestID] = domain.Lottery{
		GuildID:     "guild-1",
		ID:          lotteryTestID,
		OwnerID:     "user-1",
		Gift:        "  Nitro  ",
		WinnerCount: 2,
		ChannelID:   "channel-1",
		Participants: []domain.LotteryParticipant{
			{UserID: "user-1"},
			{UserID: "user-2"},
		},
	}
	info := &fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{OwnerID: "guild-owner", BotDisplayColor: 0x123456}}
	messages := fakediscord.NewSideEffects()
	module := NewComponentModule(repo, info, nil, messages, lotteryFixedClock{})
	module.randomIndex = func(int) (int, error) { return 1, nil }
	interaction := fakediscord.ComponentInteractionFromID(lotteryTestID + "restart")
	responder := fakediscord.NewResponder()

	if err := module.RerollHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("reroll: %v", err)
	}
	if len(messages.Sent) != 1 || messages.Sent[0].ChannelID != "channel-1" {
		t.Fatalf("sent = %#v", messages.Sent)
	}
	sent := messages.Sent[0].Message
	wantDescription := "\n**<:celebration:997374188060946495> 恭喜:**\n<@user-2>\n<@user-2>\n<:gift:994585975445528576> **抽中:**   Nitro  \n"
	if sent.Content != "<@user-2><@user-2>" || len(sent.Embeds) != 1 || sent.Embeds[0].Color != 0x123456 || sent.Embeds[0].Description != wantDescription || len(sent.AllowedMentions.UserIDs) != 1 || sent.AllowedMentions.UserIDs[0] != "user-2" {
		t.Fatalf("winner message = %#v", sent)
	}
	if !repo.Lotteries["guild-1:"+lotteryTestID].Ended || len(repo.Ended) != 1 {
		t.Fatalf("repo = %#v", repo)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Content, "成功重抽") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestLotteryRerollCapsOversizedWinnerCount(t *testing.T) {
	repo := fakemongo.NewLotteryRepository()
	repo.Lotteries["guild-1:"+lotteryTestID] = domain.Lottery{
		GuildID:      "guild-1",
		ID:           lotteryTestID,
		OwnerID:      "user-1",
		WinnerCount:  legacyLotteryWinnerLimit + 1,
		ChannelID:    "channel-1",
		Participants: []domain.LotteryParticipant{{UserID: "user-1"}},
	}
	messages := fakediscord.NewSideEffects()
	module := NewComponentModule(repo, &fakebotinfo.DiscordInfoProvider{}, nil, messages, lotteryFixedClock{})
	draws := 0
	module.randomIndex = func(int) (int, error) {
		draws++
		return 0, nil
	}
	interaction := fakediscord.ComponentInteractionFromID(lotteryTestID + "restart")
	responder := fakediscord.NewResponder()

	if err := module.RerollHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("reroll: %v", err)
	}
	if draws != legacyLotteryWinnerLimit || len(messages.Sent) != 1 {
		t.Fatalf("draws=%d sent=%#v", draws, messages.Sent)
	}
	sent := messages.Sent[0].Message
	if strings.Count(sent.Content, "<@user-1>") != legacyLotteryWinnerLimit || strings.Count(sent.Embeds[0].Description, "<@user-1>") != legacyLotteryWinnerLimit || len(sent.AllowedMentions.UserIDs) != 1 {
		t.Fatalf("winner message = %#v", sent)
	}
	if !repo.Lotteries["guild-1:"+lotteryTestID].Ended || len(responder.Edits) != 1 {
		t.Fatalf("lottery=%#v edits=%#v", repo.Lotteries["guild-1:"+lotteryTestID], responder.Edits)
	}
}

func TestLotteryRerollPreservesLegacyEmptyParticipantMessage(t *testing.T) {
	repo := fakemongo.NewLotteryRepository()
	repo.Lotteries["guild-1:"+lotteryTestID] = domain.Lottery{
		GuildID:     "guild-1",
		ID:          lotteryTestID,
		OwnerID:     "user-1",
		WinnerCount: 1,
		ChannelID:   "channel-1",
	}
	info := &fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{BotDisplayColor: 0x123456}}
	messages := fakediscord.NewSideEffects()
	module := NewComponentModule(repo, info, nil, messages, lotteryFixedClock{})
	module.randomIndex = func(int) (int, error) {
		t.Fatal("random index called without participants")
		return 0, nil
	}
	interaction := fakediscord.ComponentInteractionFromID(lotteryTestID + "restart")
	responder := fakediscord.NewResponder()

	if err := module.RerollHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("reroll: %v", err)
	}
	if len(messages.Sent) != 1 {
		t.Fatalf("sent = %#v", messages.Sent)
	}
	sent := messages.Sent[0].Message
	if sent.Content != "<@>" || len(sent.Embeds) != 1 || sent.Embeds[0].Description != "**沒有人參加抽獎欸QQ**" || sent.Embeds[0].Color != 0x123456 {
		t.Fatalf("winner message = %#v", sent)
	}
	if !repo.Lotteries["guild-1:"+lotteryTestID].Ended || len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Content, "成功重抽") {
		t.Fatalf("lottery=%#v edits=%#v", repo.Lotteries["guild-1:"+lotteryTestID], responder.Edits)
	}
}

func TestLotteryRerollPreservesLegacyNonPositiveWinnerNoOp(t *testing.T) {
	for _, winnerCount := range []int{0, -1} {
		t.Run(strconv.Itoa(winnerCount), func(t *testing.T) {
			repo := fakemongo.NewLotteryRepository()
			repo.Lotteries["guild-1:"+lotteryTestID] = domain.Lottery{
				GuildID:      "guild-1",
				ID:           lotteryTestID,
				OwnerID:      "user-1",
				WinnerCount:  winnerCount,
				ChannelID:    "channel-1",
				Participants: []domain.LotteryParticipant{{UserID: "user-1"}},
			}
			messages := fakediscord.NewSideEffects()
			module := NewComponentModule(repo, &fakebotinfo.DiscordInfoProvider{}, nil, messages, lotteryFixedClock{})
			interaction := fakediscord.ComponentInteractionFromID(lotteryTestID + "restart")
			responder := fakediscord.NewResponder()

			if err := module.RerollHandler()(context.Background(), interaction, responder); err != nil {
				t.Fatalf("reroll: %v", err)
			}
			if len(responder.Defers) != 1 || !responder.Defers[0].Ephemeral || len(responder.Edits) != 0 || len(messages.Sent) != 0 || repo.Lotteries["guild-1:"+lotteryTestID].Ended {
				t.Fatalf("defers=%#v edits=%#v sent=%#v lottery=%#v", responder.Defers, responder.Edits, messages.Sent, repo.Lotteries["guild-1:"+lotteryTestID])
			}
		})
	}
}

func TestLotteryStopRequiresOwnerAndEndsLegacyOwnerlessRowsForModerator(t *testing.T) {
	repo := fakemongo.NewLotteryRepository()
	repo.Lotteries["guild-1:"+lotteryTestID] = domain.Lottery{GuildID: "guild-1", ID: lotteryTestID, OwnerID: "owner-1"}
	info := &fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{OwnerID: "guild-owner"}}
	module := NewComponentModule(repo, info, nil, nil, lotteryFixedClock{})
	interaction := fakediscord.ComponentInteractionFromID(lotteryTestID + "stop")
	interaction.Actor.PermissionBits = permissionManageMessages
	denied := fakediscord.NewResponder()

	if err := module.StopHandler()(context.Background(), interaction, denied); err != nil {
		t.Fatalf("stop denied: %v", err)
	}
	if len(denied.Edits) != 1 || !strings.Contains(denied.Edits[0].Content, "沒有權限") || repo.Lotteries["guild-1:"+lotteryTestID].Ended {
		t.Fatalf("denied edits=%#v lottery=%#v", denied.Edits, repo.Lotteries["guild-1:"+lotteryTestID])
	}

	legacyID := "1700000000000888lotter"
	repo.Lotteries["guild-1:"+legacyID] = domain.Lottery{GuildID: "guild-1", ID: legacyID}
	interaction.CustomID = legacyID + "stop"
	allowed := fakediscord.NewResponder()
	if err := module.StopHandler()(context.Background(), interaction, allowed); err != nil {
		t.Fatalf("stop legacy: %v", err)
	}
	if !repo.Lotteries["guild-1:"+legacyID].Ended || len(allowed.Edits) != 1 || allowed.Edits[0].Embeds[0].Title != "<a:green_tick:994529015652163614> | 成功取消此次抽獎!" {
		t.Fatalf("allowed edits=%#v lottery=%#v", allowed.Edits, repo.Lotteries["guild-1:"+legacyID])
	}
}

type lotteryFixedClock struct {
	now time.Time
}

func (c lotteryFixedClock) Now() time.Time {
	return c.now
}
