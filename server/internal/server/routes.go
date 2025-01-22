package server

import (
	"net/http"
	"server/internal/controller"

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
	db:=s.dbInstance.GetDb()

	auth:=controller.NewAuthController(db)
	controller.NewPlaygroundController(db)

	r.GET("/health", s.healthHandler)
	api:= r.Group("/api")
	{
		api.POST("/login",auth.Login)
		api.POST("/signup",auth.Signup)
		api.GET("/playground", )
	}

	return r
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.dbInstance.Health())
}
