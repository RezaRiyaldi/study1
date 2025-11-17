package user

import "study1/internal/core/database"

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
