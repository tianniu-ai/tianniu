package repository_storage

import (
	"context"

	"github.com/tianniu-ai/tianniu/pkg/repository"
)

type RepositoryStorage struct {
	repo *repository.Repository
}

func NewRepositoryStorage(repo *repository.Repository) *RepositoryStorage {
	return &RepositoryStorage{
		repo: repo,
	}
}

func (r *RepositoryStorage) Load(ctx context.Context, key string) (string, error) {
	return r.repo.GetKVData(ctx, key)
}

func (r *RepositoryStorage) Store(ctx context.Context, key string, value string) error {
	return r.repo.SaveKVData(ctx, key, value)
}
