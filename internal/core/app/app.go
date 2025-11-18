package app

import (
	"study1/internal/core/config"
	"study1/internal/core/database"
	"study1/internal/core/http"
	"study1/internal/modules/activity"
	"study1/internal/modules/user"
)

type App struct {
	config *config.Config
	server *http.Server
	db     *database.DB
}

func New(cfg *config.Config) (*App, error) {
	db, err := database.NewDB(cfg.Database)
	if err != nil {
		return nil, err
	}

	// Initialize modules
	userModule := user.NewUserModule(db)
	activityModule := activity.NewActivityModule(db)

	// Pass semua modules ke server (provide DB for middleware)
	server := http.NewServer(cfg, db, userModule, activityModule) //, otherModule, anotherModule)

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
