package http

import (
	"net/http"
	"study1/internal/core/config"
	"study1/internal/core/database"
	httpmw "study1/internal/core/http/middleware"

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

// NewServer creates a new HTTP server and registers provided modules. Accepts
// a database handle so middleware (e.g., activity logging) can write logs.
func NewServer(cfg *config.Config, db *database.DB, modules ...RouteRegistrar) *Server {
	router := gin.Default()

	// Middleware: include standard logger/recovery and activity DB logger
	router.Use(gin.Logger(), gin.Recovery(), httpmw.ActivityLogger(db))

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	api := router.Group(cfg.Server.BasePath)
	{
		api.GET("/", apiRoot(cfg))
		api.GET("/info", apiInfo(cfg))
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

// @Summary API Information
// @Description Get API Information
// @Tags general
// @Produce json
// @Success 200 {object} map[string]string
// @Router /info [get]
func apiInfo(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"env":       cfg.Server.Environtment,
			"name":      cfg.Server.Name,
			"version":   cfg.Server.Version,
			"protocol":  cfg.Server.Protocol,
			"host":      cfg.Server.Host,
			"base_path": cfg.Server.BasePath,
			"port":      cfg.Server.Port,
			"url":       cfg.Server.URL,
		})
	}
}

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
