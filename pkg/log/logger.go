// Package log holds the global logger instance.
package log

import (
	"sync"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

var (
	// Synchronizes access to the global logger variable.
	loggerMu sync.Mutex
	global   logr.Logger
)

func init() {
	zapLog := zap.NewNop()
	SetLogger(zapr.NewLogger(zapLog))
}

// Logger returns the global logger instance.
func Logger() logr.Logger {
	loggerMu.Lock()
	logger := global
	loggerMu.Unlock()
	return logger
}

// SetLogger sets the global logger instance.
func SetLogger(logger logr.Logger) {
	loggerMu.Lock()
	global = logger
	loggerMu.Unlock()
}
