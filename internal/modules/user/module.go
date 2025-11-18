package user

import (
	"study1/internal/core/database"

	"github.com/gin-gonic/gin"
)

type UserModule struct {
	Repository UserRepository
	Service    UserService
	Handler    *UserHandler
}

func NewUserModule(db *database.DB) *UserModule {
	repo := NewUserRepository(db)
	service := NewUserService(repo)
	handler := NewUserHandler(service)

	return &UserModule{
		Repository: repo,
		Service:    service,
		Handler:    handler,
	}
}

func (m *UserModule) RegisterRoutes(router *gin.RouterGroup) {
	m.Handler.RegisterRoutes(router)
}
