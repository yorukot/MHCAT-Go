package work

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

type fakeClock struct{ now time.Time }

func (c fakeClock) Now() time.Time { return c.now }

type fakeRepo struct {
	config domain.WorkConfig
	items  []domain.WorkItem
	user   domain.WorkUserState
	start  domain.WorkUserState
	grant  domain.WorkUserState
	all    domain.WorkEnergyGrantAllResult
	err    error
}

func (r fakeRepo) GetWorkConfig(context.Context, string) (domain.WorkConfig, error) {
	if r.err != nil {
		return domain.WorkConfig{}, r.err
	}
	return r.config, nil
}

func (r fakeRepo) ListWorkItems(context.Context, string) ([]domain.WorkItem, error) {
	return append([]domain.WorkItem(nil), r.items...), nil
}

func (r fakeRepo) GetWorkUser(context.Context, string, string) (domain.WorkUserState, error) {
	if r.user.UserID == "" {
		return domain.WorkUserState{}, ports.ErrWorkUserMissing
	}
	return r.user, nil
}

func (r fakeRepo) StartWork(_ context.Context, command domain.WorkStartCommand) (domain.WorkUserState, error) {
	if command.GuildID == "" || command.UserID == "" || command.WorkName == "" || command.NowUnix <= 0 {
		return domain.WorkUserState{}, domain.ErrInvalidWorkQuery
	}
	if r.start.UserID != "" {
		return r.start, nil
	}
	return domain.WorkUserState{GuildID: command.GuildID, UserID: command.UserID, State: command.WorkName, EndTimeUnix: command.NowUnix + command.DurationSec, Energy: command.MaxEnergy - command.EnergyCost, GetCoin: command.CoinReward, Initialized: true}, nil
}

func (r fakeRepo) EnsureWorkUser(_ context.Context, guildID string, userID string, maxEnergy int64, _ string) (domain.WorkUserState, error) {
	return domain.WorkUserState{GuildID: guildID, UserID: userID, State: domain.WorkIdleState, Energy: maxEnergy, Initialized: true}, nil
}

func (r fakeRepo) SaveWorkConfig(_ context.Context, command domain.WorkConfigCommand) (domain.WorkConfig, error) {
	if command.GuildID == "" {
		return domain.WorkConfig{}, domain.ErrInvalidWorkQuery
	}
	return domain.WorkConfig{GuildID: command.GuildID, DailyEnergy: command.DailyEnergy, MaxEnergy: command.MaxEnergy, Captcha: command.Captcha}, nil
}

func (r fakeRepo) DeleteWorkItem(_ context.Context, command domain.WorkDeleteItemCommand) error {
	if command.GuildID == "" || command.Name == "" {
		return domain.ErrInvalidWorkQuery
	}
	for _, item := range r.items {
		if item.Name == command.Name {
			return nil
		}
	}
	return ports.ErrWorkItemMissing
}

func (r fakeRepo) GrantWorkEnergy(_ context.Context, command domain.WorkEnergyGrantCommand) (domain.WorkUserState, error) {
	if command.GuildID == "" || command.UserID == "" {
		return domain.WorkUserState{}, domain.ErrInvalidWorkQuery
	}
	if r.grant.UserID != "" {
		return r.grant, nil
	}
	return domain.WorkUserState{GuildID: command.GuildID, UserID: command.UserID, State: domain.WorkIdleState, Energy: command.MaxEnergy, Initialized: true}, nil
}

func (r fakeRepo) GrantWorkEnergyToAll(_ context.Context, command domain.WorkEnergyGrantAllCommand) (domain.WorkEnergyGrantAllResult, error) {
	if command.GuildID == "" {
		return domain.WorkEnergyGrantAllResult{}, domain.ErrInvalidWorkQuery
	}
	if r.all.Matched != 0 || r.all.Modified != 0 {
		return r.all, nil
	}
	return domain.WorkEnergyGrantAllResult{}, nil
}

func TestInterfaceDefaultsMissingUserWithoutWriting(t *testing.T) {
	service := NewService(fakeRepo{
		config: domain.WorkConfig{GuildID: "guild-1", MaxEnergy: 12},
		items:  []domain.WorkItem{{GuildID: "guild-1", Name: "礦坑", EnergyCost: 2, DurationSec: 60, CoinReward: 5}},
	}, fakeClock{now: time.Unix(100, 0)})

	view, err := service.Interface(context.Background(), InterfaceRequest{GuildID: "guild-1", UserID: "user-1"})
	if err != nil {
		t.Fatalf("interface: %v", err)
	}
	if view.User.Initialized {
		t.Fatal("missing user should be a read-only default, not initialized")
	}
	if view.User.State != domain.WorkIdleState || view.User.Energy != 12 || view.NowUnix != 100 {
		t.Fatalf("unexpected user/default view: %#v", view)
	}
}

func TestInterfaceInitializesMissingUserWhenStartRepositoryIsAvailable(t *testing.T) {
	repo := fakemongo.NewWorkInterfaceRepository()
	repo.PutConfig(domain.WorkConfig{GuildID: "guild-1", MaxEnergy: 20})
	repo.PutItems("guild-1", domain.WorkItem{GuildID: "guild-1", Name: "礦坑", DurationSec: 60})
	service := NewServiceWithStartRepository(repo, repo, fakeClock{now: time.Unix(100, 0)})

	view, err := service.Interface(context.Background(), InterfaceRequest{GuildID: "guild-1", UserID: "user-1"})
	if err != nil {
		t.Fatalf("interface: %v", err)
	}
	stored, err := repo.GetWorkUser(context.Background(), "guild-1", "user-1")
	if err != nil || !view.User.Initialized || !stored.Initialized || stored.Energy != 20 || stored.State != domain.WorkIdleState {
		t.Fatalf("view=%#v stored=%#v err=%v", view.User, stored, err)
	}
}

func TestInterfaceFiltersByRole(t *testing.T) {
	service := NewService(fakeRepo{
		config: domain.WorkConfig{GuildID: "guild-1", MaxEnergy: 10},
		items: []domain.WorkItem{
			{GuildID: "guild-1", Name: "public"},
			{GuildID: "guild-1", Name: "vip", RoleID: "role-vip"},
			{GuildID: "guild-1", Name: "hidden", RoleID: "role-hidden"},
		},
		user: domain.WorkUserState{GuildID: "guild-1", UserID: "user-1", State: domain.WorkIdleState, Energy: 5, Initialized: true},
	}, fakeClock{now: time.Unix(100, 0)})

	view, err := service.Interface(context.Background(), InterfaceRequest{GuildID: "guild-1", UserID: "user-1", RoleIDs: []string{"role-vip"}})
	if err != nil {
		t.Fatalf("interface: %v", err)
	}
	if len(view.VisibleItems) != 2 || view.VisibleItems[0].Name != "public" || view.VisibleItems[1].Name != "vip" {
		t.Fatalf("visible items = %#v", view.VisibleItems)
	}
}

func TestInterfaceReturnsNoVisibleWorkItem(t *testing.T) {
	service := NewService(fakeRepo{
		config: domain.WorkConfig{GuildID: "guild-1", MaxEnergy: 10},
		items:  []domain.WorkItem{{GuildID: "guild-1", Name: "vip", RoleID: "role-vip"}},
		user:   domain.WorkUserState{GuildID: "guild-1", UserID: "user-1", State: domain.WorkIdleState},
	}, nil)

	_, err := service.Interface(context.Background(), InterfaceRequest{GuildID: "guild-1", UserID: "user-1"})
	if !errors.Is(err, ports.ErrNoVisibleWorkItem) {
		t.Fatalf("expected no visible work item, got %v", err)
	}
}

func TestDetailFindsVisibleItemByStableKey(t *testing.T) {
	item := domain.WorkItem{GuildID: "guild-1", Name: "礦坑"}
	service := NewService(fakeRepo{
		config: domain.WorkConfig{GuildID: "guild-1", MaxEnergy: 10},
		items:  []domain.WorkItem{item},
		user:   domain.WorkUserState{GuildID: "guild-1", UserID: "user-1", State: domain.WorkIdleState},
	}, nil)

	_, got, err := service.Detail(context.Background(), InterfaceRequest{GuildID: "guild-1", UserID: "user-1"}, item.Key())
	if err != nil {
		t.Fatalf("detail: %v", err)
	}
	if got.Name != item.Name {
		t.Fatalf("got item = %#v", got)
	}
}

func TestStartWorkUsesRepository(t *testing.T) {
	item := domain.WorkItem{GuildID: "guild-1", Name: "礦坑", DurationSec: 60, EnergyCost: 2, CoinReward: 5}
	repo := fakeRepo{
		config: domain.WorkConfig{GuildID: "guild-1", MaxEnergy: 10},
		items:  []domain.WorkItem{item},
		user:   domain.WorkUserState{GuildID: "guild-1", UserID: "user-1", State: domain.WorkIdleState, Energy: 10, Initialized: true},
	}
	service := NewServiceWithStartRepository(repo, repo, fakeClock{now: time.Unix(100, 0)})

	_, _, updated, err := service.Start(context.Background(), InterfaceRequest{GuildID: "guild-1", UserID: "user-1"}, item.Key(), false)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if updated.State != "礦坑" || updated.EndTimeUnix != 160 || updated.Energy != 8 || updated.GetCoin != 5 {
		t.Fatalf("updated = %#v", updated)
	}
}

func TestStartWorkPreservesMixedScalarArithmetic(t *testing.T) {
	repo := fakemongo.NewWorkInterfaceRepository()
	item := domain.WorkItem{
		GuildID: "guild-1", Name: "礦坑", DurationText: "0.5",
		EnergyCostText: "2.25", CoinRewardText: "3.75",
	}
	repo.PutConfig(domain.WorkConfig{GuildID: "guild-1", MaxEnergy: 10})
	repo.PutItems("guild-1", item)
	repo.PutUser(domain.WorkUserState{
		GuildID: "guild-1", UserID: "user-1", State: domain.WorkIdleState,
		Energy: 5, EnergyText: "5.5", Initialized: true,
	})
	service := NewServiceWithStartRepository(repo, repo, fakeClock{now: time.Unix(100, 0)})

	_, _, updated, err := service.Start(context.Background(), InterfaceRequest{GuildID: "guild-1", UserID: "user-1"}, item.Key(), false)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if updated.EndTimeText != "100.5" || updated.EnergyText != "3.25" || updated.GetCoinText != "3.75" {
		t.Fatalf("updated = %#v", updated)
	}
}

func TestStartWorkTreatsEmptyLegacyStateAsBusy(t *testing.T) {
	repo := fakemongo.NewWorkInterfaceRepository()
	item := domain.WorkItem{GuildID: "guild-1", Name: "礦坑", DurationSec: 60, EnergyCost: 2, CoinReward: 5}
	repo.PutConfig(domain.WorkConfig{GuildID: "guild-1", MaxEnergy: 10})
	repo.PutItems("guild-1", item)
	repo.PutUser(domain.WorkUserState{GuildID: "guild-1", UserID: "user-1", State: "", Energy: 10, Initialized: true})
	service := NewServiceWithStartRepository(repo, repo, fakeClock{now: time.Unix(100, 0)})

	_, _, _, err := service.Start(context.Background(), InterfaceRequest{GuildID: "guild-1", UserID: "user-1"}, item.Key(), false)
	if !errors.Is(err, domain.ErrWorkAlreadyBusy) {
		t.Fatalf("expected busy state, got %v", err)
	}
}

func TestStartWorkRequiresExplicitStartRepository(t *testing.T) {
	item := domain.WorkItem{GuildID: "guild-1", Name: "礦坑", DurationSec: 60, EnergyCost: 2, CoinReward: 5}
	service := NewService(fakeRepo{
		config: domain.WorkConfig{GuildID: "guild-1", MaxEnergy: 10},
		items:  []domain.WorkItem{item},
		user:   domain.WorkUserState{GuildID: "guild-1", UserID: "user-1", State: domain.WorkIdleState, Energy: 10, Initialized: true},
	}, fakeClock{now: time.Unix(100, 0)})

	if service.CanStart() {
		t.Fatal("read-only constructor should not enable start writes")
	}
	_, _, _, err := service.Start(context.Background(), InterfaceRequest{GuildID: "guild-1", UserID: "user-1"}, item.Key(), false)
	if !errors.Is(err, domain.ErrWorkStartUnavailable) {
		t.Fatalf("expected start unavailable, got %v", err)
	}
}

func TestAdminMethodsRequireExplicitAdminRepository(t *testing.T) {
	service := NewService(fakeRepo{}, nil)
	if service.CanAdmin() {
		t.Fatal("read-only constructor should not enable admin writes")
	}
	_, err := service.SaveConfig(context.Background(), domain.WorkConfigCommand{GuildID: "guild-1", MaxEnergy: 10})
	if !errors.Is(err, domain.ErrWorkAdminUnavailable) {
		t.Fatalf("expected admin unavailable, got %v", err)
	}
}

func TestAdminMethodsUseRepositoryAndConfig(t *testing.T) {
	repo := fakeRepo{
		config: domain.WorkConfig{GuildID: "guild-1", MaxEnergy: 10},
		items:  []domain.WorkItem{{GuildID: "guild-1", Name: "礦坑"}},
		grant:  domain.WorkUserState{GuildID: "guild-1", UserID: "target", State: domain.WorkIdleState, Energy: 10, Initialized: true},
		all:    domain.WorkEnergyGrantAllResult{Matched: 2, Modified: 1},
	}
	service := NewServiceWithAdminRepository(repo, nil)
	if !service.CanStart() || !service.CanAdmin() {
		t.Fatal("admin constructor should enable start and admin writes")
	}
	config, err := service.SaveConfig(context.Background(), domain.WorkConfigCommand{GuildID: "guild-1", DailyEnergy: 3, MaxEnergy: 10, Captcha: true})
	if err != nil || !config.Captcha || config.DailyEnergy != 3 {
		t.Fatalf("save config = %#v, %v", config, err)
	}
	if err := service.DeleteItem(context.Background(), domain.WorkDeleteItemCommand{GuildID: "guild-1", Name: "礦坑"}); err != nil {
		t.Fatalf("delete item: %v", err)
	}
	user, err := service.GrantEnergy(context.Background(), domain.WorkEnergyGrantCommand{GuildID: "guild-1", UserID: "target", Amount: 5})
	if err != nil || user.UserID != "target" || user.Energy != 10 {
		t.Fatalf("grant energy = %#v, %v", user, err)
	}
	all, err := service.GrantEnergyToAll(context.Background(), domain.WorkEnergyGrantAllCommand{GuildID: "guild-1", Amount: 5})
	if err != nil || all.Matched != 2 || all.Modified != 1 {
		t.Fatalf("grant all = %#v, %v", all, err)
	}
}

func TestAdminMethodsAcceptLegacySignedEnergyAmounts(t *testing.T) {
	repo := fakemongo.NewWorkInterfaceRepository()
	repo.PutConfig(domain.WorkConfig{GuildID: "guild-1", MaxEnergy: 10})
	repo.PutUser(domain.WorkUserState{GuildID: "guild-1", UserID: "target", Energy: 5, Initialized: true})
	service := NewServiceWithAdminRepository(repo, nil)
	user, err := service.GrantEnergy(context.Background(), domain.WorkEnergyGrantCommand{GuildID: "guild-1", UserID: "target", Amount: -2})
	if err != nil || user.Energy != 3 {
		t.Fatalf("negative grant = %#v, %v", user, err)
	}
	result, err := service.GrantEnergyToAll(context.Background(), domain.WorkEnergyGrantAllCommand{GuildID: "guild-1", Amount: 0})
	if err != nil || result.Matched != 1 || result.Modified != 0 {
		t.Fatalf("zero grant = %#v, %v", result, err)
	}
}

func TestInvalidRequestFails(t *testing.T) {
	service := NewService(fakeRepo{}, nil)
	if _, err := service.Interface(context.Background(), InterfaceRequest{GuildID: "", UserID: "user-1"}); !errors.Is(err, domain.ErrInvalidWorkQuery) {
		t.Fatalf("expected invalid work query, got %v", err)
	}
}
