package activity

import (
	"study1/internal/core/database"

	"github.com/gin-gonic/gin"
)

type ActivityModule struct {
	Repository ActivityRepository
	Service    *ActivityService
	Handler    *ActivityHandler
}

func NewActivityModule(db *database.DB) *ActivityModule {
	repo := NewActivityRepository(db)
	service := NewActivityService(repo)
	handler := NewActivityHandler(service)

	return &ActivityModule{
		Repository: repo,
		Service:    service,
		Handler:    handler,
	}
}

func (m *ActivityModule) RegisterRoutes(router *gin.RouterGroup) {
	m.Handler.RegisterRoutes(router)
}
