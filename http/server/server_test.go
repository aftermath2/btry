package server_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/aftermath2/BTRY/config"
	"github.com/aftermath2/BTRY/http/server"
	"github.com/aftermath2/BTRY/logger"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	cfg := config.Server{
		Address: "127.0.0.1:21000",
		Logger:  config.Logger{Level: uint8(logger.DISABLED)},
	}

	response := "sats"
	mux := http.DefaultServeMux
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, response)
	})

	srv, err := server.New(cfg, routerMock{mux: mux})
	assert.NoError(t, err)

	ctx := context.Background()
	go srv.Run(ctx)

	time.Sleep(50 * time.Millisecond)
	resp, err := http.Get("http://" + cfg.Address + "/")
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	data, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, response+"\n", string(data))

	err = srv.Shutdown(ctx)
	assert.NoError(t, err)

	err = srv.Close()
	assert.NoError(t, err)
}

type routerMock struct {
	mux *http.ServeMux
}

func (rm routerMock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rm.mux.ServeHTTP(w, r)
}

func (rm routerMock) Close() error {
	return nil
}
