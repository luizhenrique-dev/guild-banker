package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luizhenrique-dev/guild-banker/api/internal/auth"
)

type Authenticator interface {
	Authenticate(ctx context.Context, raw string) (auth.Principal, error)
}

func Auth(service Authenticator, logger *slog.Logger, timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path != "/api" && !strings.HasPrefix(path, "/api/") {
			c.Next()
			return
		}
		headers := c.Request.Header.Values("Authorization")
		if len(headers) == 0 {
			reject(c, logger, http.StatusUnauthorized, "missing_token")
			return
		}
		if len(headers) != 1 {
			reject(c, logger, http.StatusUnauthorized, "malformed_authorization")
			return
		}
		raw, ok := extractBearerToken(headers[0])
		if !ok {
			reject(c, logger, http.StatusUnauthorized, "malformed_authorization")
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		principal, err := service.Authenticate(ctx, raw)
		cancel()
		if err != nil {
			status, reason := http.StatusInternalServerError, "lookup_failed"
			if errors.Is(err, auth.ErrInvalidToken) {
				status, reason = http.StatusUnauthorized, "invalid_token"
			}
			if errors.Is(err, auth.ErrForbidden) {
				status, reason = http.StatusForbidden, "access_denied"
			}
			if rejection, ok := errors.AsType[*auth.RejectionError](err); ok {
				reason = rejection.Reason.String()
			}
			reject(c, logger, status, reason)
			return
		}
		c.Request = c.Request.WithContext(auth.WithPrincipal(c.Request.Context(), principal))
		c.Next()
	}
}

func reject(c *gin.Context, logger *slog.Logger, status int, reason string) {
	message := "internal server error"
	switch status {
	case http.StatusUnauthorized:
		message = "unauthorized"
		c.Header("WWW-Authenticate", "Bearer")
	case http.StatusForbidden:
		message = "forbidden"
	}
	logger.WarnContext(c.Request.Context(), "authentication rejected", "reason", reason, "status", status)
	c.AbortWithStatusJSON(status, gin.H{"error": message})
}

func extractBearerToken(header string) (string, bool) {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	raw := strings.TrimLeft(parts[1], " ")
	if raw == "" || strings.ContainsAny(raw, " \t\r\n,") {
		return "", false
	}
	return raw, true
}
