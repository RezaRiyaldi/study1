package user

import (
	"net/http"

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
		users.GET("", h.GetManys)
		users.GET(":uuid", h.GetOnes)
		users.POST("", h.CreateOnes)
		users.PUT(":uuid", h.UpdateOnes)
		users.DELETE(":uuid", h.DeleteOnes)
	}
}

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
func (h *UserHandler) GetManys(c *gin.Context) {
	var params types.QueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err.Error()))
		return
	}

	params.SetDefaultPagination()

	users, meta, err := h.service.GetManys(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse(users, meta))
}

// @Summary Get a user
// @Description Get user by UUID
// @Tags users
// @Accept json
// @Produce json
// @Param uuid path string true "User UUID"
// @Success 200 {object} types.Response
// @Failure 400 {object} types.Response
// @Failure 404 {object} types.Response
// @Router /users/{uuid} [get]
func (h *UserHandler) GetOnes(c *gin.Context) {
	uuid := c.Param("uuid")
	if uuid == "" {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse("Invalid user UUID"))
		return
	}

	user, err := h.service.GetOnes(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, types.NewErrorResponse("User not found"))
		return
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse(user, nil))
}

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
func (h *UserHandler) CreateOnes(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err.Error()))
		return
	}

	user, err := h.service.CreateOnes(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, types.NewSuccessResponse(user, nil))
}

// @Summary Update a user
// @Description Update user by UUID
// @Tags users
// @Accept json
// @Produce json
// @Param uuid path string true "User UUID"
// @Param body body UpdateUserRequest true "Update user payload"
// @Success 200 {object} types.Response
// @Failure 400 {object} types.Response
// @Failure 404 {object} types.Response
// @Router /users/{uuid} [put]
func (h *UserHandler) UpdateOnes(c *gin.Context) {
	uuid := c.Param("uuid")
	if uuid == "" {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse("Invalid user UUID"))
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err.Error()))
		return
	}

	user, err := h.service.UpdateOnes(uuid, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse(user, nil))
}

// @Summary Delete a user
// @Description Delete user by UUID
// @Tags users
// @Accept json
// @Produce json
// @Param uuid path string true "User UUID"
// @Success 200 {object} types.Response
// @Failure 400 {object} types.Response
// @Failure 500 {object} types.Response
// @Router /users/{uuid} [delete]
func (h *UserHandler) DeleteOnes(c *gin.Context) {
	uuid := c.Param("uuid")
	if uuid == "" {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse("Invalid user UUID"))
		return
	}

	if err := h.service.DeleteOnes(uuid); err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse("User deleted successfully", nil))
}
