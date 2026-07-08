package audit

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, item *Log) error
	List(ctx context.Context, query ListQuery) ([]Log, int64, error)
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

func (r *GormRepository) List(ctx context.Context, query ListQuery) ([]Log, int64, error) {
	query = normalizeListQuery(query)

	db := r.db.WithContext(ctx).Model(&Log{})
	db = applyListFilters(db, query)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []Log
	err := db.Order("id desc").Limit(query.Limit).Offset(query.Offset).Find(&items).Error
	return items, total, err
}

func applyListFilters(db *gorm.DB, query ListQuery) *gorm.DB {
	if query.ActorUsername != "" {
		db = db.Where("actor_username = ?", query.ActorUsername)
	}
	if query.Action != "" {
		db = db.Where("action = ?", query.Action)
	}
	if query.Resource != "" {
		db = db.Where("resource = ?", query.Resource)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.TraceID != "" {
		db = db.Where("trace_id = ?", query.TraceID)
	}
	return db
}
