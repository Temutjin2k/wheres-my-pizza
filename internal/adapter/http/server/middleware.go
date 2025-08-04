package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/types"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/logger"
)

// RequestLoggingMiddleware injects a request ID into the context and logs the request details.
func (a *API) RequestLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 1. Get or generate request ID
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = newRequestID()
		}

		// 2. Set X-Request-ID in response for tracing/debugging
		w.Header().Set("X-Request-ID", reqID)

		// 3. Inject request ID into context
		ctx := logger.WithRequestID(r.Context(), reqID)

		// 4. Log request start
		a.log.Debug(
			ctx,
			types.ActionRequestReceived,
			"Started",
			"method", r.Method,
			"URL", r.URL.Path,
		)

		// 5. Serve the request
		next.ServeHTTP(w, r.WithContext(ctx))

		// 6. Log request end
		duration := time.Since(start)
		a.log.Debug(
			ctx,
			types.ActionRequestReceived,
			"Completed",
			"method", r.Method,
			"URL", r.URL.Path,
			"duration", duration,
		)
	})
}

// newRequestID returns a 16-byte random hex string, e.g. “9f86d081884c7d65…”
func newRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// fallback to timestamp if crypto/rand fails
		return hex.EncodeToString(fmt.Appendf(nil, "%d", time.Now().UnixNano()))
	}
	return hex.EncodeToString(b)
}
