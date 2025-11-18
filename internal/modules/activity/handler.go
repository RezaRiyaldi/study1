package activity

import (
	"net/http"

	"study1/internal/core/types"

	"github.com/gin-gonic/gin"
)

type ActivityHandler struct {
	service *ActivityService
}

func NewActivityHandler(service *ActivityService) *ActivityHandler {
	return &ActivityHandler{service: service}
}

func (h *ActivityHandler) RegisterRoutes(router *gin.RouterGroup) {
	g := router.Group("/activity-logs")
	{
		g.GET("", h.GetManys)
		g.GET("/:uuid", h.GetOnes)
	}
}

// List activity logs
// @Summary List activity logs
// @Description Retrieve paginated activity logs
// @Tags activity
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param search query string false "Search term"
// @Success 200 {object} types.Response
// @Failure 400 {object} types.Response
// @Failure 500 {object} types.Response
// @Router /activity-logs [get]
func (h *ActivityHandler) GetManys(c *gin.Context) {
	var params types.QueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err.Error()))
		return
	}
	params.SetDefaultPagination()

	logs, meta, err := h.service.GetManys(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse(logs, meta))
}

// @Summary Get an activity log
// @Description Get activity log by UUID
// @Tags activity
// @Accept json
// @Produce json
// @Param uuid path string true "Activity Log UUID"
// @Success 200 {object} types.Response
// @Failure 400 {object} types.Response
// @Failure 404 {object} types.Response
// @Router /activity-logs/{uuid} [get]
func (h *ActivityHandler) GetOnes(c *gin.Context) {
	uuid := c.Param("uuid")

	if uuid == "" {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse("Invalid UUID"))
		return
	}
	rec, err := h.service.GetOnes(uuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err.Error()))
		return
	}
	if rec == nil {
		c.JSON(http.StatusNotFound, types.NewErrorResponse("Not found"))
		return
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse(rec, nil))
}
