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
		api.POST("/guilds", s.guild.Create)
		api.PUT("/guilds/:id", s.guild.UpdateName)
		api.PATCH("/guilds/:id/enable", s.guild.Enable)
		api.PATCH("/guilds/:id/disable", s.guild.Disable)
		api.GET("/guilds", s.guild.ListByMember)
		api.POST("/guilds/:id/invites", s.guild.InviteUser)
		api.DELETE("/guilds/:id/members/:userID", s.guild.RemoveUser)

		api.POST("/fixed-expenses", s.fixedExpense.Create)
		api.GET("/fixed-expenses", s.fixedExpense.ListActiveByUser)
		api.PATCH("/fixed-expenses/:id", s.fixedExpense.Update)
		api.PATCH("/fixed-expenses/:id/deactivate", s.fixedExpense.Deactivate)
	}
	_ = api
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
