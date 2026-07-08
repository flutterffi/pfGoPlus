package role

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, item *Role) error
	Delete(ctx context.Context, name string) error
	FindByName(ctx context.Context, name string) (*Role, error)
	List(ctx context.Context) ([]Role, error)
	Update(ctx context.Context, item *Role) error
}

type GormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Create(ctx context.Context, item *Role) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *GormRepository) Delete(ctx context.Context, name string) error {
	return r.db.WithContext(ctx).Where("name = ?", name).Delete(&Role{}).Error
}

func (r *GormRepository) FindByName(ctx context.Context, name string) (*Role, error) {
	var item Role
	err := r.db.WithContext(ctx).Where("name = ?", name).Take(&item).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) List(ctx context.Context) ([]Role, error) {
	var items []Role
	err := r.db.WithContext(ctx).Order("id asc").Find(&items).Error
	return items, err
}

func (r *GormRepository) Update(ctx context.Context, item *Role) error {
	return r.db.WithContext(ctx).Save(item).Error
}
