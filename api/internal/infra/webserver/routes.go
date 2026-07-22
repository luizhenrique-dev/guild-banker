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
		api.PUT("/guilds/:guildID", s.guild.UpdateName)
		api.PATCH("/guilds/:guildID/enable", s.guild.Enable)
		api.PATCH("/guilds/:guildID/disable", s.guild.Disable)
		api.GET("/guilds", s.guild.ListByMember)
		api.POST("/guilds/:guildID/invites", s.guild.InviteUser)
		api.DELETE("/guilds/:guildID/members/:userID", s.guild.RemoveUser)

		api.POST("/fixed-expenses", s.fixedExpense.Create)
		api.GET("/fixed-expenses", s.fixedExpense.ListActiveByUser)
		api.PATCH("/fixed-expenses/:id", s.fixedExpense.Update)
		api.PATCH("/fixed-expenses/:id/deactivate", s.fixedExpense.Deactivate)

		api.POST("/guilds/:guildID/transactions", s.transaction.Create)
		api.GET("/guilds/:guildID/transactions", s.transaction.List)
		api.PATCH("/guilds/:guildID/transactions/:id", s.transaction.Update)
		api.DELETE("/guilds/:guildID/transactions/:id", s.transaction.Delete)
		api.POST("/guilds/:guildID/transactions/bulk-categorize", s.transaction.BulkCategorize)
		api.PATCH("/guilds/:guildID/transactions/:id/visibility", s.transaction.SetVisibility)

		api.POST("/guilds/:guildID/imports", s.importer.Upload)
		api.GET("/guilds/:guildID/imports/:importID", s.importer.GetByID)
		api.PATCH("/guilds/:guildID/imports/:importID/items/:itemID", s.importer.UpdateItem)
		api.DELETE("/guilds/:guildID/imports/:importID/items/:itemID", s.importer.DeleteItem)
		api.POST("/guilds/:guildID/imports/:importID/confirm", s.importer.Confirm)
	}
	_ = api
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
