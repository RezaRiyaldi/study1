package user

import (
	"errors"
	"study1/internal/core/types"
)

// UserService defines the business logic operations for users.
type UserService interface {
	GetManys(params types.QueryParams) ([]UserResponse, *types.Meta, error)
	GetOnes(uuid string) (*UserResponse, error)
	CreateManys(req []CreateUserRequest) ([]UserResponse, error)
	CreateOnes(req CreateUserRequest) (*UserResponse, error)
	UpdateManys(req []UpdateUserRequest) ([]UserResponse, error)
	UpdateOnes(uuid string, req UpdateUserRequest) (*UserResponse, error)
	DeleteManys(uuids []string) error
	DeleteOnes(uuid string) error
}

// userService implements the UserService interface.
type userService struct {
	repo UserRepository
}

// NewUserService creates a new instance of UserService.
func NewUserService(repo UserRepository) UserService {
	return &userService{repo: repo}
}

// GetManys users retrieves all users with pagination and filtering.
func (s *userService) GetManys(params types.QueryParams) ([]UserResponse, *types.Meta, error) {
	// Ensure pagination values are set
	params.SetDefaultPagination()

	users, meta, err := s.repo.FindManys(params)
	if err != nil {
		return nil, nil, err
	}

	responses := make([]UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}

	return responses, meta, nil
}

// GetOnes user by UUID retrieves a user by their UUID.
func (s *userService) GetOnes(uuid string) (*UserResponse, error) {
	user, err := s.repo.FindOnes(uuid)
	if err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

// CreateOnes creates a new user with the provided data.
func (s *userService) CreateOnes(req CreateUserRequest) (*UserResponse, error) {
	// Check if email already exists
	existingUser, _ := s.repo.FindByEmail(req.Email)
	if existingUser != nil {
		return nil, errors.New("email already exists")
	}

	user := &User{
		Name:  req.Name,
		Email: req.Email,
		Age:   req.Age,
	}

	if err := s.repo.CreateOnes(user); err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

// UpdateUser updates an existing user with the provided data.
func (s *userService) UpdateOnes(uuid string, req UpdateUserRequest) (*UserResponse, error) {
	user, err := s.repo.FindOnes(uuid)
	if err != nil {
		return nil, err
	}

	// Check if email is being updated and if it's already taken by another user
	if req.Email != "" && req.Email != user.Email {
		existingUser, _ := s.repo.FindByEmail(req.Email)
		if existingUser != nil && existingUser.ID != user.ID {
			return nil, errors.New("email already exists")
		}
		user.Email = req.Email
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Age > 0 {
		user.Age = req.Age
	}

	if err := s.repo.UpdateOnes(user); err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

// CreateManys creates multiple users by delegating to CreateOnes for each request.
func (s *userService) CreateManys(reqs []CreateUserRequest) ([]UserResponse, error) {
	responses := make([]UserResponse, 0, len(reqs))
	for _, r := range reqs {
		resp, err := s.CreateOnes(r)
		if err != nil {
			return nil, err
		}
		responses = append(responses, *resp)
	}
	return responses, nil
}

func (s *userService) UpdateManys(reqs []UpdateUserRequest) ([]UserResponse, error) {
	responses := make([]UserResponse, 0, len(reqs))
	for _, r := range reqs {
		resp, err := s.UpdateOnes(r.UUID, r)
		if err != nil {
			return nil, err
		}
		responses = append(responses, *resp)
	}

	return responses, nil
}

// DeleteManys deletes multiple users by calling DeleteOnes for each UUID.
func (s *userService) DeleteManys(uuids []string) error {
	for _, u := range uuids {
		if err := s.DeleteOnes(u); err != nil {
			return err
		}
	}
	return nil
}

// DeleteUser removes a user by their ID.
func (s *userService) DeleteOnes(uuid string) error {
	return s.repo.DeleteOnes(uuid)
}
