package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/aftermath2/BTRY/config"
	"github.com/aftermath2/BTRY/http/api/middleware"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter(t *testing.T) {
	config := config.RateLimiter{
		Tokens:   3,
		Interval: 30,
	}
	rl, err := middleware.NewRateLimiter(config)
	assert.NoError(t, err)

	handler := rl.Handle(&noopHandler{})

	rec := httptest.NewRecorder()
	expectedLimit := strconv.FormatUint(config.Tokens, 10)

	for i := 0; i < 4; i++ {
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if i == int(config.Tokens) {
			continue
		}

		assert.Equal(t, expectedLimit, rec.Header().Get("X-Ratelimit-Limit"))
		assert.Equal(t, strconv.FormatUint(config.Tokens-uint64(i)-1, 10), rec.Header().Get("X-Ratelimit-Remaining"))
	}

	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
}
