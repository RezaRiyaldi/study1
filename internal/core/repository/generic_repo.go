package repository

import (
	"time"

	"study1/internal/core/database"
	"study1/internal/core/types"
)

type GenericRepository[T any] struct {
	db         *database.DB
	softDelete bool
}

// NewGenericRepository creates a repository without soft-delete behavior.
func NewGenericRepository[T any](db *database.DB) *GenericRepository[T] {
	return &GenericRepository[T]{db: db, softDelete: false}
}

// NewGenericRepositoryWithSoftDelete creates a repository with soft-delete behavior enabled.
func NewGenericRepositoryWithSoftDelete[T any](db *database.DB, softDelete bool) *GenericRepository[T] {
	return &GenericRepository[T]{db: db, softDelete: softDelete}
}

func (r *GenericRepository[T]) FindManys(params types.QueryParams) ([]T, *types.Meta, error) {
	var models []T
	var total int64

	// Ensure pagination values are set
	params.SetDefaultPagination()

	// Count total records (before pagination)
	var countModel T
	countQuery := r.db.Model(&countModel)
	if r.softDelete {
		countQuery = countQuery.Where("deleted_at IS NULL")
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, nil, err
	}

	// Build query with all conditions
	query := database.NewQueryBuilder[T](r.db.DB, params).Build()
	if r.softDelete {
		query = query.Where("deleted_at IS NULL")
	}

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

func (r *GenericRepository[T]) FindOnes(uuid string) (*T, error) {
	var model T
	q := r.db.Model(&model)
	if r.softDelete {
		q = q.Where("deleted_at IS NULL")
	}
	if err := q.First(&model, "uuid = ?", uuid).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

func (r *GenericRepository[T]) CreateManys(models []T) error {
	return r.db.Create(&models).Error
}

func (r *GenericRepository[T]) CreateOnes(model *T) error {
	return r.db.Create(model).Error
}

func (r *GenericRepository[T]) UpdateManys(models []T) error {
	return r.db.Save(&models).Error
}

func (r *GenericRepository[T]) UpdateOnes(model *T) error {
	return r.db.Save(model).Error
}

func (r *GenericRepository[T]) DeleteManys(uuids []string) error {
	return r.DeleteManysWithActor(uuids, nil)
}

// DeleteManysWithActor deletes multiple records. If soft-deletes are enabled,
// it updates `deleted_at` and `deleted_by` instead of hard-deleting.
func (r *GenericRepository[T]) DeleteManysWithActor(uuids []string, deletedBy *uint) error {
	var model T
	if r.softDelete {
		data := map[string]interface{}{"deleted_at": time.Now()}
		if deletedBy != nil {
			data["deleted_by"] = deletedBy
		}
		return r.db.Model(&model).Where("uuid in (?)", uuids).Updates(data).Error
	}
	return r.db.Delete(&model, "uuid in (?)", uuids).Error
}

func (r *GenericRepository[T]) DeleteOnes(uuid string) error {
	return r.DeleteOnesWithActor(uuid, nil)
}

// DeleteOnesWithActor deletes a record by UUID. If soft-deletes are enabled,
// it updates `deleted_at` and `deleted_by` instead of hard-deleting.
func (r *GenericRepository[T]) DeleteOnesWithActor(uuid string, deletedBy *uint) error {
	var model T
	if r.softDelete {
		data := map[string]interface{}{"deleted_at": time.Now()}
		if deletedBy != nil {
			data["deleted_by"] = deletedBy
		}
		return r.db.Model(&model).Where("uuid = ?", uuid).Updates(data).Error
	}
	return r.db.Delete(&model, "uuid = ?", uuid).Error
}

// Compatibility wrappers for previously-named methods used across the codebase.
func (r *GenericRepository[T]) FindAll(params types.QueryParams) ([]T, *types.Meta, error) {
	return r.FindManys(params)
}

func (r *GenericRepository[T]) FindOne(uuid string) (*T, error) {
	return r.FindOnes(uuid)
}

func (r *GenericRepository[T]) Create(model *T) error {
	return r.CreateOnes(model)
}

func (r *GenericRepository[T]) Update(model *T) error {
	return r.UpdateOnes(model)
}

func (r *GenericRepository[T]) Delete(uuid string) error {
	return r.DeleteOnes(uuid)
}

func (r *GenericRepository[T]) Count() (int64, error) {
	var model T
	var count int64
	q := r.db.Model(&model)
	if r.softDelete {
		q = q.Where("deleted_at IS NULL")
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
