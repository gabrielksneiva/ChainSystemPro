package modules

import (
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/ports"
	"github.com/gabrielksneiva/ChainSystemPro/internal/infrastructure/logger"
	"github.com/gabrielksneiva/ChainSystemPro/internal/infrastructure/registry"
	"go.uber.org/fx"
)

// RegistryModule provides chain registry dependency
var RegistryModule = fx.Module("registry",
	fx.Provide(
		func(log *logger.ZapLogger) ports.ChainRegistry {
			return registry.NewChainRegistry(log)
		},
	),
)
