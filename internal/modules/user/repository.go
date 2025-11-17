package user

import (
	"study1/internal/core/database"
	"study1/internal/core/repository"
	"study1/internal/core/types"
)

// UserRepository defines the interface for user data operations.
type UserRepository interface {
	FindAll(params types.QueryParams) ([]User, *types.Meta, error)
	FindByID(id uint) (*User, error)
	FindByEmail(email string) (*User, error)
	Create(user *User) error
	Update(user *User) error
	Delete(id uint) error
}

// userRepository implements the UserRepository interface.
type userRepository struct {
	genericRepo *repository.GenericRepository[User]
	db          *database.DB
}

// NewUserRepository creates a new instance of UserRepository.
func NewUserRepository(db *database.DB) UserRepository {
	return &userRepository{
		genericRepo: repository.NewGenericRepository[User](db),
		db:          db,
	}
}

// FindAll retrieves all users with pagination and filtering.
func (r *userRepository) FindAll(params types.QueryParams) ([]User, *types.Meta, error) {
	return r.genericRepo.FindAll(params)
}

// FindByID retrieves a user by their ID.
func (r *userRepository) FindByID(id uint) (*User, error) {
	return r.genericRepo.FindByID(id)
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
func (r *userRepository) Create(user *User) error {
	return r.genericRepo.Create(user)
}

// Update modifies an existing user in the database.
func (r *userRepository) Update(user *User) error {
	return r.genericRepo.Update(user)
}

// Delete removes a user from the database by ID.
func (r *userRepository) Delete(id uint) error {
	return r.genericRepo.Delete(id)
}
