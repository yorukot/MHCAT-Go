package economy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestCoinGameHigherLowerAcceptShowsDrawBeforeDelayedSettlement(t *testing.T) {
	repo := coinGameTestRepository()
	clock := &coinGameTestClock{now: time.Unix(100, 0)}
	messages := fakediscord.NewSideEffects()
	module, scheduler := newCoinGameLifecycleTestModule(t, repo, messages, clock)
	module.gameRandInt = fixedCoinGameRandom(90, 10)
	module.color = func() int { return 0x123456 }

	start := coinGameSlash(domain.CoinGameKindHigherLower, "10")
	responder := fakediscord.NewResponder()
	if err := module.CoinGameHandler()(context.Background(), start, responder); err != nil {
		t.Fatalf("start game: %v", err)
	}
	if len(responder.Defers) != 1 || len(responder.Follow) != 1 {
		t.Fatalf("start response defers=%#v follow=%#v", responder.Defers, responder.Follow)
	}
	if responder.Follow[0].Components[0].Components[0].CustomID != "yesssss" {
		t.Fatalf("invite buttons = %#v", responder.Follow[0].Components)
	}

	accept := fakediscord.ComponentInteractionFromID("yesssss")
	accept.Actor = interactions.Actor{UserID: "user-2", Username: "Opponent", GuildID: "guild-1"}
	accept.MessageID = "message-1"
	responder = fakediscord.NewResponder()
	if err := module.CoinGameComponentHandler()(context.Background(), accept, responder); err != nil {
		t.Fatalf("accept game: %v", err)
	}
	if len(responder.Updates) != 1 || !strings.Contains(responder.Updates[0].Content, "正在為您隨機抽取數字") || len(responder.Updates[0].Embeds) != 0 || len(responder.Updates[0].Components) != 0 {
		t.Fatalf("drawing update = %#v", responder.Updates)
	}
	if len(responder.Follow) != 0 {
		t.Fatalf("higher/lower accept unexpectedly sent follow-up = %#v", responder.Follow)
	}
	session, ok := module.gameSessions.GetForComponent("guild-1", "user-1", "channel-1", "message-1")
	if !ok || session.Phase != coinGamePhaseHigherLowerDrawing || session.HigherLowerChallenger != 90 || session.HigherLowerOpponent != 10 {
		t.Fatalf("drawing session = %#v ok=%v", session, ok)
	}
	entry, ok := scheduler.Only()
	if !ok || entry.deadline != time.Unix(105, 0) || entry.generation != session.TurnGeneration {
		t.Fatalf("drawing schedule = %#v ok=%v", entry, ok)
	}
	assertCoinGameBalances(t, repo, 40, 40)
	if len(messages.Edited) != 0 {
		t.Fatalf("result rendered before delay: %#v", messages.Edited)
	}

	clock.Advance(coinGameResultDelay)
	if !scheduler.TriggerOnly() {
		t.Fatal("higher/lower result was not scheduled")
	}
	challenger, err := repo.GetCoinBalance(context.Background(), "guild-1", "user-1")
	if err != nil {
		t.Fatalf("get challenger balance: %v", err)
	}
	opponent, err := repo.GetCoinBalance(context.Background(), "guild-1", "user-2")
	if err != nil {
		t.Fatalf("get opponent balance: %v", err)
	}
	if challenger.Coins != 60 || opponent.Coins != 40 {
		t.Fatalf("balances challenger=%#v opponent=%#v", challenger, opponent)
	}
	if len(messages.Edited) != 1 || len(messages.Edited[0].Message.Embeds) != 1 || !strings.Contains(messages.Edited[0].Message.Embeds[0].Title, "比大小結果") || messages.Edited[0].Message.Content != "" {
		t.Fatalf("delayed result = %#v", messages.Edited)
	}
	if _, ok := module.gameSessions.GetForComponent("guild-1", "user-1", "channel-1", "message-1"); ok {
		t.Fatal("higher/lower session remained after settlement")
	}
}

func TestCoinGameChallengerAcceptDoesNotDebitPlayers(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 50})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-2", Coins: 50})
	module := NewCoinGameModule(repo, nil, nil, shopFixedClock{now: time.Unix(100, 0)})

	if err := module.CoinGameHandler()(context.Background(), coinGameSlash(domain.CoinGameKindHigherLower, "10"), fakediscord.NewResponder()); err != nil {
		t.Fatalf("start game: %v", err)
	}
	accept := fakediscord.ComponentInteractionFromID("yesssss")
	accept.Actor = interactions.Actor{UserID: "user-1", Username: "User", GuildID: "guild-1"}
	accept.MessageID = "message-1"
	responder := fakediscord.NewResponder()
	if err := module.CoinGameComponentHandler()(context.Background(), accept, responder); err != nil {
		t.Fatalf("challenger accept: %v", err)
	}
	if len(responder.Replies) != 1 || !strings.Contains(responder.Replies[0].Embeds[0].Title, "你不是被邀請者") {
		t.Fatalf("reply = %#v", responder.Replies)
	}
	challenger, _ := repo.GetCoinBalance(context.Background(), "guild-1", "user-1")
	opponent, _ := repo.GetCoinBalance(context.Background(), "guild-1", "user-2")
	if challenger.Coins != 50 || opponent.Coins != 50 {
		t.Fatalf("balances mutated challenger=%#v opponent=%#v", challenger, opponent)
	}
}

func TestCoinGameInviteAndTutorialUIMatchesLegacy(t *testing.T) {
	for _, test := range []struct {
		kind             domain.CoinGameKind
		tutorialID       string
		tutorialTitle    string
		tutorialSnippets []string
		componentCount   int
	}{
		{kind: domain.CoinGameKindKnowledge, componentCount: 2},
		{
			kind:             domain.CoinGameKindBlackjack,
			tutorialID:       "teach21point",
			tutorialTitle:    "<:creativeteaching:986060052949524600> 以下是21點介紹",
			tutorialSnippets: []string{"機器人自己發一張排給自己", "給遊玩的兩個人各兩張牌", "莊如果大於21點", "不會的話，玩玩看就知道ㄌ"},
			componentCount:   3,
		},
		{
			kind:             domain.CoinGameKindHigherLower,
			tutorialID:       "thansize",
			tutorialTitle:    "<:creativeteaching:986060052949524600> 以下為比大小介紹",
			tutorialSnippets: []string{"由機器人抽取兩位的數字(1-100)", "大的拿走所有賭注", "不會的話，玩玩看就知道ㄌ"},
			componentCount:   3,
		},
	} {
		t.Run(string(test.kind), func(t *testing.T) {
			repo := coinGameTestRepository()
			clock := &coinGameTestClock{now: time.Unix(100, 0)}
			module, _ := newCoinGameLifecycleTestModule(t, repo, fakediscord.NewSideEffects(), clock)
			start := coinGameSlash(test.kind, "10")
			start.ChannelID = "channel-1"
			responder := fakediscord.NewResponder()
			if err := module.CoinGameHandler()(context.Background(), start, responder); err != nil {
				t.Fatalf("start %s game: %v", test.kind, err)
			}
			if len(responder.Follow) != 1 || len(responder.Follow[0].Embeds) != 1 || responder.Follow[0].Embeds[0].Color != 0x123456 {
				t.Fatalf("%s invite = %#v", test.kind, responder.Follow)
			}
			invite := responder.Follow[0]
			components := invite.Components[0].Components
			if len(components) != test.componentCount {
				t.Fatalf("%s invite components = %#v", test.kind, components)
			}
			if test.tutorialID != "" {
				tutorialResponder := fakediscord.NewResponder()
				if err := module.CoinGameComponentHandler()(context.Background(), coinGameComponent(test.tutorialID, "user-2", "Opponent"), tutorialResponder); err != nil {
					t.Fatalf("%s tutorial: %v", test.kind, err)
				}
				if len(tutorialResponder.Replies) != 1 || !tutorialResponder.Replies[0].Ephemeral || len(tutorialResponder.Replies[0].Embeds) != 1 {
					t.Fatalf("%s tutorial response = %#v", test.kind, tutorialResponder.Replies)
				}
				embed := tutorialResponder.Replies[0].Embeds[0]
				if embed.Title != test.tutorialTitle || embed.Color != 0x123456 {
					t.Fatalf("%s tutorial embed = %#v", test.kind, embed)
				}
				for _, snippet := range test.tutorialSnippets {
					if !strings.Contains(embed.Description, snippet) {
						t.Fatalf("%s tutorial missing %q: %q", test.kind, snippet, embed.Description)
					}
				}
			}

			rejectResponder := fakediscord.NewResponder()
			if err := module.CoinGameComponentHandler()(context.Background(), coinGameComponent("nooooo", "user-2", "Opponent"), rejectResponder); err != nil {
				t.Fatalf("reject %s game: %v", test.kind, err)
			}
			if len(rejectResponder.Updates) != 1 || rejectResponder.Updates[0].Content != invite.Content || len(rejectResponder.Updates[0].Embeds) != 1 || rejectResponder.Updates[0].Embeds[0].Title != invite.Embeds[0].Title || rejectResponder.Updates[0].Embeds[0].Description != invite.Embeds[0].Description || rejectResponder.Updates[0].Embeds[0].Color != invite.Embeds[0].Color {
				t.Fatalf("%s reject update replaced invite = %#v", test.kind, rejectResponder.Updates)
			}
			disabled := rejectResponder.Updates[0].Components[0].Components
			if !disabled[0].Disabled || !disabled[1].Disabled {
				t.Fatalf("%s accept/reject buttons remained enabled: %#v", test.kind, disabled)
			}
			if len(disabled) == 3 && disabled[2].Disabled {
				t.Fatalf("%s tutorial button was disabled after rejection: %#v", test.kind, disabled[2])
			}
		})
	}
}

func TestCoinGameRejectsOpponentWithoutEnoughCoins(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 50})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-2", Coins: 1})
	module := NewCoinGameModule(repo, nil, nil, shopFixedClock{now: time.Unix(100, 0)})

	responder := fakediscord.NewResponder()
	if err := module.CoinGameHandler()(context.Background(), coinGameSlash(domain.CoinGameKindHigherLower, "10"), responder); err != nil {
		t.Fatalf("start game: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "對方沒有這麼多代幣") || responder.Edits[0].Embeds[0].Color != 0xFF0000 {
		t.Fatalf("insufficient response = %#v", responder.Edits)
	}
}

func TestCoinGameInviteExpiresAtLegacyDeadline(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 50})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-2", Coins: 50})
	clock := &coinGameTestClock{now: time.Unix(100, 0)}
	module := NewCoinGameModule(repo, nil, nil, clock)

	if err := module.CoinGameHandler()(context.Background(), coinGameSlash(domain.CoinGameKindHigherLower, "10"), fakediscord.NewResponder()); err != nil {
		t.Fatalf("start game: %v", err)
	}
	clock.Advance(coinGameInviteTTL)
	accept := fakediscord.ComponentInteractionFromID("yesssss")
	accept.Actor = interactions.Actor{UserID: "user-2", Username: "Opponent", GuildID: "guild-1"}
	accept.MessageID = "message-1"
	responder := fakediscord.NewResponder()
	if err := module.CoinGameComponentHandler()(context.Background(), accept, responder); err != nil {
		t.Fatalf("accept expired game: %v", err)
	}
	if len(responder.Replies) != 1 || !strings.Contains(responder.Replies[0].Embeds[0].Title, "找不到這場遊戲") {
		t.Fatalf("expired reply = %#v", responder.Replies)
	}
	challenger, _ := repo.GetCoinBalance(context.Background(), "guild-1", "user-1")
	opponent, _ := repo.GetCoinBalance(context.Background(), "guild-1", "user-2")
	if challenger.Coins != 50 || opponent.Coins != 50 {
		t.Fatalf("expired invite mutated balances challenger=%#v opponent=%#v", challenger, opponent)
	}
}

func TestCoinGameInviteCanBeAcceptedBeforeLegacyDeadline(t *testing.T) {
	repo := coinGameTestRepository()
	clock := &coinGameTestClock{now: time.Unix(100, 0)}
	module, scheduler := newCoinGameLifecycleTestModule(t, repo, fakediscord.NewSideEffects(), clock)
	module.gameRandInt = fixedCoinGameRandom(90, 10)

	if err := module.CoinGameHandler()(context.Background(), coinGameSlash(domain.CoinGameKindHigherLower, "10"), fakediscord.NewResponder()); err != nil {
		t.Fatalf("start game: %v", err)
	}
	clock.Advance(coinGameInviteTTL - time.Nanosecond)
	accept := fakediscord.ComponentInteractionFromID("yesssss")
	accept.Actor = interactions.Actor{UserID: "user-2", Username: "Opponent", GuildID: "guild-1"}
	accept.MessageID = "message-1"
	responder := fakediscord.NewResponder()
	if err := module.CoinGameComponentHandler()(context.Background(), accept, responder); err != nil {
		t.Fatalf("accept game: %v", err)
	}
	if len(responder.Updates) != 1 || !strings.Contains(responder.Updates[0].Content, "正在為您隨機抽取數字") || len(responder.Follow) != 0 || scheduler.Len() != 1 {
		t.Fatalf("accept update = %#v", responder.Updates)
	}
	assertCoinGameBalances(t, repo, 40, 40)
}

func TestCoinGameKnowledgeAcceptanceStartsAfterLegacyDelay(t *testing.T) {
	repo := coinGameTestRepository()
	clock := &coinGameTestClock{now: time.Unix(100, 0)}
	messages := fakediscord.NewSideEffects()
	module, scheduler := newCoinGameLifecycleTestModule(t, repo, messages, clock)

	start := coinGameSlash(domain.CoinGameKindKnowledge, "10")
	start.ChannelID = "channel-1"
	if err := module.CoinGameHandler()(context.Background(), start, fakediscord.NewResponder()); err != nil {
		t.Fatalf("start knowledge game: %v", err)
	}
	responder := fakediscord.NewResponder()
	if err := module.CoinGameComponentHandler()(context.Background(), coinGameComponent("yesssss", "user-2", "Opponent"), responder); err != nil {
		t.Fatalf("accept knowledge game: %v", err)
	}
	if len(responder.Replies) != 1 || !responder.Replies[0].Ephemeral || !strings.Contains(responder.Replies[0].Embeds[0].Title, "你成功接受了邀請") || len(responder.Updates) != 0 {
		t.Fatalf("knowledge accept response = replies %#v updates %#v", responder.Replies, responder.Updates)
	}
	session, ok := module.gameSessions.GetForComponent("guild-1", "user-1", "channel-1", "message-1")
	if !ok || session.Phase != coinGamePhaseKnowledgeStarting || session.KnowledgeQuestion.Question == "" || session.TurnDeadline != time.Unix(100, int64(coinGameKnowledgeStart)) {
		t.Fatalf("starting knowledge session = %#v ok=%v", session, ok)
	}
	entry, ok := scheduler.Only()
	if !ok || entry.deadline != session.TurnDeadline || entry.generation != session.TurnGeneration {
		t.Fatalf("knowledge start schedule = %#v ok=%v", entry, ok)
	}
	if len(messages.Edited) != 0 {
		t.Fatalf("knowledge question rendered before delay: %#v", messages.Edited)
	}

	clock.Advance(coinGameKnowledgeStart)
	if !scheduler.TriggerOnly() {
		t.Fatal("knowledge start transition was not scheduled")
	}
	if len(messages.Edited) != 1 || len(messages.Edited[0].Message.Components) != 1 || len(messages.Edited[0].Message.Components[0].Components) != 4 {
		t.Fatalf("initial knowledge question = %#v", messages.Edited)
	}
	message := messages.Edited[0].Message
	if !strings.Contains(message.Embeds[0].Title, "遊戲已開始") || !strings.Contains(message.Embeds[0].Description, "<t:116:R>") {
		t.Fatalf("initial knowledge UI = %#v", message)
	}
	session = currentCoinGameSession(t, module, session)
	if session.Phase != coinGamePhaseKnowledgeQuestion || session.QuestionShownAt != time.Unix(100, int64(coinGameKnowledgeStart)) || session.TurnDeadline != time.Unix(121, 0) {
		t.Fatalf("active knowledge session = %#v", session)
	}
	entry, ok = scheduler.Only()
	if !ok || entry.deadline != time.Unix(121, 0) || entry.generation != session.TurnGeneration {
		t.Fatalf("knowledge timeout schedule = %#v ok=%v", entry, ok)
	}
}

func TestCoinGameBlackjackAcceptanceFeedbackMatchesLegacy(t *testing.T) {
	repo := coinGameTestRepository()
	clock := &coinGameTestClock{now: time.Unix(100, 0)}
	module, scheduler := newCoinGameLifecycleTestModule(t, repo, fakediscord.NewSideEffects(), clock)

	start := coinGameSlash(domain.CoinGameKindBlackjack, "10")
	start.ChannelID = "channel-1"
	if err := module.CoinGameHandler()(context.Background(), start, fakediscord.NewResponder()); err != nil {
		t.Fatalf("start blackjack game: %v", err)
	}
	responder := fakediscord.NewResponder()
	if err := module.CoinGameComponentHandler()(context.Background(), coinGameComponent("yesssss", "user-2", "Opponent"), responder); err != nil {
		t.Fatalf("accept blackjack game: %v", err)
	}
	if len(responder.Updates) != 1 || !strings.Contains(responder.Updates[0].Embeds[0].Title, "遊戲已開始") {
		t.Fatalf("blackjack start update = %#v", responder.Updates)
	}
	if len(responder.Follow) != 1 || !responder.Follow[0].Ephemeral || !strings.Contains(responder.Follow[0].Content, "成功接受") {
		t.Fatalf("blackjack accept feedback = %#v", responder.Follow)
	}
	session, ok := module.gameSessions.GetForComponent("guild-1", "user-1", "channel-1", "message-1")
	if !ok || session.Phase != coinGamePhaseBlackjackTurn || session.TurnDeadline != time.Unix(131, 0) || scheduler.Len() != 1 {
		t.Fatalf("blackjack session = %#v ok=%v timers=%d", session, ok, scheduler.Len())
	}
}

func TestCoinGameBlackjackPostActionUIMatchesLegacy(t *testing.T) {
	repo := coinGameTestRepository()
	clock := &coinGameTestClock{now: time.Unix(100, 0)}
	module, scheduler := newCoinGameLifecycleTestModule(t, repo, fakediscord.NewSideEffects(), clock)
	session := acceptCoinGameForTest(t, module, domain.CoinGameKindBlackjack)

	challengerResponder := fakediscord.NewResponder()
	if err := module.CoinGameComponentHandler()(context.Background(), coinGameComponent("main_no_card", "user-1", "User"), challengerResponder); err != nil {
		t.Fatalf("challenger stand: %v", err)
	}
	if len(challengerResponder.Updates) != 1 || len(challengerResponder.Follow) != 1 {
		t.Fatalf("challenger action responses = updates %#v follow %#v", challengerResponder.Updates, challengerResponder.Follow)
	}
	challengerUpdate := challengerResponder.Updates[0]
	if challengerUpdate.Content != "<a:arrow_pink:996242460294512690> | **這回合是<@user-2>的，另一位只能查看牌組喔!**" || challengerUpdate.Embeds[0].Title != "<:startbutton1:1005838813274325022> 21點小遊戲" {
		t.Fatalf("challenger action update = %#v", challengerUpdate)
	}
	if !strings.Contains(challengerUpdate.Embeds[0].Description, "<@user-1>**選擇了:**`略過\n`**") || strings.Contains(challengerUpdate.Embeds[0].Description, "已為各位各發一張牌") || challengerUpdate.Components[0].Components[0].CustomID != "user_no_card" {
		t.Fatalf("challenger action description/components = %#v", challengerUpdate)
	}
	if challengerResponder.Follow[0].Embeds[0].Title != "<a:green_tick:994529015652163614> | 你選擇了略過" || challengerResponder.Follow[0].Embeds[0].Color != 0x123456 {
		t.Fatalf("challenger action feedback = %#v", challengerResponder.Follow[0])
	}
	session = currentCoinGameSession(t, module, session)
	if timeout := coinGameTimeoutOutbound(session); timeout.Content != challengerUpdate.Content {
		t.Fatalf("opponent-turn timeout content = %q want %q", timeout.Content, challengerUpdate.Content)
	}

	opponentResponder := fakediscord.NewResponder()
	if err := module.CoinGameComponentHandler()(context.Background(), coinGameComponent("user_get_card", "user-2", "Opponent"), opponentResponder); err != nil {
		t.Fatalf("opponent hit: %v", err)
	}
	if len(opponentResponder.Updates) != 1 || len(opponentResponder.Follow) != 1 {
		t.Fatalf("opponent action responses = updates %#v follow %#v", opponentResponder.Updates, opponentResponder.Follow)
	}
	opponentUpdate := opponentResponder.Updates[0]
	if opponentUpdate.Content != "<a:arrow_pink:996242460294512690> | **這回合是<@user-1>的，另一位只能查看牌組喔!**" || !strings.Contains(opponentUpdate.Embeds[0].Description, "<@user-2>**選擇了:**`抽牌\n`**") || opponentUpdate.Components[0].Components[0].CustomID != "main_no_card" {
		t.Fatalf("opponent action update = %#v", opponentUpdate)
	}
	if !strings.Contains(opponentResponder.Follow[0].Embeds[0].Title, "你抽到了:") || opponentResponder.Follow[0].Embeds[0].Color != 0x123456 {
		t.Fatalf("opponent action feedback = %#v", opponentResponder.Follow[0])
	}
	session = currentCoinGameSession(t, module, session)
	if timeout := coinGameTimeoutOutbound(session); timeout.Content != opponentUpdate.Content {
		t.Fatalf("challenger-turn timeout content = %q want %q", timeout.Content, opponentUpdate.Content)
	}

	cardResponder := fakediscord.NewResponder()
	if err := module.CoinGameComponentHandler()(context.Background(), coinGameComponent("lookmenumber", "user-2", "Opponent"), cardResponder); err != nil {
		t.Fatalf("show opponent cards: %v", err)
	}
	if len(cardResponder.Replies) != 1 || !strings.Contains(cardResponder.Replies[0].Embeds[0].Description, ", ") || cardResponder.Replies[0].Embeds[0].Color != 0x123456 {
		t.Fatalf("private card list = %#v", cardResponder.Replies)
	}
	if scheduler.Len() != 1 {
		t.Fatalf("blackjack action timer count = %d", scheduler.Len())
	}
}

func TestCoinGameKnowledgeDuplicateAnswerUsesLegacyErrorUI(t *testing.T) {
	repo := coinGameTestRepository()
	clock := &coinGameTestClock{now: time.Unix(100, 0)}
	module, scheduler := newCoinGameLifecycleTestModule(t, repo, fakediscord.NewSideEffects(), clock)
	session := startKnowledgeQuestionForTest(t, module, scheduler, clock)
	answer := session.KnowledgeQuestion.Answer
	if err := module.CoinGameComponentHandler()(context.Background(), coinGameComponent(answer, "user-1", "User"), fakediscord.NewResponder()); err != nil {
		t.Fatalf("first knowledge answer: %v", err)
	}

	responder := fakediscord.NewResponder()
	if err := module.CoinGameComponentHandler()(context.Background(), coinGameComponent(answer, "user-1", "User"), responder); err != nil {
		t.Fatalf("duplicate knowledge answer: %v", err)
	}
	if len(responder.Replies) != 1 || !responder.Replies[0].Ephemeral || responder.Replies[0].Embeds[0].Title != "<a:error:980086028113182730> | 你已經選取過了!!!" || responder.Replies[0].Embeds[0].Color != 0xEA0000 {
		t.Fatalf("duplicate answer response = %#v", responder.Replies)
	}
}

func TestCoinGameKnowledgeTimeoutUsesLegacyTickAndPaysOnlyResponder(t *testing.T) {
	repo := coinGameTestRepository()
	clock := &coinGameTestClock{now: time.Unix(100, 0)}
	messages := fakediscord.NewSideEffects()
	module, scheduler := newCoinGameLifecycleTestModule(t, repo, messages, clock)
	session := startKnowledgeQuestionForTest(t, module, scheduler, clock)
	if currentCoinGameSession(t, module, session).KnowledgeQuestion.Question == "" {
		t.Fatal("knowledge question was not initialized")
	}
	session = currentCoinGameSession(t, module, session)
	if session.TurnDeadline != time.Unix(121, 0) {
		t.Fatalf("knowledge deadline = %v", session.TurnDeadline)
	}
	if message := knowledgeQuestionMessage(session, 0); !strings.Contains(message.Embeds[0].Description, "<t:116:R>") {
		t.Fatalf("knowledge deadline message = %q", message.Embeds[0].Description)
	}

	answer := coinGameComponent(session.KnowledgeQuestion.Answer, "user-1", "User")
	if err := module.CoinGameComponentHandler()(context.Background(), answer, fakediscord.NewResponder()); err != nil {
		t.Fatalf("answer knowledge question: %v", err)
	}
	clock.Advance(20 * time.Second)
	if !scheduler.TriggerOnly() {
		t.Fatal("knowledge timer was not scheduled")
	}
	assertCoinGameBalances(t, repo, 40, 40)
	if len(messages.Edited) != 1 {
		t.Fatalf("timeout fired before strict legacy tick: %#v", messages.Edited)
	}

	clock.Advance(time.Second)
	if !scheduler.TriggerOnly() {
		t.Fatal("knowledge timeout retry was not scheduled")
	}
	assertCoinGameBalances(t, repo, 60, 40)
	if len(messages.Edited) != 2 {
		t.Fatalf("timeout edits = %#v", messages.Edited)
	}
	edit := messages.Edited[1]
	if edit.Ref != (ports.MessageRef{ChannelID: "channel-1", MessageID: "message-1"}) {
		t.Fatalf("timeout ref = %#v", edit.Ref)
	}
	if len(edit.Message.Components) != 0 || len(edit.Message.Embeds) != 1 || !strings.Contains(edit.Message.Embeds[0].Title, "Opponent") || strings.Contains(edit.Message.Embeds[0].Title, "User ") {
		t.Fatalf("timeout message = %#v", edit.Message)
	}
}

func TestCoinGameKnowledgeTimeoutBurnsPotWhenNeitherPlayerAnswers(t *testing.T) {
	repo := coinGameTestRepository()
	clock := &coinGameTestClock{now: time.Unix(100, 0)}
	messages := fakediscord.NewSideEffects()
	module, scheduler := newCoinGameLifecycleTestModule(t, repo, messages, clock)
	session := startKnowledgeQuestionForTest(t, module, scheduler, clock)
	clock.Advance(session.TurnDeadline.Sub(clock.Now()))
	if !scheduler.TriggerOnly() {
		t.Fatal("knowledge timer was not scheduled")
	}
	assertCoinGameBalances(t, repo, 40, 40)
	if len(messages.Edited) != 2 {
		t.Fatalf("timeout edits = %#v", messages.Edited)
	}
	title := messages.Edited[1].Message.Embeds[0].Title
	if !strings.Contains(title, "Opponent User") {
		t.Fatalf("timeout title = %q", title)
	}
}

func TestCoinGameKnowledgeRevealWaitsFiveSecondsAndCarriesCountdown(t *testing.T) {
	repo := coinGameTestRepository()
	clock := &coinGameTestClock{now: time.Unix(100, 0)}
	messages := fakediscord.NewSideEffects()
	module, scheduler := newCoinGameLifecycleTestModule(t, repo, messages, clock)
	session := startKnowledgeQuestionForTest(t, module, scheduler, clock)
	firstResponder := fakediscord.NewResponder()
	if err := module.CoinGameComponentHandler()(context.Background(), coinGameComponent(session.KnowledgeQuestion.Answer, "user-1", "User"), firstResponder); err != nil {
		t.Fatalf("challenger answer: %v", err)
	}
	if len(firstResponder.Replies) != 1 || !strings.Contains(firstResponder.Replies[0].Embeds[0].Description, "`1000`分") {
		t.Fatalf("challenger answer response = %#v", firstResponder.Replies)
	}
	clock.Advance(2 * time.Second)
	wrongAnswer := knowledgeWrongAnswer(session)
	secondResponder := fakediscord.NewResponder()
	if err := module.CoinGameComponentHandler()(context.Background(), coinGameComponent(wrongAnswer, "user-2", "Opponent"), secondResponder); err != nil {
		t.Fatalf("opponent answer: %v", err)
	}
	if len(secondResponder.Updates) != 1 || len(secondResponder.Follow) != 1 || !secondResponder.Follow[0].Ephemeral {
		t.Fatalf("second answer responses = updates %#v follow %#v", secondResponder.Updates, secondResponder.Follow)
	}
	reveal := secondResponder.Updates[0]
	if len(reveal.Components) != 0 || len(reveal.Embeds) != 1 {
		t.Fatalf("knowledge reveal components = %#v", reveal)
	}
	description := reveal.Embeds[0].Description
	for _, want := range []string{session.KnowledgeQuestion.Answer, wrongAnswer, "正確答案", "目前得分", "還剩下**`4`**題", "<a:green_tick:994529015652163614>", "<a:Discord_AnimatedNo:1015989839809757295>"} {
		if !strings.Contains(description, want) {
			t.Fatalf("knowledge reveal missing %q: %q", want, description)
		}
	}
	session = currentCoinGameSession(t, module, session)
	if session.Phase != coinGamePhaseKnowledgeReveal || session.KnowledgeRound != 1 || session.QuestionStartedAt != time.Unix(102, int64(coinGameKnowledgeStart)) || session.TurnDeadline != time.Unix(107, int64(coinGameKnowledgeStart)) {
		t.Fatalf("knowledge reveal session = %#v", session)
	}
	transition, ok := scheduler.Only()
	if !ok || transition.deadline != session.TurnDeadline || transition.generation != session.TurnGeneration {
		t.Fatalf("knowledge reveal schedule = %#v ok=%v", transition, ok)
	}
	assertCoinGameBalances(t, repo, 40, 40)
	if len(messages.Edited) != 1 {
		t.Fatalf("next question rendered before reveal delay: %#v", messages.Edited)
	}

	staleResponder := fakediscord.NewResponder()
	if err := module.CoinGameComponentHandler()(context.Background(), coinGameComponent(session.KnowledgeQuestion.Answer, "user-1", "User"), staleResponder); err != nil {
		t.Fatalf("stale reveal answer: %v", err)
	}
	if len(staleResponder.Replies) != 1 || !strings.Contains(staleResponder.Replies[0].Embeds[0].Title, "還沒開始") {
		t.Fatalf("stale reveal response = %#v", staleResponder.Replies)
	}

	clock.Advance(coinGameResultDelay)
	if !scheduler.TriggerOnly() {
		t.Fatal("next knowledge question was not scheduled")
	}
	if len(messages.Edited) != 2 {
		t.Fatalf("next knowledge question edits = %#v", messages.Edited)
	}
	next := messages.Edited[1].Message
	if len(next.Components) != 1 || len(next.Components[0].Components) != 4 || !strings.Contains(next.Embeds[0].Title, "知識王") || strings.Contains(next.Embeds[0].Title, "遊戲已開始") || !strings.Contains(next.Embeds[0].Description, "<t:123:R>") {
		t.Fatalf("next knowledge question = %#v", next)
	}
	session = currentCoinGameSession(t, module, session)
	if session.Phase != coinGamePhaseKnowledgeQuestion || session.KnowledgeQuestion.Question == "" || session.QuestionShownAt != time.Unix(107, int64(coinGameKnowledgeStart)) || session.TurnDeadline != time.Unix(123, int64(coinGameKnowledgeStart)) {
		t.Fatalf("next knowledge session = %#v", session)
	}
	if points := module.knowledgePoints(session); points != 750 {
		t.Fatalf("next-question starting points = %d, want 750", points)
	}

	transition.handler(context.Background())
	if len(messages.Edited) != 2 || scheduler.Len() != 1 {
		t.Fatalf("stale transition changed lifecycle: edits=%#v timers=%d", messages.Edited, scheduler.Len())
	}
}

func TestCoinGameKnowledgeFinalSettlementWaitsForFifthReveal(t *testing.T) {
	repo := coinGameTestRepository()
	clock := &coinGameTestClock{now: time.Unix(100, 0)}
	messages := fakediscord.NewSideEffects()
	module, scheduler := newCoinGameLifecycleTestModule(t, repo, messages, clock)
	session := startKnowledgeQuestionForTest(t, module, scheduler, clock)

	for round := 1; round <= 5; round++ {
		session = currentCoinGameSession(t, module, session)
		answer := session.KnowledgeQuestion.Answer
		if err := module.CoinGameComponentHandler()(context.Background(), coinGameComponent(answer, "user-1", "User"), fakediscord.NewResponder()); err != nil {
			t.Fatalf("round %d challenger answer: %v", round, err)
		}
		responder := fakediscord.NewResponder()
		if err := module.CoinGameComponentHandler()(context.Background(), coinGameComponent(answer, "user-2", "Opponent"), responder); err != nil {
			t.Fatalf("round %d opponent answer: %v", round, err)
		}
		if len(responder.Updates) != 1 || !strings.Contains(responder.Updates[0].Embeds[0].Description, fmt.Sprintf("還剩下**`%d`**題", 5-round)) {
			t.Fatalf("round %d reveal = %#v", round, responder.Updates)
		}
		if !strings.Contains(responder.Updates[0].Embeds[0].Description, "<a:error:980086028113182730>") {
			t.Fatalf("round %d correct-answer reveal lost legacy marker alternation: %q", round, responder.Updates[0].Embeds[0].Description)
		}
		session = currentCoinGameSession(t, module, session)
		if session.Phase != coinGamePhaseKnowledgeReveal || session.KnowledgeRound != round {
			t.Fatalf("round %d reveal session = %#v", round, session)
		}
		assertCoinGameBalances(t, repo, 40, 40)
		clock.Advance(coinGameResultDelay)
		if !scheduler.TriggerOnly() {
			t.Fatalf("round %d transition was not scheduled", round)
		}
		if round < 5 {
			session = currentCoinGameSession(t, module, session)
			continue
		}
	}

	assertCoinGameBalances(t, repo, 50, 50)
	if _, ok := module.gameSessions.GetForComponent("guild-1", "user-1", "channel-1", "message-1"); ok {
		t.Fatal("knowledge session remained after final settlement")
	}
	if scheduler.Len() != 0 || len(messages.Edited) != 6 || !strings.Contains(messages.Edited[5].Message.Embeds[0].Title, "遊戲已結束") || len(messages.Edited[5].Message.Components) != 0 {
		t.Fatalf("final knowledge lifecycle timers=%d edits=%#v", scheduler.Len(), messages.Edited)
	}
}

func TestCoinGameBlackjackTimeoutPaysOtherPlayerForEachTurn(t *testing.T) {
	for _, test := range []struct {
		name           string
		moveToOpponent bool
		wantChallenger int64
		wantOpponent   int64
		wantTimedOut   string
	}{
		{name: "challenger turn", wantChallenger: 40, wantOpponent: 60, wantTimedOut: "User"},
		{name: "opponent turn", moveToOpponent: true, wantChallenger: 60, wantOpponent: 40, wantTimedOut: "Opponent"},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := coinGameTestRepository()
			clock := &coinGameTestClock{now: time.Unix(100, 0)}
			messages := fakediscord.NewSideEffects()
			module, scheduler := newCoinGameLifecycleTestModule(t, repo, messages, clock)
			session := acceptCoinGameForTest(t, module, domain.CoinGameKindBlackjack)
			if test.moveToOpponent {
				action := coinGameComponent("main_get_card", "user-1", "User")
				if err := module.CoinGameComponentHandler()(context.Background(), action, fakediscord.NewResponder()); err != nil {
					t.Fatalf("move to opponent turn: %v", err)
				}
				session = currentCoinGameSession(t, module, session)
			}
			if session.TurnDeadline != time.Unix(131, 0) {
				t.Fatalf("blackjack deadline = %v", session.TurnDeadline)
			}
			if message := blackjackTurnMessage(session, session.BlackjackTurn == session.ChallengerID, 0); !strings.Contains(message.Embeds[0].Description, "<t:130:R>") {
				t.Fatalf("blackjack deadline message = %q", message.Embeds[0].Description)
			}
			clock.Advance(coinGameBlackjackTTL)
			if !scheduler.TriggerOnly() {
				t.Fatal("blackjack timer was not scheduled")
			}
			assertCoinGameBalances(t, repo, test.wantChallenger, test.wantOpponent)
			if len(messages.Edited) != 1 || !strings.Contains(messages.Edited[0].Message.Embeds[0].Title, "`"+test.wantTimedOut+"`") || len(messages.Edited[0].Message.Components) != 0 {
				t.Fatalf("timeout edit = %#v", messages.Edited)
			}
		})
	}
}

func TestCoinGameTerminalActionExcludesConcurrentTimeoutSettlement(t *testing.T) {
	base := coinGameTestRepository()
	release := make(chan struct{})
	repository := &blockingCoinGameRepository{
		next:          base,
		settleStarted: make(chan struct{}),
		release:       release,
	}
	clock := &coinGameTestClock{now: time.Unix(100, 0)}
	module, scheduler := newCoinGameLifecycleTestModule(t, repository, fakediscord.NewSideEffects(), clock)
	acceptCoinGameForTest(t, module, domain.CoinGameKindBlackjack)
	if err := module.CoinGameComponentHandler()(context.Background(), coinGameComponent("main_no_card", "user-1", "User"), fakediscord.NewResponder()); err != nil {
		t.Fatalf("challenger stand: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- module.CoinGameComponentHandler()(context.Background(), coinGameComponent("user_no_card", "user-2", "Opponent"), fakediscord.NewResponder())
	}()
	select {
	case <-repository.settleStarted:
	case <-time.After(time.Second):
		t.Fatal("terminal settlement did not start")
	}
	clock.Advance(coinGameBlackjackTTL)
	if !scheduler.TriggerOnly() {
		t.Fatal("blackjack timer was not scheduled")
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatalf("terminal action: %v", err)
	}
	if calls := repository.SettleCalls(); calls != 1 {
		t.Fatalf("settlement calls = %d", calls)
	}
	challenger, _ := base.GetCoinBalance(context.Background(), "guild-1", "user-1")
	opponent, _ := base.GetCoinBalance(context.Background(), "guild-1", "user-2")
	if challenger.Coins+opponent.Coins != 100 {
		t.Fatalf("settled balance total = %d (%d/%d)", challenger.Coins+opponent.Coins, challenger.Coins, opponent.Coins)
	}
	if scheduler.Len() != 0 {
		t.Fatalf("timeout remained scheduled after terminal action: %d", scheduler.Len())
	}
}

func TestCoinGameHigherLowerDelayedSettlementFailureFailsClosed(t *testing.T) {
	base := coinGameTestRepository()
	repository := &blockingCoinGameRepository{next: base, settleErr: errors.New("settlement failed")}
	clock := &coinGameTestClock{now: time.Unix(100, 0)}
	messages := fakediscord.NewSideEffects()
	module, scheduler := newCoinGameLifecycleTestModule(t, repository, messages, clock)
	module.gameRandInt = fixedCoinGameRandom(90, 10)
	acceptCoinGameForTest(t, module, domain.CoinGameKindHigherLower)
	clock.Advance(coinGameResultDelay)
	if !scheduler.TriggerOnly() {
		t.Fatal("higher/lower result was not scheduled")
	}
	assertCoinGameBalances(t, base, 40, 40)
	if repository.SettleCalls() != 1 || scheduler.Len() != 0 || len(messages.Edited) != 1 || !strings.Contains(messages.Edited[0].Message.Embeds[0].Title, "未知的錯誤") {
		t.Fatalf("failed delayed settlement calls=%d timers=%d edits=%#v", repository.SettleCalls(), scheduler.Len(), messages.Edited)
	}
	responder := fakediscord.NewResponder()
	if err := module.CoinGameComponentHandler()(context.Background(), coinGameComponent("yesssss", "user-2", "Opponent"), responder); err != nil {
		t.Fatalf("component after failed delayed settlement: %v", err)
	}
	if len(responder.Replies) != 1 || !strings.Contains(responder.Replies[0].Embeds[0].Title, "找不到這場遊戲") {
		t.Fatalf("fail-closed reply = %#v", responder.Replies)
	}
}

func TestCoinGameTimeoutSettlementFailureFailsClosed(t *testing.T) {
	base := coinGameTestRepository()
	repository := &blockingCoinGameRepository{next: base, settleErr: errors.New("settlement failed")}
	clock := &coinGameTestClock{now: time.Unix(100, 0)}
	messages := fakediscord.NewSideEffects()
	module, scheduler := newCoinGameLifecycleTestModule(t, repository, messages, clock)
	session := startKnowledgeQuestionForTest(t, module, scheduler, clock)
	clock.Advance(session.TurnDeadline.Sub(clock.Now()))
	if !scheduler.TriggerOnly() {
		t.Fatal("knowledge timer was not scheduled")
	}
	assertCoinGameBalances(t, base, 40, 40)
	if len(messages.Edited) != 1 || scheduler.Len() != 0 || repository.SettleCalls() != 1 {
		t.Fatalf("failure state edits=%#v timers=%d settlements=%d", messages.Edited, scheduler.Len(), repository.SettleCalls())
	}
	responder := fakediscord.NewResponder()
	if err := module.CoinGameComponentHandler()(context.Background(), coinGameComponent("answer", "user-1", "User"), responder); err != nil {
		t.Fatalf("component after failed settlement: %v", err)
	}
	if len(responder.Replies) != 1 || !strings.Contains(responder.Replies[0].Embeds[0].Title, "找不到這場遊戲") {
		t.Fatalf("fail-closed reply = %#v", responder.Replies)
	}
}

func TestCoinGameUnknownReserveFailureFailsClosed(t *testing.T) {
	base := coinGameTestRepository()
	repository := &blockingCoinGameRepository{next: base, reserveErr: errors.New("reserve outcome unknown")}
	clock := &coinGameTestClock{now: time.Unix(100, 0)}
	module, scheduler := newCoinGameLifecycleTestModule(t, repository, fakediscord.NewSideEffects(), clock)
	start := coinGameSlash(domain.CoinGameKindKnowledge, "10")
	start.ChannelID = "channel-1"
	if err := module.CoinGameHandler()(context.Background(), start, fakediscord.NewResponder()); err != nil {
		t.Fatalf("start knowledge game: %v", err)
	}
	responder := fakediscord.NewResponder()
	if err := module.CoinGameComponentHandler()(context.Background(), coinGameComponent("yesssss", "user-2", "Opponent"), responder); err != nil {
		t.Fatalf("accept with failed reserve: %v", err)
	}
	if len(responder.Updates) != 1 || !strings.Contains(responder.Updates[0].Embeds[0].Title, "未知的錯誤") {
		t.Fatalf("reserve failure response = %#v", responder.Updates)
	}
	responder = fakediscord.NewResponder()
	if err := module.CoinGameComponentHandler()(context.Background(), coinGameComponent("yesssss", "user-2", "Opponent"), responder); err != nil {
		t.Fatalf("second accept: %v", err)
	}
	if len(responder.Replies) != 1 || !strings.Contains(responder.Replies[0].Embeds[0].Title, "找不到這場遊戲") {
		t.Fatalf("fail-closed reserve reply = %#v", responder.Replies)
	}
	assertCoinGameBalances(t, base, 50, 50)
	if repository.ReserveCalls() != 1 || scheduler.Len() != 0 {
		t.Fatalf("reserve calls=%d timers=%d", repository.ReserveCalls(), scheduler.Len())
	}
}

func coinGameSlash(kind domain.CoinGameKind, wager string) interactions.Interaction {
	return fakediscord.SlashInteractionWithOptions(CoinGameCommandName, string(kind), map[string]string{
		coinGameOptionOpponent: "user-2",
		coinGameOptionWager:    wager,
	})
}

func fixedCoinGameRandom(values ...int) func(int) int {
	index := 0
	return func(max int) int {
		if len(values) == 0 {
			return 0
		}
		value := values[index%len(values)]
		index++
		if max <= 0 {
			return value
		}
		return value % max
	}
}

type coinGameTestClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *coinGameTestClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *coinGameTestClock) Advance(duration time.Duration) {
	c.mu.Lock()
	c.now = c.now.Add(duration)
	c.mu.Unlock()
}

type manualCoinGameTimeout struct {
	generation uint64
	deadline   time.Time
	handler    func(context.Context)
}

type manualCoinGameTimeoutScheduler struct {
	mu      sync.Mutex
	entries map[string]manualCoinGameTimeout
	stopped bool
}

func newManualCoinGameTimeoutScheduler() *manualCoinGameTimeoutScheduler {
	return &manualCoinGameTimeoutScheduler{entries: map[string]manualCoinGameTimeout{}}
}

func (s *manualCoinGameTimeoutScheduler) Schedule(key string, generation uint64, deadline time.Time, handler func(context.Context)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopped {
		return
	}
	if current, ok := s.entries[key]; ok && current.generation > generation {
		return
	}
	s.entries[key] = manualCoinGameTimeout{generation: generation, deadline: deadline, handler: handler}
}

func (s *manualCoinGameTimeoutScheduler) Cancel(key string) {
	s.mu.Lock()
	delete(s.entries, key)
	s.mu.Unlock()
}

func (s *manualCoinGameTimeoutScheduler) Stop(context.Context) error {
	s.mu.Lock()
	s.stopped = true
	s.entries = map[string]manualCoinGameTimeout{}
	s.mu.Unlock()
	return nil
}

func (s *manualCoinGameTimeoutScheduler) TriggerOnly() bool {
	s.mu.Lock()
	if len(s.entries) != 1 {
		s.mu.Unlock()
		return false
	}
	var key string
	var entry manualCoinGameTimeout
	for key, entry = range s.entries {
	}
	delete(s.entries, key)
	s.mu.Unlock()
	entry.handler(context.Background())
	return true
}

func (s *manualCoinGameTimeoutScheduler) Only() (manualCoinGameTimeout, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.entries) != 1 {
		return manualCoinGameTimeout{}, false
	}
	for _, entry := range s.entries {
		return entry, true
	}
	return manualCoinGameTimeout{}, false
}

func (s *manualCoinGameTimeoutScheduler) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.entries)
}

func newCoinGameLifecycleTestModule(t *testing.T, repository ports.EconomyCoinGameRepository, messages ports.DiscordMessagePort, clock *coinGameTestClock) (Module, *manualCoinGameTimeoutScheduler) {
	t.Helper()
	module := NewCoinGameModuleWithMessages(repository, nil, messages, nil, clock).
		WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil)))
	_ = module.gameTimeouts.Stop(context.Background())
	scheduler := newManualCoinGameTimeoutScheduler()
	module.gameTimeouts = scheduler
	module.gameRandInt = fixedCoinGameRandom(0, 1, 2, 3, 4, 5, 6, 7, 8, 9)
	module.color = func() int { return 0x123456 }
	t.Cleanup(func() { _ = scheduler.Stop(context.Background()) })
	return module, scheduler
}

func coinGameTestRepository() *fakemongo.EconomyRepository {
	repository := fakemongo.NewEconomyRepository()
	repository.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 50})
	repository.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-2", Coins: 50})
	return repository
}

func acceptCoinGameForTest(t *testing.T, module Module, kind domain.CoinGameKind) coinGameSession {
	t.Helper()
	start := coinGameSlash(kind, "10")
	start.ChannelID = "channel-1"
	if err := module.CoinGameHandler()(context.Background(), start, fakediscord.NewResponder()); err != nil {
		t.Fatalf("start %s game: %v", kind, err)
	}
	accept := coinGameComponent("yesssss", "user-2", "Opponent")
	responder := fakediscord.NewResponder()
	if err := module.CoinGameComponentHandler()(context.Background(), accept, responder); err != nil {
		t.Fatalf("accept %s game: %v", kind, err)
	}
	session, ok := module.gameSessions.GetForComponent("guild-1", "user-1", "channel-1", "message-1")
	if !ok {
		t.Fatalf("accepted %s session missing", kind)
	}
	return session
}

func startKnowledgeQuestionForTest(t *testing.T, module Module, scheduler *manualCoinGameTimeoutScheduler, clock *coinGameTestClock) coinGameSession {
	t.Helper()
	session := acceptCoinGameForTest(t, module, domain.CoinGameKindKnowledge)
	if session.Phase != coinGamePhaseKnowledgeStarting {
		t.Fatalf("accepted knowledge phase = %q", session.Phase)
	}
	clock.Advance(coinGameKnowledgeStart)
	if !scheduler.TriggerOnly() {
		t.Fatal("knowledge start transition was not scheduled")
	}
	return currentCoinGameSession(t, module, session)
}

func currentCoinGameSession(t *testing.T, module Module, previous coinGameSession) coinGameSession {
	t.Helper()
	session, ok := module.gameSessions.GetForComponent(previous.GuildID, previous.ChallengerID, previous.ChannelID, previous.MessageID)
	if !ok {
		t.Fatal("coin game session missing")
	}
	return session
}

func coinGameComponent(customID string, userID string, username string) interactions.Interaction {
	interaction := fakediscord.ComponentInteractionFromID(customID)
	interaction.Actor = interactions.Actor{UserID: userID, Username: username, GuildID: "guild-1"}
	interaction.ChannelID = "channel-1"
	interaction.MessageID = "message-1"
	return interaction
}

func knowledgeWrongAnswer(session coinGameSession) string {
	for _, answer := range session.KnowledgeAnswers {
		if answer != session.KnowledgeQuestion.Answer {
			return answer
		}
	}
	return "wrong answer"
}

func assertCoinGameBalances(t *testing.T, repository *fakemongo.EconomyRepository, challenger int64, opponent int64) {
	t.Helper()
	challengerBalance, err := repository.GetCoinBalance(context.Background(), "guild-1", "user-1")
	if err != nil {
		t.Fatalf("get challenger balance: %v", err)
	}
	opponentBalance, err := repository.GetCoinBalance(context.Background(), "guild-1", "user-2")
	if err != nil {
		t.Fatalf("get opponent balance: %v", err)
	}
	if challengerBalance.Coins != challenger || opponentBalance.Coins != opponent {
		t.Fatalf("balances=%d/%d want=%d/%d", challengerBalance.Coins, opponentBalance.Coins, challenger, opponent)
	}
}

type blockingCoinGameRepository struct {
	next          ports.EconomyCoinGameRepository
	settleStarted chan struct{}
	release       <-chan struct{}
	reserveErr    error
	settleErr     error
	startOnce     sync.Once
	mu            sync.Mutex
	reserveCalls  int
	settleCalls   int
}

func (r *blockingCoinGameRepository) CheckCoinGameBalances(ctx context.Context, command domain.CoinGameCommand) (domain.CoinGameBalanceResult, error) {
	return r.next.CheckCoinGameBalances(ctx, command)
}

func (r *blockingCoinGameRepository) ReserveCoinGameWager(ctx context.Context, command domain.CoinGameCommand) (domain.CoinGameBalanceResult, error) {
	r.mu.Lock()
	r.reserveCalls++
	r.mu.Unlock()
	if r.reserveErr != nil {
		return domain.CoinGameBalanceResult{}, r.reserveErr
	}
	return r.next.ReserveCoinGameWager(ctx, command)
}

func (r *blockingCoinGameRepository) SettleCoinGameWager(ctx context.Context, command domain.CoinGameSettlementCommand) (domain.CoinGameSettlementResult, error) {
	r.mu.Lock()
	r.settleCalls++
	r.mu.Unlock()
	if r.settleStarted != nil {
		r.startOnce.Do(func() { close(r.settleStarted) })
	}
	if r.release != nil {
		select {
		case <-r.release:
		case <-ctx.Done():
			return domain.CoinGameSettlementResult{}, ctx.Err()
		}
	}
	if r.settleErr != nil {
		return domain.CoinGameSettlementResult{}, r.settleErr
	}
	return r.next.SettleCoinGameWager(ctx, command)
}

func (r *blockingCoinGameRepository) SettleCalls() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.settleCalls
}

func (r *blockingCoinGameRepository) ReserveCalls() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.reserveCalls
}
