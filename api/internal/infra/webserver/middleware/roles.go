package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luizhenrique-dev/guild-banker/api/internal/auth"
)

func RequireRole(role string, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := auth.ClaimsFromContext(c.Request.Context())
		if !ok {
			reject(c, logger, http.StatusUnauthorized, "missing_principal")
			return
		}
		if !claims.HasRole(role) {
			reject(c, logger, http.StatusForbidden, "missing_role")
			return
		}
		c.Next()
	}
}
