package lottery

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestJoinPreservesLegacyEligibilityRules(t *testing.T) {
	repo := fakemongo.NewLotteryRepository()
	repo.Lotteries["guild-1:id-1"] = domain.Lottery{
		GuildID:         "guild-1",
		ID:              "id-1",
		EndsAtUnix:      200,
		RequiredRoleID:  "role-required",
		ForbiddenRoleID: "role-forbidden",
		MaxParticipants: 2,
	}
	service := NewService(repo)
	now := time.Unix(100, 0)

	joined, err := service.Join(context.Background(), domain.LotteryJoinRequest{GuildID: "guild-1", ID: "id-1", UserID: "user-1"}, []string{"role-required"}, now)
	if err != nil {
		t.Fatalf("join: %v", err)
	}
	if len(joined.Participants) != 1 || joined.Participants[0].JoinedAtMillis != now.UnixMilli() {
		t.Fatalf("joined = %#v", joined)
	}
	_, err = service.Join(context.Background(), domain.LotteryJoinRequest{GuildID: "guild-1", ID: "id-1", UserID: "user-1"}, []string{"role-required"}, now)
	if !errors.Is(err, ports.ErrLotteryAlreadyJoined) {
		t.Fatalf("duplicate error = %v", err)
	}
	_, err = service.Join(context.Background(), domain.LotteryJoinRequest{GuildID: "guild-1", ID: "id-1", UserID: "user-2"}, []string{"role-required", "role-forbidden"}, now)
	if !errors.Is(err, ports.ErrLotteryRoleDenied) {
		t.Fatalf("role error = %v", err)
	}
}

func TestManagedLotteryUsesLegacyOwnerFallback(t *testing.T) {
	repo := fakemongo.NewLotteryRepository()
	service := NewService(repo)
	repo.Lotteries["guild-1:owned"] = domain.Lottery{GuildID: "guild-1", ID: "owned", OwnerID: "owner-1"}
	repo.Lotteries["guild-1:legacy"] = domain.Lottery{GuildID: "guild-1", ID: "legacy"}

	if _, err := service.GetManaged(context.Background(), "guild-1", "owned", "moderator-1", "guild-owner", true); !errors.Is(err, ports.ErrLotteryManagerOnly) {
		t.Fatalf("owned moderator error = %v", err)
	}
	if _, err := service.GetManaged(context.Background(), "guild-1", "owned", "guild-owner", "guild-owner", false); err != nil {
		t.Fatalf("guild owner: %v", err)
	}
	if _, err := service.GetManaged(context.Background(), "guild-1", "legacy", "moderator-1", "guild-owner", true); err != nil {
		t.Fatalf("legacy moderator: %v", err)
	}
}
