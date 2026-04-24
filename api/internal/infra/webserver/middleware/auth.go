package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/luizhenrique-dev/guild-banker/api/config"
)

func Auth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := extractBearerToken(c.GetHeader("Authorization"))
		if !ok || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		_ = cfg.Keycloak.JWKSURI

		// TODO: implement Keycloak JWT validation.
		c.AbortWithStatusJSON(
			http.StatusNotImplemented,
			gin.H{"error": "keycloak jwt validation is not implemented"},
		)
	}
}

func extractBearerToken(authHeader string) (string, bool) {
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return "", false
	}

	if !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	return strings.TrimSpace(parts[1]), true
}
