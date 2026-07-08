package user

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, item *User) error
	FindByUsername(ctx context.Context, username string) (*User, error)
	List(ctx context.Context) ([]User, error)
}

type GormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Create(ctx context.Context, item *User) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *GormRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
	var item User
	err := r.db.WithContext(ctx).Where("username = ?", username).Take(&item).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) List(ctx context.Context) ([]User, error) {
	var items []User
	err := r.db.WithContext(ctx).Order("id asc").Find(&items).Error
	return items, err
}
