package moderation

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

func TestDeleteDataServiceValidatesAndDeletes(t *testing.T) {
	repo := &deleteDataRepo{}
	service := DeleteDataService{Repository: repo}
	request := domain.DeleteDataRequest{GuildID: " guild-1 ", Target: domain.DeleteDataTargetAutoChat}
	if err := service.Delete(context.Background(), request); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if repo.request.GuildID != "guild-1" || repo.request.Target != domain.DeleteDataTargetAutoChat {
		t.Fatalf("request = %#v", repo.request)
	}
}

func TestDeleteDataServiceMapsInvalidAndMissing(t *testing.T) {
	service := DeleteDataService{}
	if err := service.Delete(context.Background(), domain.DeleteDataRequest{GuildID: "guild-1", Target: domain.DeleteDataTargetAutoChat}); !errors.Is(err, domain.ErrInvalidDeleteDataRequest) {
		t.Fatalf("nil repo err = %v", err)
	}
	repo := &deleteDataRepo{err: ports.ErrDeleteDataTargetMissing}
	service.Repository = repo
	err := service.Delete(context.Background(), domain.DeleteDataRequest{GuildID: "guild-1", Target: domain.DeleteDataTargetAutoChat})
	if !errors.Is(err, ports.ErrDeleteDataTargetMissing) {
		t.Fatalf("missing err = %v", err)
	}
}

type deleteDataRepo struct {
	request domain.DeleteDataRequest
	err     error
}

func (r *deleteDataRepo) DeleteGuildConfig(ctx context.Context, request domain.DeleteDataRequest) error {
	r.request = request
	return r.err
}
