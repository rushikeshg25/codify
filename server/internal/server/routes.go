package server

import (
	"net/http"
	"os"
	"server/internal/controller"
	"server/internal/middleware"
	"server/internal/queue"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// corsOrigin returns the allowed frontend origin, configurable via
// CORS_ALLOW_ORIGIN and defaulting to the local dev frontend.
func corsOrigin() string {
	if origin := os.Getenv("CORS_ALLOW_ORIGIN"); origin != "" {
		return origin
	}
	return "http://localhost:3000"
}

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{corsOrigin()},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	db := s.dbInstance.GetDb()

	conn, ch, q := queue.InitQueue("codeground-queue")
	queueGlobal := queue.Queue{Conn: conn, Ch: ch, Q: q}

	authController := controller.NewAuthController(db)
	CodegroundController := controller.NewCodegroundController(db, &queueGlobal)

	r.GET("/health", s.healthHandler)
	api := r.Group("/api")
	{
		api.POST("/login", authController.Login)
		api.POST("/signup", authController.Signup)
		api.POST("/logout", authController.Logout)
		codeground := api.Group("/codeground")
		codeground.Use(middleware.AuthMiddleware)
		{
			codeground.GET("", CodegroundController.GetCodegrounds)
			codeground.POST("", CodegroundController.CreateCodeground)
			codeground.GET("/:codegroundId", CodegroundController.GetCodeground)
			codeground.GET("/:codegroundId/token", CodegroundController.GetCodegroundToken)
			codeground.GET("/:codegroundId/status", CodegroundController.GetCodegroundStatus)
			codeground.POST("/:codegroundId/heartbeat", CodegroundController.Heartbeat)
			codeground.PUT("/:codegroundId", CodegroundController.UpdateCodeground)
			codeground.DELETE("/:codegroundId", CodegroundController.DeleteCodeground)
			//prob a start codeground
			//prob a end codeground
			//prob get status of codeground
		}
	}

	return r
}

func (s *Server) healthHandler(c *gin.Context) {
	health := s.dbInstance.Health()
	if health["status"] == "down" {
		c.JSON(http.StatusServiceUnavailable, health)
		return
	}
	c.JSON(http.StatusOK, health)
}
