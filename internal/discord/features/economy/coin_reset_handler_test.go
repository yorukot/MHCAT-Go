package economy

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakebotinfo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestCoinResetDeletesGuildBalancesAfterOwnerConfirmation(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 100})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-2", Coins: 200})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-2", UserID: "user-3", Coins: 300})
	sideEffects := fakediscord.NewSideEffects()
	usage := &fakeusage.Tracker{}
	module, clock := coinResetTestModule(repo, sideEffects, usage)

	responder := fakediscord.NewResponder()
	if err := module.CoinResetHandler()(context.Background(), coinResetInteraction(nil, "user-1"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Replies) != 1 || responder.Replies[0].Content != coinResetWarningContent || responder.Replies[0].Ephemeral {
		t.Fatalf("warning reply = %#v", responder.Replies)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != CoinResetCommandName || usage.Events[0].Feature != "economy-coin-reset" {
		t.Fatalf("usage = %#v", usage.Events)
	}

	if err := module.CoinResetConfirmationHandler()(context.Background(), coinResetMessage(coinResetConfirmContent, "user-1", clock.now.Add(time.Second))); err != nil {
		t.Fatalf("confirmation: %v", err)
	}
	if _, err := repo.GetCoinBalance(context.Background(), "guild-1", "user-1"); !errors.Is(err, ports.ErrCoinBalanceNotFound) {
		t.Fatalf("user-1 should be deleted, err=%v", err)
	}
	if balance, err := repo.GetCoinBalance(context.Background(), "guild-2", "user-3"); err != nil || balance.Coins != 300 {
		t.Fatalf("other guild balance = %#v err=%v", balance, err)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].Message.Embeds[0].Title != coinResetDeleteEmoji+"成功重製伺服器內所有代幣" {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
}

func TestCoinResetDividesGuildBalancesWithLegacyRounding(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 101})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-2", Coins: 100})
	sideEffects := fakediscord.NewSideEffects()
	module, clock := coinResetTestModule(repo, sideEffects, nil)
	divisor := int64(2)

	responder := fakediscord.NewResponder()
	if err := module.CoinResetHandler()(context.Background(), coinResetInteraction(&divisor, "user-1"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if err := module.CoinResetConfirmationHandler()(context.Background(), coinResetMessage(coinResetConfirmContent, "user-1", clock.now.Add(time.Second))); err != nil {
		t.Fatalf("confirmation: %v", err)
	}
	if balance, _ := repo.GetCoinBalance(context.Background(), "guild-1", "user-1"); balance.Coins != 51 {
		t.Fatalf("user-1 balance = %#v", balance)
	}
	if balance, _ := repo.GetCoinBalance(context.Background(), "guild-1", "user-2"); balance.Coins != 50 {
		t.Fatalf("user-2 balance = %#v", balance)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].Message.Embeds[0].Color != coinResetSuccessColor {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
}

func TestCoinResetWrongConfirmationCancelsWithoutMutation(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 100})
	sideEffects := fakediscord.NewSideEffects()
	module, clock := coinResetTestModule(repo, sideEffects, nil)

	responder := fakediscord.NewResponder()
	if err := module.CoinResetHandler()(context.Background(), coinResetInteraction(nil, "user-1"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if err := module.CoinResetConfirmationHandler()(context.Background(), coinResetMessage("確認", "user-1", clock.now.Add(time.Second))); err != nil {
		t.Fatalf("confirmation: %v", err)
	}
	if balance, err := repo.GetCoinBalance(context.Background(), "guild-1", "user-1"); err != nil || balance.Coins != 100 {
		t.Fatalf("balance = %#v err=%v", balance, err)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].Message.Embeds[0].Title != coinResetErrorPrefix+"你輸入了錯誤的確認!因此視為取消還原" {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
}

func TestCoinResetRejectsNonOwner(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	sideEffects := fakediscord.NewSideEffects()
	info := &fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{OwnerID: "owner-1"}}
	module := NewCoinResetModule(repo, info, sideEffects, nil, coinResetTestClock{now: coinResetTestNow})

	responder := fakediscord.NewResponder()
	if err := module.CoinResetHandler()(context.Background(), coinResetInteraction(nil, "user-1"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Replies) != 1 || !responder.Replies[0].Ephemeral || responder.Replies[0].Embeds[0].Title != coinResetErrorPrefix+"只有服主可以使用這個指令!" {
		t.Fatalf("reply = %#v", responder.Replies)
	}
	if err := module.CoinResetConfirmationHandler()(context.Background(), coinResetMessage(coinResetConfirmContent, "user-1", coinResetTestNow.Add(time.Second))); err != nil {
		t.Fatalf("confirmation: %v", err)
	}
	if len(sideEffects.Sent) != 0 {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
}

func TestCoinResetMissingGuildDataSendsLegacyError(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	sideEffects := fakediscord.NewSideEffects()
	module, clock := coinResetTestModule(repo, sideEffects, nil)

	responder := fakediscord.NewResponder()
	if err := module.CoinResetHandler()(context.Background(), coinResetInteraction(nil, "user-1"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if err := module.CoinResetConfirmationHandler()(context.Background(), coinResetMessage(coinResetConfirmContent, "user-1", clock.now.Add(time.Second))); err != nil {
		t.Fatalf("confirmation: %v", err)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].Message.Embeds[0].Title != coinResetErrorPrefix+"這伺服器沒有任何的代幣喔!" {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
}

var coinResetTestNow = time.Date(2026, 7, 9, 12, 0, 0, 0, time.UTC)

type coinResetTestClock struct {
	now time.Time
}

func (c coinResetTestClock) Now() time.Time {
	return c.now
}

func coinResetTestModule(repo *fakemongo.EconomyRepository, sideEffects *fakediscord.SideEffects, usage ports.UsageTracker) (Module, coinResetTestClock) {
	clock := coinResetTestClock{now: coinResetTestNow}
	info := &fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{OwnerID: "user-1"}}
	return NewCoinResetModule(repo, info, sideEffects, usage, clock), clock
}

func coinResetInteraction(divisor *int64, actorUserID string) interactions.Interaction {
	interaction := fakediscord.SlashInteraction(CoinResetCommandName)
	interaction.Actor.UserID = actorUserID
	interaction.ChannelID = "channel-1"
	interaction.CommandOptions = map[string]interactions.CommandOptionValue{}
	if divisor != nil {
		interaction.Options[coinResetOptionDivisor] = strconv.FormatInt(*divisor, 10)
		interaction.CommandOptions[coinResetOptionDivisor] = interactions.CommandOptionValue{
			Type:   interactions.CommandOptionInteger,
			Int:    *divisor,
			String: strconv.FormatInt(*divisor, 10),
		}
	}
	return interaction
}

func coinResetMessage(content string, userID string, createdAt time.Time) events.Event {
	return events.Event{
		Type:      events.TypeMessageCreate,
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		UserID:    userID,
		Content:   content,
		CreatedAt: createdAt,
	}
}
