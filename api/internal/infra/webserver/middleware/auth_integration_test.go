//go:build integration

package middleware

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
	"github.com/luizhenrique-dev/guild-banker/api/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type authenticatorFunc func(context.Context, string) (auth.Principal, error)

func (f authenticatorFunc) Authenticate(ctx context.Context, raw string) (auth.Principal, error) {
	return f(ctx, raw)
}

func TestAuth(t *testing.T) {
	for _, tt := range []struct {
		name    string
		headers []string
		err     error
		status  int
		reason  string
	}{
		{name: "absent", status: 401, reason: "missing_token"},
		{name: "malformed", headers: []string{"Basic secret"}, status: 401, reason: "malformed_authorization"},
		{name: "duplicated", headers: []string{"Bearer secret-token", "Bearer secret-token"}, status: 401, reason: "malformed_authorization"},
		{name: "valid", headers: []string{"Bearer secret-token"}, status: 200},
		{name: "invalid", headers: []string{"Bearer secret-token"}, err: &auth.RejectionError{Reason: auth.ReasonInvalidToken}, status: 401, reason: "invalid_token"},
		{name: "expired", headers: []string{"Bearer secret-token"}, err: &auth.RejectionError{Reason: auth.ReasonTokenExpired}, status: 401, reason: "token_expired"},
		{name: "forbidden", headers: []string{"Bearer secret-token"}, err: &auth.RejectionError{Reason: auth.ReasonUserDisabled}, status: 403, reason: "user_disabled"},
		{name: "technical", headers: []string{"Bearer secret-token"}, err: errors.New("raw failure secret-email@example.test secret-sub"), status: 500, reason: "lookup_failed"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			var calls, handled int
			principal := auth.Principal{Claims: auth.Claims{UserID: "secret-sub", PreferredUsername: "secret-user"}, Actor: audit.Actor{UserID: 42, Email: "secret-email@example.test"}}
			service := authenticatorFunc(func(ctx context.Context, raw string) (auth.Principal, error) {
				calls++
				assert.Equal(t, "secret-token", raw)
				deadline, ok := ctx.Deadline()
				assert.True(t, ok)
				assert.LessOrEqual(t, time.Until(deadline), time.Second)
				return principal, tt.err
			})
			router := gin.New()
			requestCtx, cancelRequest := context.WithTimeout(context.Background(), time.Minute)
			defer cancelRequest()
			requestDeadline, _ := requestCtx.Deadline()
			router.Use(Auth(service, slog.New(slog.NewJSONHandler(&logs, nil)), time.Second))
			router.GET("/api/probe", func(c *gin.Context) {
				handled++
				assert.NoError(t, c.Request.Context().Err())
				deadline, ok := c.Request.Context().Deadline()
				assert.True(t, ok)
				assert.Equal(t, requestDeadline, deadline)
				got, ok := auth.PrincipalFromContext(c.Request.Context())
				assert.True(t, ok)
				assert.Equal(t, principal, got)
				c.Status(200)
			})
			req := httptest.NewRequest(http.MethodGet, "/api/probe?token=secret-token", nil)
			req = req.WithContext(requestCtx)
			req.AddCookie(&http.Cookie{Name: "access_token", Value: "secret-token"})
			for _, header := range tt.headers {
				req.Header.Add("Authorization", header)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.status, w.Code)
			if tt.status == 200 {
				assert.Equal(t, 1, handled)
				assert.Empty(t, logs.String())
				return
			}
			assert.Zero(t, handled)
			if tt.name == "absent" || tt.name == "malformed" || tt.name == "duplicated" {
				assert.Zero(t, calls)
			}
			body := map[int]string{401: `{"error":"unauthorized"}`, 403: `{"error":"forbidden"}`, 500: `{"error":"internal server error"}`}
			assert.JSONEq(t, body[tt.status], w.Body.String())
			if tt.status == 401 {
				assert.Equal(t, "Bearer", w.Header().Get("WWW-Authenticate"))
			}
			assert.Contains(t, logs.String(), `"reason":"`+tt.reason+`"`)
			assert.Contains(t, logs.String(), `"status":`)
			assert.Equal(t, 1, bytes.Count(logs.Bytes(), []byte("\n")))
			for _, sensitive := range []string{"secret-token", "secret-sub", "secret-user", "secret-email@example.test", "raw failure"} {
				assert.NotContains(t, logs.String(), sensitive)
				assert.NotContains(t, w.Body.String(), sensitive)
			}
		})
	}
}

func TestAuthTimeoutAndCancellation(t *testing.T) {
	for _, cancelled := range []bool{false, true} {
		t.Run(map[bool]string{false: "timeout", true: "cancelled"}[cancelled], func(t *testing.T) {
			service := authenticatorFunc(func(ctx context.Context, _ string) (auth.Principal, error) {
				<-ctx.Done()
				if cancelled {
					assert.ErrorIs(t, ctx.Err(), context.Canceled)
				} else {
					assert.ErrorIs(t, ctx.Err(), context.DeadlineExceeded)
				}
				return auth.Principal{}, auth.ErrInvalidToken
			})
			router := gin.New()
			var logs bytes.Buffer
			router.Use(Auth(service, slog.New(slog.NewJSONHandler(&logs, nil)), 20*time.Millisecond))
			router.GET("/api/probe", func(c *gin.Context) { t.Error("rejected request reached handler") })
			req := httptest.NewRequest("GET", "/api/probe", nil)
			ctx, cancel := context.WithCancel(req.Context())
			defer cancel()
			if cancelled {
				cancel()
			}
			req = req.WithContext(ctx)
			req.Header.Set("Authorization", "Bearer token")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			require.Equal(t, 401, w.Code)
		})
	}
}
