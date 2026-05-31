package todo

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, item *Todo) error
	List(ctx context.Context) ([]Todo, error)
}

type GormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Create(ctx context.Context, item *Todo) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *GormRepository) List(ctx context.Context) ([]Todo, error) {
	var items []Todo
	err := r.db.WithContext(ctx).Order("id desc").Find(&items).Error
	return items, err
}
