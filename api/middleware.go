package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/pidrive/pidrive/internal/auth"
)

type contextKey string

const agentContextKey contextKey = "agent"

// AuthMiddleware returns a middleware that authenticates requests using Bearer tokens.
// It extracts the API key from the Authorization header and validates it against the
// auth service. On success, the authenticated agent is stored in the request context.
func AuthMiddleware(authService *auth.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				writeError(w, http.StatusUnauthorized, "missing Authorization header")
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				writeError(w, http.StatusUnauthorized, "invalid Authorization header format")
				return
			}

			apiKey := parts[1]
			agent, err := authService.Authenticate(apiKey)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "invalid API key")
				return
			}

			if !agent.Verified {
				writeError(w, http.StatusForbidden, "account not verified")
				return
			}

			ctx := context.WithValue(r.Context(), agentContextKey, agent)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetAgent retrieves the authenticated agent from the request context.
// Returns nil if no agent is present (e.g., for unauthenticated routes).
func GetAgent(r *http.Request) *auth.Agent {
	agent, _ := r.Context().Value(agentContextKey).(*auth.Agent)
	return agent
}
