package user

import "time"

// CreateUserRequest represents the data required to create a new user.
// @Description Payload to create a new user
type CreateUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Age   int    `json:"age" binding:"min=0"`
}

// UpdateUserRequest represents the data required to update an existing user.
// @Description Payload to update an existing user
type UpdateUserRequest struct {
	UUID  string `json:"uuid" binding:"required,uuid4"`
	Name  string `json:"name"`
	Email string `json:"email" binding:"omitempty,email"`
	Age   int    `json:"age" binding:"min=0"`
}

// UserResponse represents the user data returned in API responses.
// @Description User data returned by the API
type UserResponse struct {
	ID        uint      `json:"id"`
	UUID      string    `json:"uuid"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts a User model to a UserResponse DTO.
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		UUID:      u.UUID,
		Name:      u.Name,
		Email:     u.Email,
		Age:       u.Age,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
