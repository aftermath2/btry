package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aftermath2/BTRY/http/api/middleware"

	"github.com/stretchr/testify/assert"
)

func TestCors(t *testing.T) {
	corsHandler := middleware.Cors(&noopHandler{})

	rec := httptest.NewRecorder()
	corsHandler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	headers := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Headers": "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, accept, origin, Cache-Control, X-Requested-With",
		"Access-Control-Allow-Methods": "GET, POST, HEAD",
		"Access-Control-Max-Age":       "600",
		"Vary":                         "Origin, Access-Control-Request-Method, Access-Control-Request-Headers",
		"X-Xss-Protection":             "1; mode=block",
		"Strict-Transport-Security":    "max-age=63072000; includeSubDomains; preload",
		"X-Frame-Options":              "DENY",
		"X-Content-Type-Options":       "nosniff",
		// "Content-Security-Policy":           "default-src 'self'",
		"X-Permitted-Cross-Domain-Policies": "none",
		"Referrer-Policy":                   "no-referrer",
		"Feature-Policy":                    "microphone 'none'; camera 'none'",
	}
	for k, v := range headers {
		assert.Equal(t, v, rec.Header().Get(k))
	}
}

func TestCorsOptionsMethod(t *testing.T) {
	corsHandler := middleware.Cors(&noopHandler{})

	rec := httptest.NewRecorder()
	corsHandler.ServeHTTP(rec, httptest.NewRequest(http.MethodOptions, "/", nil))

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

type noopHandler struct{}

func (h *noopHandler) ServeHTTP(_ http.ResponseWriter, _ *http.Request) {}
