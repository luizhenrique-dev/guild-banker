//go:build integration

package middleware

import (
	"bytes"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/luizhenrique-dev/guild-banker/api/internal/auth"
	"github.com/stretchr/testify/assert"
)

func TestRequireRole(t *testing.T) {
	for _, tt := range []struct {
		name         string
		hasPrincipal bool
		roles        []string
		status       int
	}{
		{name: "no principal", status: 401},
		{name: "missing role", hasPrincipal: true, roles: []string{"member"}, status: 403},
		{name: "admin", hasPrincipal: true, roles: []string{"admin"}, status: 200},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			router := gin.New()
			router.GET("/api/test-role", RequireRole("admin", slog.New(slog.NewJSONHandler(&logs, nil))), func(c *gin.Context) { c.Status(200) })
			req := httptest.NewRequest("GET", "/api/test-role", nil)
			if tt.hasPrincipal {
				req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{Claims: auth.Claims{RealmRoles: tt.roles}}))
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.status, w.Code)
			if tt.status == 401 {
				assert.JSONEq(t, `{"error":"unauthorized"}`, w.Body.String())
			}
			if tt.status == 403 {
				assert.JSONEq(t, `{"error":"forbidden"}`, w.Body.String())
			}
		})
	}
}
