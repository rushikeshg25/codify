package server

import (
	"net/http"
	"server/internal/controller"
	"server/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Add your frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true, // Enable cookies/auth
	}))
	db := s.dbInstance.GetDb()

	authController := controller.NewAuthController(db)
	CodegroundController := controller.NewCodegroundController(db)

	r.GET("/health", s.healthHandler)
	api := r.Group("/api")
	{
		api.POST("/login", authController.Login)
		api.POST("/signup", authController.Signup)
		api.POST("/logout", authController.Logout)
		playground := api.Group("/playground")
		playground.Use(middleware.AuthMiddleware)
		{
			playground.GET("/", CodegroundController.GetPlaygrounds)
			playground.POST("/", CodegroundController.CreatePlayground)
			playground.GET("/:playGroundId", CodegroundController.GetPlayground)
			playground.PUT("/:playGroundId", CodegroundController.UpdatePlayground)
			playground.DELETE("/:playGroundId", CodegroundController.DeletePlayground)
			//prob a start codeground 
			//prob a end codeground
			//prob get status of codeground
		}
	}

	return r
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.dbInstance.Health())
}
