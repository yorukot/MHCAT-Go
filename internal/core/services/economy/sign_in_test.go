package economy

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestSignInUsesTaipeiDateFields(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	now := time.Date(2026, 7, 3, 17, 30, 0, 0, time.UTC)
	repo.PutSignInResult(domain.SignInResult{
		Balance: domain.CoinBalance{GuildID: "guild", UserID: "user", Coins: 25},
		Reward:  25,
	})
	_, err := (SignInService{Repository: repo}).SignIn(context.Background(), "guild", "user", now)
	if err != nil {
		t.Fatalf("sign in: %v", err)
	}
	if len(repo.SignInCommands) != 1 {
		t.Fatalf("commands = %#v", repo.SignInCommands)
	}
	command := repo.SignInCommands[0]
	if command.Year != "2026" || command.Month != "07" || command.Day != "4" {
		t.Fatalf("unexpected Taipei date command: %#v", command)
	}
}

func TestSignInRequiresRepositoryAndIDs(t *testing.T) {
	_, err := (SignInService{}).SignIn(context.Background(), "guild", "user", time.Now())
	if !errors.Is(err, domain.ErrInvalidSignIn) {
		t.Fatalf("expected invalid sign in for nil repo, got %v", err)
	}
	_, err = (SignInService{Repository: fakemongo.NewEconomyRepository()}).SignIn(context.Background(), "", "user", time.Now())
	if !errors.Is(err, domain.ErrInvalidSignIn) {
		t.Fatalf("expected invalid sign in for missing ids, got %v", err)
	}
}
