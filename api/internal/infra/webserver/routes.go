package webserver

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/luizhenrique-dev/guild-banker/api/internal/infra/webserver/middleware"
)

func (s *Server) registerRoutes() {
	s.router.GET("/health", s.healthHandler)

	api := s.router.Group("/api/v1", middleware.Auth(s.cfg))
	{
		// api.GET("/users", listUsers)
	}
	_ = api
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
