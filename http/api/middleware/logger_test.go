package middleware_test

import (
	"bufio"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/aftermath2/BTRY/config"
	"github.com/aftermath2/BTRY/http/api/middleware"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	file, err := os.CreateTemp("", "*")
	assert.NoError(t, err)
	t.Cleanup(func() { file.Close() })

	config := config.Logger{
		OutFile: file.Name(),
		Level:   2,
	}
	logger, err := middleware.NewLogger(config)
	assert.NoError(t, err)

	handler := logger.Log((&noopHandler{}))
	reader := bufio.NewReader(file)

	t.Run("Standard path", func(t *testing.T) {
		method := http.MethodGet
		path := "/test"
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(method, path, nil))

		log, err := reader.ReadString('\n')
		assert.NoError(t, err)

		expectedLogPortion := fmt.Sprintf("[%d] %s %s -", http.StatusOK, method, path)
		assert.True(t, strings.Contains(log, expectedLogPortion))
	})

	t.Run("Events path", func(t *testing.T) {
		method := http.MethodGet
		path := "/api/events"
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(method, path, nil))

		log, err := reader.ReadString('\n')
		assert.NoError(t, err)

		expectedLogEnd := fmt.Sprintf("%s %s\n", method, path)
		assert.True(t, strings.HasSuffix(log, expectedLogEnd))
	})
}
