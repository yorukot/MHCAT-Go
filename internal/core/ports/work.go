package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

var (
	ErrWorkConfigMissing = errors.New("work config not found")
	ErrWorkItemsMissing  = errors.New("work items not found")
	ErrWorkUserMissing   = errors.New("work user not found")
	ErrWorkItemMissing   = errors.New("work item not found")
	ErrNoVisibleWorkItem = errors.New("no visible work item")
)

type WorkInterfaceRepository interface {
	GetWorkConfig(ctx context.Context, guildID string) (domain.WorkConfig, error)
	ListWorkItems(ctx context.Context, guildID string) ([]domain.WorkItem, error)
	GetWorkUser(ctx context.Context, guildID string, userID string) (domain.WorkUserState, error)
}

type WorkStartRepository interface {
	WorkInterfaceRepository
	EnsureWorkUser(ctx context.Context, guildID string, userID string, maxEnergy int64, maxEnergyText string) (domain.WorkUserState, error)
	StartWork(ctx context.Context, command domain.WorkStartCommand) (domain.WorkUserState, error)
}

type WorkAdminRepository interface {
	WorkStartRepository
	SaveWorkConfig(ctx context.Context, command domain.WorkConfigCommand) (domain.WorkConfig, error)
	DeleteWorkItem(ctx context.Context, command domain.WorkDeleteItemCommand) error
	GrantWorkEnergy(ctx context.Context, command domain.WorkEnergyGrantCommand) (domain.WorkUserState, error)
	GrantWorkEnergyToAll(ctx context.Context, command domain.WorkEnergyGrantAllCommand) (domain.WorkEnergyGrantAllResult, error)
}
