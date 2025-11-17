package user

import (
	"errors"
	"study1/internal/core/types"
)

// UserService defines the business logic operations for users.
type UserService interface {
	GetAllUsers(params types.QueryParams) ([]UserResponse, *types.Meta, error)
	GetUserByID(id uint) (*UserResponse, error)
	CreateUser(req CreateUserRequest) (*UserResponse, error)
	UpdateUser(id uint, req UpdateUserRequest) (*UserResponse, error)
	DeleteUser(id uint) error
}

// userService implements the UserService interface.
type userService struct {
	repo UserRepository
}

// NewUserService creates a new instance of UserService.
func NewUserService(repo UserRepository) UserService {
	return &userService{repo: repo}
}

// GetAllUsers retrieves all users with pagination and filtering.
func (s *userService) GetAllUsers(params types.QueryParams) ([]UserResponse, *types.Meta, error) {
	// Ensure pagination values are set
	params.SetDefaultPagination()

	users, meta, err := s.repo.FindAll(params)
	if err != nil {
		return nil, nil, err
	}

	responses := make([]UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}

	return responses, meta, nil
}

// GetUserByID retrieves a user by their ID.
func (s *userService) GetUserByID(id uint) (*UserResponse, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

// CreateUser creates a new user with the provided data.
func (s *userService) CreateUser(req CreateUserRequest) (*UserResponse, error) {
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

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

// UpdateUser updates an existing user with the provided data.
func (s *userService) UpdateUser(id uint, req UpdateUserRequest) (*UserResponse, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Check if email is being updated and if it's already taken by another user
	if req.Email != "" && req.Email != user.Email {
		existingUser, _ := s.repo.FindByEmail(req.Email)
		if existingUser != nil && existingUser.ID != id {
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

	if err := s.repo.Update(user); err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

// DeleteUser removes a user by their ID.
func (s *userService) DeleteUser(id uint) error {
	return s.repo.Delete(id)
}
