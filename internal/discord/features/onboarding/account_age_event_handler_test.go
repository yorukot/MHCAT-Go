package onboarding

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	discordevents "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakebotinfo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

type accountAgeEventClock struct {
	now time.Time
}

func (c accountAgeEventClock) Now() time.Time {
	return c.now
}

func TestAccountAgeGateStopsLaterMemberAddHandlersAfterKick(t *testing.T) {
	now := time.Unix(2_000_000, 0)
	repo := fakemongo.NewAccountAgeConfigRepository()
	repo.Configs["guild-1"] = domain.AccountAgeConfig{GuildID: "guild-1", RequiredSeconds: 3600}
	sideEffects := fakediscord.NewSideEffects()
	info := &fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{Name: "測試伺服器"}}
	module := NewAccountAgePolicyModule(repo, sideEffects, sideEffects, sideEffects, info, accountAgeEventClock{now: now})
	dispatcher := discordevents.NewDispatcher(nil)
	module.RegisterEventRoutes(dispatcher)
	laterCalled := false
	dispatcher.Register(discordevents.TypeMemberAdd, func(ctx context.Context, event discordevents.Event) error {
		laterCalled = true
		return nil
	})

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
	if laterCalled {
		t.Fatalf("later member-add handler should not run after account-age kick")
	}
	if len(sideEffects.Kicked) != 1 {
		t.Fatalf("kicked = %#v", sideEffects.Kicked)
	}
}

func TestAccountAgeGateUsesEventGuildNameInDM(t *testing.T) {
	now := time.Unix(2_000_000, 0)
	repo := fakemongo.NewAccountAgeConfigRepository()
	repo.Configs["guild-1"] = domain.AccountAgeConfig{GuildID: "guild-1", RequiredSeconds: 3600}
	sideEffects := fakediscord.NewSideEffects()
	module := NewAccountAgePolicyModule(repo, sideEffects, sideEffects, sideEffects, nil, accountAgeEventClock{now: now})
	dispatcher := discordevents.NewDispatcher(nil)
	module.RegisterEventRoutes(dispatcher)

	err := dispatcher.Dispatch(context.Background(), discordevents.Event{
		Type:      discordevents.TypeMemberAdd,
		GuildID:   "guild-1",
		GuildName: "測試伺服器",
		Member: &discordevents.Member{
			UserID:           "user-1",
			UserTag:          "Tester#0001",
			AccountCreatedAt: now.Add(-time.Minute),
		},
	})
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if len(sideEffects.DirectMessages) != 1 {
		t.Fatalf("direct messages = %#v", sideEffects.DirectMessages)
	}
	if got := sideEffects.DirectMessages[0].Message.Embeds[0].Description; !strings.Contains(got, "已將您踢出`測試伺服器`") {
		t.Fatalf("dm description = %q", got)
	}
}

func TestAccountAgeGateAllowsLaterHandlersForOldEnoughMember(t *testing.T) {
	now := time.Unix(2_000_000, 0)
	repo := fakemongo.NewAccountAgeConfigRepository()
	repo.Configs["guild-1"] = domain.AccountAgeConfig{GuildID: "guild-1", RequiredSeconds: 3600}
	sideEffects := fakediscord.NewSideEffects()
	module := NewAccountAgePolicyModule(repo, sideEffects, sideEffects, sideEffects, nil, accountAgeEventClock{now: now})
	dispatcher := discordevents.NewDispatcher(nil)
	module.RegisterEventRoutes(dispatcher)
	laterCalled := false
	dispatcher.Register(discordevents.TypeMemberAdd, func(ctx context.Context, event discordevents.Event) error {
		laterCalled = true
		return nil
	})

	err := dispatcher.Dispatch(context.Background(), discordevents.Event{
		Type:    discordevents.TypeMemberAdd,
		GuildID: "guild-1",
		Member: &discordevents.Member{
			UserID:           "user-1",
			AccountCreatedAt: now.Add(-2 * time.Hour),
		},
	})
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if !laterCalled {
		t.Fatalf("later member-add handler should run for old enough member")
	}
	if len(sideEffects.Kicked) != 0 {
		t.Fatalf("kicked = %#v", sideEffects.Kicked)
	}
}

func TestAccountAgeGateAllowsLaterHandlersForInvalidLegacyConfig(t *testing.T) {
	now := time.Unix(2_000_000, 0)
	repo := fakemongo.NewAccountAgeConfigRepository()
	repo.Configs["guild-1"] = domain.AccountAgeConfig{GuildID: "guild-1"}
	sideEffects := fakediscord.NewSideEffects()
	module := NewAccountAgePolicyModule(repo, sideEffects, sideEffects, sideEffects, nil, accountAgeEventClock{now: now})
	dispatcher := discordevents.NewDispatcher(nil)
	module.RegisterEventRoutes(dispatcher)
	laterCalled := false
	dispatcher.Register(discordevents.TypeMemberAdd, func(ctx context.Context, event discordevents.Event) error {
		laterCalled = true
		return nil
	})

	err := dispatcher.Dispatch(context.Background(), discordevents.Event{
		Type:    discordevents.TypeMemberAdd,
		GuildID: "guild-1",
		Member: &discordevents.Member{
			UserID:           "user-1",
			AccountCreatedAt: now.Add(-time.Minute),
		},
	})
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if !laterCalled {
		t.Fatal("invalid legacy threshold should not block later member-add handlers")
	}
	if len(sideEffects.Kicked) != 0 {
		t.Fatalf("kicked = %#v", sideEffects.Kicked)
	}
}

func TestAccountAgeReadFailureDoesNotSuppressLaterMemberAddHandler(t *testing.T) {
	repo := fakemongo.NewAccountAgeConfigRepository()
	wantErr := errors.New("account age read failed")
	repo.Err = wantErr
	sideEffects := fakediscord.NewSideEffects()
	module := NewAccountAgePolicyModule(repo, sideEffects, sideEffects, sideEffects, nil, accountAgeEventClock{})
	dispatcher := discordevents.NewDispatcher(nil)
	module.RegisterEventRoutes(dispatcher)
	laterCalled := false
	dispatcher.Register(discordevents.TypeMemberAdd, func(context.Context, discordevents.Event) error {
		laterCalled = true
		return nil
	})

	err := dispatcher.Dispatch(context.Background(), discordevents.Event{
		Type:    discordevents.TypeMemberAdd,
		GuildID: "guild-1",
		Member:  &discordevents.Member{UserID: "user-1", AccountCreatedAt: time.Unix(1_000_000, 0)},
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("dispatch error = %v", err)
	}
	if !laterCalled {
		t.Fatal("later member-add handler was suppressed")
	}
}
