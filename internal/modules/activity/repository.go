package activity

import (
	"study1/internal/core/database"
	"study1/internal/core/repository"
	"study1/internal/core/types"
)

type ActivityRepository struct {
	genericRepo *repository.GenericRepository[ActivityLog]
	db          *database.DB
}

func NewActivityRepository(db *database.DB) ActivityRepository {
	return ActivityRepository{
		genericRepo: repository.NewGenericRepository[ActivityLog](db),
		db:          db,
	}
}

// List returns activity logs with pagination.
func (r *ActivityRepository) FindManys(params types.QueryParams) ([]ActivityLog, *types.Meta, error) {
	return r.genericRepo.FindManys(params)
}

// GetOnes returns a single activity log by UUID.
func (r *ActivityRepository) FindOnes(uuid string) (*ActivityLog, error) {
	return r.genericRepo.FindOnes(uuid)
}

func (r *ActivityRepository) CreateManys(logs []ActivityLog) error {
	return r.genericRepo.CreateManys(logs)
}

func (r *ActivityRepository) CreateOnes(log *ActivityLog) error {
	return r.genericRepo.CreateOnes(log)
}

func (r *ActivityRepository) UpdateManys(logs []ActivityLog) error {
	return r.genericRepo.UpdateManys(logs)
}

func (r *ActivityRepository) UpdateOnes(log *ActivityLog) error {
	return r.genericRepo.UpdateOnes(log)
}

func (r *ActivityRepository) DeleteManys(uuids []string) error {
	return r.genericRepo.DeleteManys(uuids)
}

func (r *ActivityRepository) DeleteOnes(uuid string) error {
	return r.genericRepo.DeleteOnes(uuid)
}
