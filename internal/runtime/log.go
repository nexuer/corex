package runtime

import (
	"log/slog"

	kratoslog "github.com/go-kratos/kratos/v3/log"
	"github.com/nexuer/log"
)

// Logger adapts the Nex logger at the runtime construction boundary.
func Logger(logger *log.Logger) *slog.Logger {
	return slog.New(log.NewSlogHandler(logger))
}

// SetDefaultLogger routes internal runtime logs through the Nex logger.
func SetDefaultLogger(logger *log.Logger) {
	kratoslog.SetDefault(Logger(logger))
}
