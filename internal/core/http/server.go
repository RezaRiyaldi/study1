package http

import (
	"net/http"
	"study1/internal/core/config"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	router *gin.Engine
	config *config.Config
}

// Interface untuk module yang bisa register routes
type RouteRegistrar interface {
	RegisterRoutes(router *gin.RouterGroup)
}

func NewServer(cfg *config.Config, modules ...RouteRegistrar) *Server {
	router := gin.Default()

	// Middleware
	router.Use(gin.Logger(), gin.Recovery())

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	api := router.Group("/api/v1")
	{
		api.GET("/", apiRoot(cfg))
		api.GET("/app-information", apiAppInformation(cfg))
		api.GET("/health", apiHealth(cfg))

		// Register semua modules
		for _, module := range modules {
			module.RegisterRoutes(api)
		}
	}

	return &Server{router: router, config: cfg}
}

func (s *Server) Start(port string) error {
	return s.router.Run(":" + port)
}

func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

// apiVersion returns a handler that reports the current API version.
// This is used at runtime and for docs.
// @Summary API Version
// @Description Get API version information
// @Tags general
// @Produce json
// @Success 200 {object} map[string]string
// @Router /app-information [get]
func apiAppInformation(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":      cfg.Server.Name,
			"version":   cfg.Server.Version,
			"base_path": "/api/v1",
		})
	}
}

// apiRoot returns a handler for the root endpoint (used for docs and runtime).
// @Summary Welcome message
// @Description Get welcome message
// @Tags general
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router / [get]
func apiRoot(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to " + cfg.Server.Name + " API",
		})
	}
}

// apiHealth returns a handler for the health endpoint (used for docs and runtime).
// @Summary Health check
// @Description Check API health status
// @Tags general
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func apiHealth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "OK",
			"environment": cfg.Server.Environtment,
		})
	}
}
