package fakemongo

import (
	"context"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type WorkInterfaceRepository struct {
	Configs map[string]domain.WorkConfig
	Items   map[string][]domain.WorkItem
	Users   map[string]domain.WorkUserState
}

func NewWorkInterfaceRepository() *WorkInterfaceRepository {
	return &WorkInterfaceRepository{
		Configs: map[string]domain.WorkConfig{},
		Items:   map[string][]domain.WorkItem{},
		Users:   map[string]domain.WorkUserState{},
	}
}

func (r *WorkInterfaceRepository) PutConfig(config domain.WorkConfig) {
	r.Configs[strings.TrimSpace(config.GuildID)] = config
}

func (r *WorkInterfaceRepository) PutItems(guildID string, items ...domain.WorkItem) {
	r.Items[strings.TrimSpace(guildID)] = append([]domain.WorkItem(nil), items...)
}

func (r *WorkInterfaceRepository) PutUser(user domain.WorkUserState) {
	r.Users[workUserKey(user.GuildID, user.UserID)] = user
}

func (r *WorkInterfaceRepository) GetWorkConfig(_ context.Context, guildID string) (domain.WorkConfig, error) {
	config, ok := r.Configs[strings.TrimSpace(guildID)]
	if !ok {
		return domain.WorkConfig{}, ports.ErrWorkConfigMissing
	}
	return config, nil
}

func (r *WorkInterfaceRepository) ListWorkItems(_ context.Context, guildID string) ([]domain.WorkItem, error) {
	items := append([]domain.WorkItem(nil), r.Items[strings.TrimSpace(guildID)]...)
	if len(items) == 0 {
		return nil, ports.ErrWorkItemsMissing
	}
	return items, nil
}

func (r *WorkInterfaceRepository) GetWorkUser(_ context.Context, guildID string, userID string) (domain.WorkUserState, error) {
	user, ok := r.Users[workUserKey(guildID, userID)]
	if !ok {
		return domain.WorkUserState{}, ports.ErrWorkUserMissing
	}
	return user, nil
}

func (r *WorkInterfaceRepository) StartWork(_ context.Context, command domain.WorkStartCommand) (domain.WorkUserState, error) {
	command.GuildID = strings.TrimSpace(command.GuildID)
	command.UserID = strings.TrimSpace(command.UserID)
	command.WorkName = strings.TrimSpace(command.WorkName)
	if command.GuildID == "" ||
		command.UserID == "" ||
		command.WorkName == "" ||
		command.DurationSec <= 0 ||
		command.EnergyCost < 0 ||
		command.CoinReward < 0 ||
		command.NowUnix <= 0 {
		return domain.WorkUserState{}, domain.ErrInvalidWorkQuery
	}
	key := workUserKey(command.GuildID, command.UserID)
	user, ok := r.Users[key]
	if !ok {
		user = domain.WorkUserState{
			GuildID:     command.GuildID,
			UserID:      command.UserID,
			State:       domain.WorkIdleState,
			EndTimeUnix: 0,
			Energy:      command.MaxEnergy,
			GetCoin:     0,
			Initialized: false,
		}
	}
	if user.Energy < command.EnergyCost {
		return domain.WorkUserState{}, domain.ErrWorkEnergyInsufficient
	}
	if !command.Override && user.State != domain.WorkIdleState {
		return domain.WorkUserState{}, domain.ErrWorkAlreadyBusy
	}
	user.GuildID = command.GuildID
	user.UserID = command.UserID
	user.State = command.WorkName
	user.EndTimeUnix = command.NowUnix + command.DurationSec
	user.Energy -= command.EnergyCost
	user.GetCoin = command.CoinReward
	user.Initialized = true
	r.Users[key] = user
	return user, nil
}

func (r *WorkInterfaceRepository) EnsureWorkUser(_ context.Context, guildID string, userID string, maxEnergy int64, maxEnergyText string) (domain.WorkUserState, error) {
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return domain.WorkUserState{}, domain.ErrInvalidWorkQuery
	}
	key := workUserKey(guildID, userID)
	if user, ok := r.Users[key]; ok {
		return user, nil
	}
	user := domain.WorkUserState{GuildID: guildID, UserID: userID, State: domain.WorkIdleState, Energy: maxEnergy, Initialized: true}
	r.Users[key] = user
	return user, nil
}

func (r *WorkInterfaceRepository) SaveWorkConfig(_ context.Context, command domain.WorkConfigCommand) (domain.WorkConfig, error) {
	command.GuildID = strings.TrimSpace(command.GuildID)
	if command.GuildID == "" {
		return domain.WorkConfig{}, domain.ErrInvalidWorkQuery
	}
	config := domain.WorkConfig{
		GuildID:     command.GuildID,
		DailyEnergy: command.DailyEnergy,
		MaxEnergy:   command.MaxEnergy,
		Captcha:     command.Captcha,
	}
	r.Configs[command.GuildID] = config
	return config, nil
}

func (r *WorkInterfaceRepository) DeleteWorkItem(_ context.Context, command domain.WorkDeleteItemCommand) error {
	command.GuildID = strings.TrimSpace(command.GuildID)
	command.Name = strings.TrimSpace(command.Name)
	if command.GuildID == "" || command.Name == "" {
		return domain.ErrInvalidWorkQuery
	}
	items := r.Items[command.GuildID]
	for index, item := range items {
		if item.Name == command.Name {
			r.Items[command.GuildID] = append(append([]domain.WorkItem(nil), items[:index]...), items[index+1:]...)
			return nil
		}
	}
	return ports.ErrWorkItemMissing
}

func (r *WorkInterfaceRepository) GrantWorkEnergy(_ context.Context, command domain.WorkEnergyGrantCommand) (domain.WorkUserState, error) {
	command.GuildID = strings.TrimSpace(command.GuildID)
	command.UserID = strings.TrimSpace(command.UserID)
	if command.GuildID == "" || command.UserID == "" {
		return domain.WorkUserState{}, domain.ErrInvalidWorkQuery
	}
	key := workUserKey(command.GuildID, command.UserID)
	user, ok := r.Users[key]
	if !ok {
		user = domain.WorkUserState{
			GuildID:     command.GuildID,
			UserID:      command.UserID,
			State:       domain.WorkIdleState,
			EndTimeUnix: 0,
			Energy:      command.MaxEnergy,
			GetCoin:     0,
			Initialized: true,
		}
	} else {
		user.Energy = clampWorkEnergy(user.Energy+command.Amount, command.MaxEnergy)
		user.Initialized = true
	}
	user.GuildID = command.GuildID
	user.UserID = command.UserID
	r.Users[key] = user
	return user, nil
}

func (r *WorkInterfaceRepository) GrantWorkEnergyToAll(_ context.Context, command domain.WorkEnergyGrantAllCommand) (domain.WorkEnergyGrantAllResult, error) {
	command.GuildID = strings.TrimSpace(command.GuildID)
	if command.GuildID == "" {
		return domain.WorkEnergyGrantAllResult{}, domain.ErrInvalidWorkQuery
	}
	var result domain.WorkEnergyGrantAllResult
	for key, user := range r.Users {
		if user.GuildID != command.GuildID {
			continue
		}
		result.Matched++
		next := clampWorkEnergy(user.Energy+command.Amount, command.MaxEnergy)
		if next != user.Energy {
			result.Modified++
		}
		user.Energy = next
		r.Users[key] = user
	}
	return result, nil
}

func workUserKey(guildID string, userID string) string {
	return strings.TrimSpace(guildID) + "\x00" + strings.TrimSpace(userID)
}

func clampWorkEnergy(value int64, maxEnergy int64) int64 {
	if value > maxEnergy {
		return maxEnergy
	}
	return value
}

var _ ports.WorkInterfaceRepository = (*WorkInterfaceRepository)(nil)
var _ ports.WorkStartRepository = (*WorkInterfaceRepository)(nil)
var _ ports.WorkAdminRepository = (*WorkInterfaceRepository)(nil)
