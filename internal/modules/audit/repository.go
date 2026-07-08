package audit

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, item *Log) error
	List(ctx context.Context, limit int) ([]Log, error)
}

type GormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Create(ctx context.Context, item *Log) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *GormRepository) List(ctx context.Context, limit int) ([]Log, error) {
	if limit <= 0 {
		limit = 50
	}
	var items []Log
	err := r.db.WithContext(ctx).Order("id desc").Limit(limit).Find(&items).Error
	return items, err
}
