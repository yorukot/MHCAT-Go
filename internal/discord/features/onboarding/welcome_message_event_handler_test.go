package onboarding

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/onboarding"
	discordevents "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestWelcomeMessageDeliveryHandlerSendsOnMemberAdd(t *testing.T) {
	now := time.Unix(2_000_000, 0)
	repo := fakemongo.NewJoinMessageConfigRepository()
	repo.Configs["guild-1"] = domain.JoinMessageConfig{
		GuildID:        "guild-1",
		Enabled:        true,
		ChannelID:      "channel-1",
		MessageContent: "歡迎 {MEMBERNAME} {TAG}",
		Color:          "#53FF53",
	}
	sideEffects := fakediscord.NewSideEffects()
	module := NewWelcomeMessageDeliveryModule(repo, sideEffects, emptySpecialWelcome())
	dispatcher := discordevents.NewDispatcher(nil)
	module.RegisterEventRoutes(dispatcher)

	err := dispatcher.Dispatch(context.Background(), discordevents.Event{
		Type:         discordevents.TypeMemberAdd,
		GuildID:      "guild-1",
		GuildName:    "測試伺服器",
		GuildIconURL: "https://example.test/guild.png",
		BotAvatarURL: "https://example.test/bot.png",
		CreatedAt:    now,
		Member: &discordevents.Member{
			UserID:    "user-1",
			Username:  "Tester",
			UserTag:   "Tester#0001",
			AvatarURL: "https://example.test/avatar.png",
		},
	})
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if len(sideEffects.Sent) != 1 {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
	embed := sideEffects.Sent[0].Message.Embeds[0]
	if !strings.Contains(embed.Description, "歡迎 Tester <@user-1>") || embed.AuthorName != "🪂 歡迎加入 測試伺服器" {
		t.Fatalf("embed = %#v", embed)
	}
}

func TestWelcomeMessageDeliveryHandlerPreservesRawUsername(t *testing.T) {
	repo := fakemongo.NewJoinMessageConfigRepository()
	repo.Configs["guild-1"] = domain.JoinMessageConfig{
		GuildID:        "guild-1",
		Enabled:        true,
		ChannelID:      "channel-1",
		MessageContent: "歡迎 {MEMBERNAME}",
		Color:          "#53FF53",
	}
	sideEffects := fakediscord.NewSideEffects()
	module := NewWelcomeMessageDeliveryModule(repo, sideEffects, emptySpecialWelcome())

	err := module.WelcomeMessageDeliveryHandler()(context.Background(), discordevents.Event{
		Type:    discordevents.TypeMemberAdd,
		GuildID: "guild-1",
		Member: &discordevents.Member{
			UserID:   "user-1",
			Username: "  $&  ",
			UserTag:  "fallback#0001",
		},
	})
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if got := sideEffects.Sent[0].Message.Embeds[0].Description; got != "歡迎   {MEMBERNAME}  " {
		t.Fatalf("description = %q", got)
	}
}

func TestWelcomeMessageDeliveryHandlerIgnoresMemberRemove(t *testing.T) {
	repo := fakemongo.NewJoinMessageConfigRepository()
	sideEffects := fakediscord.NewSideEffects()
	module := NewWelcomeMessageDeliveryModule(repo, sideEffects, emptySpecialWelcome())
	dispatcher := discordevents.NewDispatcher(nil)
	module.RegisterEventRoutes(dispatcher)

	err := dispatcher.Dispatch(context.Background(), discordevents.Event{
		Type:    discordevents.TypeMemberRemove,
		GuildID: "guild-1",
		Member:  &discordevents.Member{UserID: "user-1"},
	})
	if err == nil {
		t.Fatal("expected no member-remove handler")
	}
	if len(sideEffects.Sent) != 0 {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
}

func TestWelcomeMessageFailureDoesNotSuppressLaterMemberAddHandler(t *testing.T) {
	repo := fakemongo.NewJoinMessageConfigRepository()
	wantErr := errors.New("join message read failed")
	repo.Err = wantErr
	sideEffects := fakediscord.NewSideEffects()
	dispatcher := discordevents.NewDispatcher(nil)
	NewWelcomeMessageDeliveryModule(repo, sideEffects, emptySpecialWelcome()).RegisterEventRoutes(dispatcher)
	laterCalled := false
	dispatcher.Register(discordevents.TypeMemberAdd, func(context.Context, discordevents.Event) error {
		laterCalled = true
		return nil
	})

	err := dispatcher.Dispatch(context.Background(), discordevents.Event{
		Type:    discordevents.TypeMemberAdd,
		GuildID: "guild-1",
		Member:  &discordevents.Member{UserID: "user-1", Username: "Tester"},
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("dispatch error = %v", err)
	}
	if !laterCalled {
		t.Fatal("later member-add handler was suppressed")
	}
}

func TestAccountAgeStopPropagationPreventsWelcomeMessage(t *testing.T) {
	now := time.Unix(2_000_000, 0)
	accountAgeRepo := fakemongo.NewAccountAgeConfigRepository()
	accountAgeRepo.Configs["guild-1"] = domain.AccountAgeConfig{GuildID: "guild-1", RequiredSeconds: 3600}
	joinMessageRepo := fakemongo.NewJoinMessageConfigRepository()
	joinMessageRepo.Configs["guild-1"] = domain.JoinMessageConfig{
		GuildID:        "guild-1",
		Enabled:        true,
		ChannelID:      "channel-1",
		MessageContent: "歡迎 {MEMBERNAME}",
		Color:          "#53FF53",
	}
	sideEffects := fakediscord.NewSideEffects()
	dispatcher := discordevents.NewDispatcher(nil)
	NewAccountAgePolicyModule(accountAgeRepo, sideEffects, sideEffects, sideEffects, nil, accountAgeEventClock{now: now}).RegisterEventRoutes(dispatcher)
	NewWelcomeMessageDeliveryModule(joinMessageRepo, sideEffects, emptySpecialWelcome()).RegisterEventRoutes(dispatcher)

	err := dispatcher.Dispatch(context.Background(), discordevents.Event{
		Type:    discordevents.TypeMemberAdd,
		GuildID: "guild-1",
		Member: &discordevents.Member{
			UserID:           "user-1",
			UserTag:          "Tester#0001",
			AccountCreatedAt: now.Add(-time.Minute),
		},
	})
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if len(sideEffects.Kicked) != 1 {
		t.Fatalf("kicked = %#v", sideEffects.Kicked)
	}
	if len(sideEffects.Sent) != 0 {
		t.Fatalf("welcome should not send after account-age stop: %#v", sideEffects.Sent)
	}
}

func emptySpecialWelcome() coreservice.SpecialWelcomeConfig {
	return coreservice.SpecialWelcomeConfig{}
}
