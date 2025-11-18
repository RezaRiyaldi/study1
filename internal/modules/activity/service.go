package activity

import (
	"study1/internal/core/types"
)

type ActivityService struct {
	repo ActivityRepository
}

func NewActivityService(repo ActivityRepository) *ActivityService {
	return &ActivityService{repo: repo}
}

func (s *ActivityService) GetManys(params types.QueryParams) ([]ActivityLog, *types.Meta, error) {
	params.SetDefaultPagination()

	logs, meta, err := s.repo.FindManys(params)
	if err != nil {
		return nil, nil, err
	}

	return logs, meta, nil
}

func (s *ActivityService) GetOnes(uuid string) (*ActivityLog, error) {
	return s.repo.FindOnes(uuid)
}
