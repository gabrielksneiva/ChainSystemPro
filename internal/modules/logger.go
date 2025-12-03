package modules

import (
	"github.com/gabrielksneiva/ChainSystemPro/internal/infrastructure/logger"
	"go.uber.org/fx"
)

// LoggerModule provides logger dependency
var LoggerModule = fx.Module("logger",
	fx.Provide(
		func() (*logger.ZapLogger, error) {
			return logger.NewDevelopmentLogger()
		},
	),
)
