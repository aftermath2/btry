// Package logger contains utilities for logging relevant information.
package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/aftermath2/BTRY/config"

	"github.com/pkg/errors"
)

const (
	// DISABLED shows no logs.
	DISABLED Level = iota
	// DEBUG designates fine-grained informational events useful to debug an application.
	DEBUG
	// INFO designates informational messages that highlight the progress of the application at coarse-grained level.
	INFO
	// WARNING displays an alert signaling caution.
	WARNING
	// ERROR designates error events.
	ERROR
	// FATAL shows an error and exits.
	FATAL
)

// Level represents the logging Level used.
type Level uint8

// Logger contains the logging options.
type Logger struct {
	out   io.Writer
	file  *os.File
	label string
	level Level
}

// New creates a new logger.
func New(config config.Logger) (*Logger, error) {
	writers := []io.Writer{os.Stderr}
	var file *os.File

	if config.OutFile != "" {
		// Create path to the log file
		if err := os.MkdirAll(filepath.Dir(config.OutFile), 0o700); err != nil {
			return nil, errors.Wrap(err, "creating log file path")
		}

		f, err := os.OpenFile(config.OutFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
		if err != nil {
			return nil, errors.Wrap(err, "opening log file")
		}

		writers = append(writers, f)
		file = f
	}

	return &Logger{
		level: Level(config.Level),
		label: config.Label,
		out:   io.MultiWriter(writers...),
		file:  file,
	}, nil
}

func (l Logger) log(level Level, message string) {
	if l.level == DISABLED || level < l.level {
		return
	}

	var source string
	if l.level == DEBUG {
		_, file, line, _ := runtime.Caller(2)
		split := strings.Split(file, "/")
		join := strings.Join(split[4:], "/")
		source = fmt.Sprintf(" (%s:%d)", join, line)
	}

	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05.000")
	// 2006-01-02 15:04:05.000 [DBG] API (source:1): message
	log := fmt.Sprintf("%s [%s] %s%s: %s", timestamp, levelName(level), l.label, source, message)
	fmt.Fprintln(l.out, log)

	if l.level == FATAL {
		if l.file != nil {
			l.file.Close()
		}
		os.Exit(1)
	}
}

// Debug provides useful information for debugging.
func (l Logger) Debug(args ...interface{}) {
	l.log(DEBUG, fmt.Sprint(args...))
}

// Debugf is like Debug but takes a formatted message.
func (l Logger) Debugf(format string, args ...interface{}) {
	l.log(DEBUG, fmt.Sprintf(format, args...))
}

// Error reports the application errors.
func (l Logger) Error(args ...interface{}) {
	l.log(ERROR, fmt.Sprint(args...))
}

// Errorf is like Error but takes a formatted message.
func (l Logger) Errorf(format string, args ...interface{}) {
	l.log(ERROR, fmt.Sprintf(format, args...))
}

// Fatal reports the application errors and exits.
func (l Logger) Fatal(args ...interface{}) {
	l.log(FATAL, fmt.Sprint(args...))
}

// Fatalf is like Fatal but takes a formatted message.
func (l Logger) Fatalf(format string, args ...interface{}) {
	l.log(FATAL, fmt.Sprintf(format, args...))
}

// Info provides useful information about the server.
func (l Logger) Info(args ...interface{}) {
	l.log(INFO, fmt.Sprint(args...))
}

// Infof is like Info but takes a formatted message.
func (l Logger) Infof(format string, args ...interface{}) {
	l.log(INFO, fmt.Sprintf(format, args...))
}

// Warning reports the application alerts.
func (l Logger) Warning(args ...interface{}) {
	l.log(WARNING, fmt.Sprint(args...))
}

// Warningf is like Warning but takes a formatted message.
func (l Logger) Warningf(format string, args ...interface{}) {
	l.log(WARNING, fmt.Sprintf(format, args...))
}

func levelName(level Level) string {
	switch level {
	case DEBUG:
		return "DBG"
	case INFO:
		return "INF"
	case WARNING:
		return "WRN"
	case ERROR:
		return "ERR"
	case FATAL:
		return "FTL"
	default:
		return ""
	}
}
