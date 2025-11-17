package app

import (
	"study1/internal/core/config"
	"study1/internal/core/database"
	"study1/internal/core/http"
	"study1/internal/modules/user"
)

type App struct {
	config *config.Config
	server *http.Server
	db     *database.DB
}

func New(cfg *config.Config) (*App, error) {
	// Initialize the database connection
	db, err := database.NewDB(cfg.Database)

	if err != nil {
		return nil, err
	}

	// Initialize the HTTP server
	server := http.NewServer(cfg)

	// TODO: Initilize Modules
	userModule := user.NewUserModule(db)

	userModule.Handler.RegisterRoutes(server.GetRouter().Group("/api/v1"))

	return &App{
		config: cfg,
		server: server,
		db:     db,
	}, nil
}

func (a *App) Start() error {
	// Start the HTTP server
	return a.server.Start(a.config.Server.Port)
}

func (a *App) GetDB() *database.DB {
	return a.db
}

func (a *App) GetConfig() *config.Config {
	return a.config
}
