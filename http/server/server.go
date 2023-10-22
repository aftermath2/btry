// Package server ..
package server

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aftermath2/BTRY/config"
	"github.com/aftermath2/BTRY/http/api"
	"github.com/aftermath2/BTRY/logger"

	"github.com/pkg/errors"
)

// Server contains the server configuration.
type Server struct {
	*http.Server
	router          api.Router
	logger          *logger.Logger
	errLogFile      *os.File
	shutdownTimeout time.Duration
}

// New create and returns a server.
func New(cfg config.Server, router api.Router) (*Server, error) {
	logger, err := logger.New(cfg.Logger)
	if err != nil {
		return nil, err
	}

	writers := []io.Writer{os.Stderr}
	var errLogFile *os.File
	if cfg.Logger.OutFile != "" {
		f, err := os.OpenFile(cfg.Logger.OutFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
		if err != nil {
			return nil, errors.Wrap(err, "opening log file")
		}

		errLogFile = f
		writers = append(writers, f)
	}

	// https://blog.gopheracademy.com/advent-2016/exposing-go-on-the-internet/
	return &Server{
		Server: &http.Server{
			Addr:         cfg.Address,
			Handler:      router,
			ReadTimeout:  cfg.Timeout.Read,
			WriteTimeout: cfg.Timeout.Write,
			IdleTimeout:  cfg.Timeout.Idle,
			TLSConfig: &tls.Config{
				MinVersion:   tls.VersionTLS12,
				Certificates: cfg.TLSCertificates,
			},
			ErrorLog: log.New(io.MultiWriter(writers...), cfg.Logger.Label, log.LstdFlags),
		},
		router:          router,
		logger:          logger,
		shutdownTimeout: cfg.Timeout.Shutdown,
		errLogFile:      errLogFile,
	}, nil
}

// Run starts the server.
func (srv *Server) Run(ctx context.Context) error {
	serverErr := make(chan error, 1)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	go srv.listenAndServe(serverErr)

	select {
	case err := <-serverErr:
		return errors.Wrap(err, "Listen and serve failed")

	case <-interrupt:
		srv.logger.Info("Start shutdown")

		// Give outstanding requests a deadline for completion
		ctx, cancel := context.WithTimeout(ctx, srv.shutdownTimeout)
		defer cancel()

		if err := srv.errLogFile.Close(); err != nil {
			return errors.Wrap(err, "Closing server log file")
		}

		if err := srv.router.Close(); err != nil {
			return errors.Wrap(err, "Closing router")
		}

		if err := srv.Shutdown(ctx); err != nil {
			return errors.Wrapf(err, "Graceful shutdown did not complete in %v", srv.shutdownTimeout)
		}

		if err := srv.Close(); err != nil {
			return errors.Wrap(err, "Couldn't stop server gracefully")
		}

		srv.logger.Info("Server shutdown gracefully")
		return nil
	}
}

func (srv *Server) listenAndServe(serverErr chan error) {
	l, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		serverErr <- err
		return
	}

	srv.logger.Infof("Listening on %s", srv.Addr)
	serverErr <- srv.Serve(l)
}
