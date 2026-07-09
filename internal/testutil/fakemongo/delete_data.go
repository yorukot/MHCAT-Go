package fakemongo

import (
	"context"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type DeleteDataRepository struct {
	Existing map[string]map[domain.DeleteDataTarget]bool
	Deleted  []domain.DeleteDataRequest
	Err      error
}

func NewDeleteDataRepository() *DeleteDataRepository {
	return &DeleteDataRepository{Existing: map[string]map[domain.DeleteDataTarget]bool{}}
}

func (r *DeleteDataRepository) Put(guildID string, target domain.DeleteDataTarget) {
	if r.Existing == nil {
		r.Existing = map[string]map[domain.DeleteDataTarget]bool{}
	}
	if r.Existing[guildID] == nil {
		r.Existing[guildID] = map[domain.DeleteDataTarget]bool{}
	}
	r.Existing[guildID][target] = true
}

func (r *DeleteDataRepository) DeleteGuildConfig(ctx context.Context, request domain.DeleteDataRequest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r.Err != nil {
		return r.Err
	}
	request = request.Normalize()
	if err := request.Validate(); err != nil {
		return err
	}
	if r.Existing == nil || r.Existing[request.GuildID] == nil || !r.Existing[request.GuildID][request.Target] {
		return ports.ErrDeleteDataTargetMissing
	}
	delete(r.Existing[request.GuildID], request.Target)
	r.Deleted = append(r.Deleted, request)
	return nil
}

var _ ports.DeleteDataRepository = (*DeleteDataRepository)(nil)
