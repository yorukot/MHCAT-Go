package economy

import (
	"context"
	"errors"
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

func TestCoinGameHigherLowerAcceptSettlesPot(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 50})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-2", Coins: 50})
	module := NewCoinGameModule(repo, nil, nil, shopFixedClock{now: time.Unix(100, 0)})
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
	if len(responder.Updates) != 1 || !strings.Contains(responder.Updates[0].Embeds[0].Title, "比大小結果") {
		t.Fatalf("accept update = %#v", responder.Updates)
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

func TestCoinGameRejectsOpponentWithoutEnoughCoins(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 50})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-2", Coins: 1})
	module := NewCoinGameModule(repo, nil, nil, shopFixedClock{now: time.Unix(100, 0)})

	responder := fakediscord.NewResponder()
	if err := module.CoinGameHandler()(context.Background(), coinGameSlash(domain.CoinGameKindHigherLower, "10"), responder); err != nil {
		t.Fatalf("start game: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "對方沒有這麼多代幣") {
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
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 50})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-2", Coins: 50})
	clock := &coinGameTestClock{now: time.Unix(100, 0)}
	module := NewCoinGameModule(repo, nil, nil, clock)
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
	if len(responder.Updates) != 1 || !strings.Contains(responder.Updates[0].Embeds[0].Title, "比大小結果") {
		t.Fatalf("accept update = %#v", responder.Updates)
	}
}

func TestCoinGameKnowledgeTimeoutUsesLegacyTickAndPaysOnlyResponder(t *testing.T) {
	repo := coinGameTestRepository()
	clock := &coinGameTestClock{now: time.Unix(100, 0)}
	messages := fakediscord.NewSideEffects()
	module, scheduler := newCoinGameLifecycleTestModule(t, repo, messages, clock)
	session := acceptCoinGameForTest(t, module, domain.CoinGameKindKnowledge)
	if currentCoinGameSession(t, module, session).KnowledgeQuestion.Question == "" {
		t.Fatal("knowledge question was not initialized")
	}
	session = currentCoinGameSession(t, module, session)
	if session.TurnDeadline != time.Unix(121, 0) {
		t.Fatalf("knowledge deadline = %v", session.TurnDeadline)
	}
	if message := knowledgeQuestionMessage(session, 0); !strings.Contains(message.Embeds[0].Description, "<t:115:R>") {
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
	if len(messages.Edited) != 0 {
		t.Fatalf("timeout fired before strict legacy tick: %#v", messages.Edited)
	}

	clock.Advance(time.Second)
	if !scheduler.TriggerOnly() {
		t.Fatal("knowledge timeout retry was not scheduled")
	}
	assertCoinGameBalances(t, repo, 60, 40)
	if len(messages.Edited) != 1 {
		t.Fatalf("timeout edits = %#v", messages.Edited)
	}
	edit := messages.Edited[0]
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
	acceptCoinGameForTest(t, module, domain.CoinGameKindKnowledge)
	clock.Advance(coinGameKnowledgeTTL)
	if !scheduler.TriggerOnly() {
		t.Fatal("knowledge timer was not scheduled")
	}
	assertCoinGameBalances(t, repo, 40, 40)
	if len(messages.Edited) != 1 {
		t.Fatalf("timeout edits = %#v", messages.Edited)
	}
	title := messages.Edited[0].Message.Embeds[0].Title
	if !strings.Contains(title, "Opponent User") {
		t.Fatalf("timeout title = %q", title)
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

func TestCoinGameTimeoutSettlementFailureFailsClosed(t *testing.T) {
	base := coinGameTestRepository()
	repository := &blockingCoinGameRepository{next: base, settleErr: errors.New("settlement failed")}
	clock := &coinGameTestClock{now: time.Unix(100, 0)}
	messages := fakediscord.NewSideEffects()
	module, scheduler := newCoinGameLifecycleTestModule(t, repository, messages, clock)
	acceptCoinGameForTest(t, module, domain.CoinGameKindKnowledge)
	clock.Advance(coinGameKnowledgeTTL)
	if !scheduler.TriggerOnly() {
		t.Fatal("knowledge timer was not scheduled")
	}
	assertCoinGameBalances(t, base, 40, 40)
	if len(messages.Edited) != 0 || scheduler.Len() != 0 || repository.SettleCalls() != 1 {
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
	if kind == domain.CoinGameKindHigherLower {
		if ok {
			t.Fatal("higher/lower session should settle immediately")
		}
		return coinGameSession{}
	}
	if !ok {
		t.Fatalf("accepted %s session missing", kind)
	}
	return session
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
