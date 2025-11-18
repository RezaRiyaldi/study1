package user

import (
	"study1/internal/core/database"
	"study1/internal/core/repository"
	"study1/internal/core/types"
)

// UserRepository defines the interface for user data operations.
type UserRepository interface {
	FindManys(params types.QueryParams) ([]User, *types.Meta, error)
	FindOnes(uuid string) (*User, error)
	FindByEmail(email string) (*User, error)
	CreateOnes(user *User) error
	UpdateOnes(user *User) error
	DeleteOnes(uuid string) error
}

// userRepository implements the UserRepository interface.
type userRepository struct {
	genericRepo *repository.GenericRepository[User]
	db          *database.DB
}

// NewUserRepository creates a new instance of UserRepository.
func NewUserRepository(db *database.DB) UserRepository {
	return &userRepository{
		genericRepo: repository.NewGenericRepositoryWithSoftDelete[User](db, true),
		db:          db,
	}
}

// FindAll retrieves all users with pagination and filtering.
func (r *userRepository) FindManys(params types.QueryParams) ([]User, *types.Meta, error) {
	return r.genericRepo.FindManys(params)
}

// FindByUUID retrieves a user by their UUID using the generic repository.
func (r *userRepository) FindOnes(uuid string) (*User, error) {
	return r.genericRepo.FindOnes(uuid)
}

// FindByEmail retrieves a user by their email address.
func (r *userRepository) FindByEmail(email string) (*User, error) {
	var user User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create adds a new user to the database.
func (r *userRepository) CreateOnes(user *User) error {
	return r.genericRepo.CreateOnes(user)
}

// Update modifies an existing user in the database.
func (r *userRepository) UpdateOnes(user *User) error {
	return r.genericRepo.UpdateOnes(user)
}

// DeleteOnes removes a user by their UUID using the generic repository.
func (r *userRepository) DeleteOnes(uuid string) error {
	return r.genericRepo.DeleteOnes(uuid)
}
