package routes

import (
	"todoapp/internal/delivery/http/handler"
	"todoapp/internal/delivery/http/middleware"
	. "todoapp/pkg/auth"
	. "todoapp/pkg/config"
	. "todoapp/pkg/tracing"

	"github.com/gin-gonic/gin"
)

type HandlersConfig struct {
	AuthHandler *handler.AuthHandler
	TodoHandler *handler.TodoHandler
}

func SetupRouter(handlers HandlersConfig, metrics *AppMetrics, logger *LokiLogger) *gin.Engine {
	return SetupRouterWithConfig(handlers, metrics, logger, GetDefaultConfig())
}

func SetupRouterWithConfig(handlers HandlersConfig, metrics *AppMetrics, logger *LokiLogger, config *AppConfig) *gin.Engine {
	if gin.Mode() == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	middleware.SetupGinMiddlewareWithConfig(router, "todoapp", metrics, logger, config)

	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	if handlers.AuthHandler != nil {
		setupPublicRoutes(router, handlers.AuthHandler)
	}

	if handlers.TodoHandler != nil {
		setupProtectedRoutes(router, handlers.TodoHandler)
	}

	return router
}

func setupPublicRoutes(router *gin.Engine, authHandler *handler.AuthHandler) {
	public := router.Group("/")
	{
		public.POST("/signup", authHandler.RegisterByEmailAndPassword)
		public.POST("/auth", authHandler.AuthByEmailAndPassword)
	}
}

func setupProtectedRoutes(router *gin.Engine, todoHandler *handler.TodoHandler) {
	protected := router.Group("/")
	protected.Use(middleware.CurrentMiddleware())
	protected.Use(GinJwtMiddleware())
	{
		protected.GET("/todos", todoHandler.GetAllTodos)
		protected.POST("/todos", todoHandler.CreateTodo)
		protected.PUT("/todo/:uuid", todoHandler.UpdateTodo)
		protected.DELETE("/todos/:uuid", todoHandler.DeleteByUUID)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func SetupRouterForTests(handlers HandlersConfig) *gin.Engine {
	if gin.Mode() == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	if handlers.AuthHandler != nil {
		setupPublicRoutes(router, handlers.AuthHandler)
	}

	if handlers.TodoHandler != nil {
		setupProtectedRoutes(router, handlers.TodoHandler)
	}

	return router
}
