package work

import (
	"context"
	"errors"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type Service struct {
	repo      ports.WorkInterfaceRepository
	startRepo ports.WorkStartRepository
	adminRepo ports.WorkAdminRepository
	clock     ports.Clock
}

type InterfaceRequest struct {
	GuildID       string
	GuildName     string
	UserID        string
	UserTag       string
	UserAvatarURL string
	RoleIDs       []string
	NowUnix       int64
}

func NewService(repo ports.WorkInterfaceRepository, clock ports.Clock) Service {
	if clock == nil {
		clock = ports.SystemClock{}
	}
	return Service{repo: repo, clock: clock}
}

func NewServiceWithStartRepository(repo ports.WorkInterfaceRepository, startRepo ports.WorkStartRepository, clock ports.Clock) Service {
	service := NewService(repo, clock)
	if startRepo != nil {
		service.startRepo = startRepo
	}
	return service
}

func NewServiceWithAdminRepository(repo ports.WorkAdminRepository, clock ports.Clock) Service {
	service := NewServiceWithStartRepository(repo, repo, clock)
	service.adminRepo = repo
	return service
}

func (s Service) CanStart() bool {
	return s.startRepo != nil
}

func (s Service) CanAdmin() bool {
	return s.adminRepo != nil
}

func (s Service) Settings(ctx context.Context, guildID string) (domain.WorkConfig, error) {
	if s.repo == nil || strings.TrimSpace(guildID) == "" {
		return domain.WorkConfig{}, domain.ErrInvalidWorkQuery
	}
	return s.repo.GetWorkConfig(ctx, strings.TrimSpace(guildID))
}

func (s Service) Interface(ctx context.Context, request InterfaceRequest) (domain.WorkInterfaceView, error) {
	if err := validateRequest(request); err != nil {
		return domain.WorkInterfaceView{}, err
	}
	config, err := s.repo.GetWorkConfig(ctx, strings.TrimSpace(request.GuildID))
	if err != nil {
		return domain.WorkInterfaceView{}, err
	}
	items, err := s.repo.ListWorkItems(ctx, strings.TrimSpace(request.GuildID))
	if err != nil {
		return domain.WorkInterfaceView{}, err
	}
	if len(items) == 0 {
		return domain.WorkInterfaceView{}, ports.ErrWorkItemsMissing
	}
	user, err := s.repo.GetWorkUser(ctx, strings.TrimSpace(request.GuildID), strings.TrimSpace(request.UserID))
	if err != nil {
		if !errors.Is(err, ports.ErrWorkUserMissing) {
			return domain.WorkInterfaceView{}, err
		}
		user = domain.WorkUserState{
			GuildID:     strings.TrimSpace(request.GuildID),
			UserID:      strings.TrimSpace(request.UserID),
			State:       domain.WorkIdleState,
			EndTimeUnix: 0,
			Energy:      config.MaxEnergy,
			GetCoin:     0,
			Initialized: false,
		}
	}
	nowUnix := request.NowUnix
	if nowUnix <= 0 {
		nowUnix = s.clock.Now().Unix()
	}
	view := domain.WorkInterfaceView{
		Config:        config,
		User:          user,
		Items:         append([]domain.WorkItem(nil), items...),
		VisibleItems:  visibleItems(items, request.RoleIDs),
		NowUnix:       nowUnix,
		GuildName:     strings.TrimSpace(request.GuildName),
		UserTag:       strings.TrimSpace(request.UserTag),
		UserAvatarURL: strings.TrimSpace(request.UserAvatarURL),
	}
	if len(view.VisibleItems) == 0 {
		return domain.WorkInterfaceView{}, ports.ErrNoVisibleWorkItem
	}
	return view, nil
}

func (s Service) Detail(ctx context.Context, request InterfaceRequest, itemKey string) (domain.WorkInterfaceView, domain.WorkItem, error) {
	itemKey = strings.TrimSpace(itemKey)
	if itemKey == "" {
		return domain.WorkInterfaceView{}, domain.WorkItem{}, domain.ErrInvalidWorkQuery
	}
	view, err := s.Interface(ctx, request)
	if err != nil {
		return domain.WorkInterfaceView{}, domain.WorkItem{}, err
	}
	var matched []domain.WorkItem
	for _, item := range view.VisibleItems {
		if item.Key() == itemKey {
			matched = append(matched, item)
		}
	}
	if len(matched) == 0 {
		return domain.WorkInterfaceView{}, domain.WorkItem{}, ports.ErrNoVisibleWorkItem
	}
	if len(matched) > 1 {
		return domain.WorkInterfaceView{}, domain.WorkItem{}, domain.ErrWorkItemKeyConflict
	}
	return view, matched[0], nil
}

func (s Service) Start(ctx context.Context, request InterfaceRequest, itemKey string, override bool) (domain.WorkInterfaceView, domain.WorkItem, domain.WorkUserState, error) {
	if s.startRepo == nil {
		return domain.WorkInterfaceView{}, domain.WorkItem{}, domain.WorkUserState{}, domain.ErrWorkStartUnavailable
	}
	view, item, err := s.Detail(ctx, request, itemKey)
	if err != nil {
		return domain.WorkInterfaceView{}, domain.WorkItem{}, domain.WorkUserState{}, err
	}
	nowUnix := request.NowUnix
	if nowUnix <= 0 {
		nowUnix = s.clock.Now().Unix()
	}
	updated, err := s.startRepo.StartWork(ctx, domain.WorkStartCommand{
		GuildID:     request.GuildID,
		UserID:      request.UserID,
		WorkName:    item.Name,
		DurationSec: item.DurationSec,
		EnergyCost:  item.EnergyCost,
		CoinReward:  item.CoinReward,
		MaxEnergy:   view.Config.MaxEnergy,
		NowUnix:     nowUnix,
		Override:    override,
	})
	if err != nil {
		return view, item, domain.WorkUserState{}, err
	}
	return view, item, updated, nil
}

func (s Service) SaveConfig(ctx context.Context, command domain.WorkConfigCommand) (domain.WorkConfig, error) {
	if s.adminRepo == nil {
		return domain.WorkConfig{}, domain.ErrWorkAdminUnavailable
	}
	command.GuildID = strings.TrimSpace(command.GuildID)
	if command.GuildID == "" || command.DailyEnergy < 0 || command.MaxEnergy < 0 {
		return domain.WorkConfig{}, domain.ErrInvalidWorkQuery
	}
	return s.adminRepo.SaveWorkConfig(ctx, command)
}

func (s Service) DeleteItem(ctx context.Context, command domain.WorkDeleteItemCommand) error {
	if s.adminRepo == nil {
		return domain.ErrWorkAdminUnavailable
	}
	command.GuildID = strings.TrimSpace(command.GuildID)
	command.Name = strings.TrimSpace(command.Name)
	if command.GuildID == "" || command.Name == "" {
		return domain.ErrInvalidWorkQuery
	}
	if _, err := s.Settings(ctx, command.GuildID); err != nil {
		return err
	}
	return s.adminRepo.DeleteWorkItem(ctx, command)
}

func (s Service) GrantEnergy(ctx context.Context, command domain.WorkEnergyGrantCommand) (domain.WorkUserState, error) {
	if s.adminRepo == nil {
		return domain.WorkUserState{}, domain.ErrWorkAdminUnavailable
	}
	command.GuildID = strings.TrimSpace(command.GuildID)
	command.UserID = strings.TrimSpace(command.UserID)
	if command.GuildID == "" || command.UserID == "" || command.Amount <= 0 {
		return domain.WorkUserState{}, domain.ErrInvalidWorkQuery
	}
	config, err := s.Settings(ctx, command.GuildID)
	if err != nil {
		return domain.WorkUserState{}, err
	}
	command.MaxEnergy = config.MaxEnergy
	return s.adminRepo.GrantWorkEnergy(ctx, command)
}

func (s Service) GrantEnergyToAll(ctx context.Context, command domain.WorkEnergyGrantAllCommand) (domain.WorkEnergyGrantAllResult, error) {
	if s.adminRepo == nil {
		return domain.WorkEnergyGrantAllResult{}, domain.ErrWorkAdminUnavailable
	}
	command.GuildID = strings.TrimSpace(command.GuildID)
	if command.GuildID == "" || command.Amount <= 0 {
		return domain.WorkEnergyGrantAllResult{}, domain.ErrInvalidWorkQuery
	}
	config, err := s.Settings(ctx, command.GuildID)
	if err != nil {
		return domain.WorkEnergyGrantAllResult{}, err
	}
	command.MaxEnergy = config.MaxEnergy
	return s.adminRepo.GrantWorkEnergyToAll(ctx, command)
}

func validateRequest(request InterfaceRequest) error {
	if strings.TrimSpace(request.GuildID) == "" || strings.TrimSpace(request.UserID) == "" {
		return domain.ErrInvalidWorkQuery
	}
	return nil
}

func visibleItems(items []domain.WorkItem, roleIDs []string) []domain.WorkItem {
	roles := make(map[string]struct{}, len(roleIDs))
	for _, roleID := range roleIDs {
		roleID = strings.TrimSpace(roleID)
		if roleID != "" {
			roles[roleID] = struct{}{}
		}
	}
	out := make([]domain.WorkItem, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.RoleID) == "" {
			out = append(out, item)
			continue
		}
		if _, ok := roles[item.RoleID]; ok {
			out = append(out, item)
		}
	}
	return out
}
