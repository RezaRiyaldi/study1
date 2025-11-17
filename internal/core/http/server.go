package http

import (
	"net/http"
	"study1/internal/core/config"

	"github.com/gin-gonic/gin"
)

type Server struct {
	router *gin.Engine
	config *config.Config
}

func NewServer(cfg *config.Config) *Server {
	if cfg.Server.Environtment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Basic Middleware
	router.Use(
		gin.Logger(),
		gin.Recovery(),
	)

	// Health Check Endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "OK",
			"environment": cfg.Server.Environtment,
		})
	})

	// API Routes Group
	api := router.Group("/api/v1")
	{
		api.GET("/", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{
				"message": "Welcome to Study1 API",
			})
		})
		// TODO: Add module routes here
		// Example: userModule.RegisterRoutes(api)
	}

	return &Server{
		router: router,
		config: cfg,
	}
}

func (s *Server) Start(port string) error {
	return s.router.Run(":" + port)
}

func (s *Server) GetRouter() *gin.Engine {
	return s.router
}
