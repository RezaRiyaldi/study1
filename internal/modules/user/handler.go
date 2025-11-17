package user

import (
	"net/http"
	"strconv"
	"study1/internal/core/types"

	"github.com/gin-gonic/gin"
)

// UserHandler handles HTTP requests for user operations.
type UserHandler struct {
	service UserService
}

// NewUserHandler creates a new instance of UserHandler.
func NewUserHandler(service UserService) *UserHandler {
	return &UserHandler{service: service}
}

// RegisterRoutes registers all user-related routes with the router.
func (h *UserHandler) RegisterRoutes(router *gin.RouterGroup) {
	users := router.Group("/users")
	{
		users.GET("", h.GetAllUsers)
		users.GET("/:id", h.GetUserByID)
		users.POST("", h.CreateUser)
		users.PUT("/:id", h.UpdateUser)
		users.DELETE("/:id", h.DeleteUser)
	}
}

// GetAllUsers handles GET /api/v1/users
// Retrieves all users with pagination and filtering.
// @Summary Get all users
// @Description Retrieves paginated list of users with optional filters
// @Tags users
// @Accept json
// @Produce json
// @Param search query string false "Search term"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} types.Response
// @Failure 400 {object} types.Response
// @Failure 500 {object} types.Response
// @Router /users [get]
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	var params types.QueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err.Error()))
		return
	}

	// Set default pagination values
	params.SetDefaultPagination()

	users, meta, err := h.service.GetAllUsers(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse(users, meta))
}

// GetUserByID handles GET /api/v1/users/:id
// Retrieves a specific user by ID.
// @Summary Get a user
// @Description Get user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} types.Response
// @Failure 400 {object} types.Response
// @Failure 404 {object} types.Response
// @Router /users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse("Invalid user ID"))
		return
	}

	user, err := h.service.GetUserByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, types.NewErrorResponse("User not found"))
		return
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse(user, nil))
}

// CreateUser handles POST /api/v1/users
// Creates a new user with the provided data.
// @Summary Create a user
// @Description Create a new user
// @Tags users
// @Accept json
// @Produce json
// @Param body body CreateUserRequest true "Create user payload"
// @Success 201 {object} types.Response
// @Failure 400 {object} types.Response
// @Failure 500 {object} types.Response
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err.Error()))
		return
	}

	user, err := h.service.CreateUser(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, types.NewSuccessResponse(user, nil))
}

// UpdateUser handles PUT /api/v1/users/:id
// Updates an existing user with the provided data.
// @Summary Update a user
// @Description Update user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param body body UpdateUserRequest true "Update user payload"
// @Success 200 {object} types.Response
// @Failure 400 {object} types.Response
// @Failure 404 {object} types.Response
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse("Invalid user ID"))
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err.Error()))
		return
	}

	user, err := h.service.UpdateUser(uint(id), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse(user, nil))
}

// DeleteUser handles DELETE /api/v1/users/:id
// Deletes a user by ID.
// @Summary Delete a user
// @Description Delete user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} types.Response
// @Failure 400 {object} types.Response
// @Failure 500 {object} types.Response
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse("Invalid user ID"))
		return
	}

	if err := h.service.DeleteUser(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse("User deleted successfully", nil))
}
