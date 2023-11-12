package middleware

import (
	"net/http"
	"time"

	"github.com/aftermath2/BTRY/config"
	"github.com/aftermath2/BTRY/logger"
)

// Logger contains a logger to log HTTP calls.
type Logger struct {
	logger *logger.Logger
}

// NewLogger returns a new logging middleware.
func NewLogger(config config.Logger) (*Logger, error) {
	logger, err := logger.New(config)
	if err != nil {
		return nil, err
	}

	return &Logger{logger}, nil
}

// Log intercepts and logs each API request.
func (l *Logger) Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/events" {
			l.logger.Info(r.Method, " ", r.URL.Path)
			next.ServeHTTP(w, r)
			return
		}

		lrw := newLoggingResponseWriter(w)
		t := time.Now().UTC()
		next.ServeHTTP(lrw, r)
		latency := time.Since(t)

		l.logger.Infof("[%d] %s %s - %dms",
			lrw.statusCode, r.Method, r.URL.Path, latency.Milliseconds())
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{ResponseWriter: w}
}

// WriteHeader captures the status code to be able to log it after the response has been sent.
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
