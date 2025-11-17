package repository

import (
	"study1/internal/core/database"
	"study1/internal/core/types"
)

type GenericRepository[T any] struct {
	db *database.DB
}

func NewGenericRepository[T any](db *database.DB) *GenericRepository[T] {
	return &GenericRepository[T]{db: db}
}

func (r *GenericRepository[T]) FindAll(params types.QueryParams) ([]T, *types.Meta, error) {
	var models []T
	var total int64

	// Ensure pagination values are set
	params.SetDefaultPagination()

	// Count total records (before pagination)
	var countModel T
	if err := r.db.Model(&countModel).Count(&total).Error; err != nil {
		return nil, nil, err
	}

	// Build query with all conditions
	query := database.NewQueryBuilder[T](r.db.DB, params).Build()

	// Execute query with pagination
	offset := (params.Page - 1) * params.PageSize
	if err := query.Offset(offset).Limit(params.PageSize).Find(&models).Error; err != nil {
		return nil, nil, err
	}

	// Calculate total pages
	pages := (int(total) + params.PageSize - 1) / params.PageSize

	meta := &types.Meta{
		Page:     params.Page,
		PageSize: params.PageSize,
		Total:    int(total),
		Pages:    pages,
	}

	return models, meta, nil
}

func (r *GenericRepository[T]) FindByID(id uint) (*T, error) {
	var model T
	if err := r.db.First(&model, id).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

func (r *GenericRepository[T]) Create(model *T) error {
	return r.db.Create(model).Error
}

func (r *GenericRepository[T]) Update(model *T) error {
	return r.db.Save(model).Error
}

func (r *GenericRepository[T]) Delete(id uint) error {
	var model T
	return r.db.Delete(&model, id).Error
}

func (r *GenericRepository[T]) Count() (int64, error) {
	var model T
	var count int64
	if err := r.db.Model(&model).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
