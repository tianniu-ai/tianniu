package repository

import (
	"context"
	"errors"

	"github.com/tianniu-ai/tianniu/pkg/model"
	"gorm.io/gorm"
)

func (r *Repository) SaveKVData(ctx context.Context, key, value string) error {
	kv := &model.KVData{
		Key:   key,
		Value: value,
	}
	return r.db.WithContext(ctx).Save(kv).Error
}

func (r *Repository) GetKVData(ctx context.Context, key string) (string, error) {
	var kv model.KVData
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&kv).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return kv.Value, nil
}
