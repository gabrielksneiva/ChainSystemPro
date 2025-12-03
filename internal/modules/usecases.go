package modules

import (
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/ports"
	"github.com/gabrielksneiva/ChainSystemPro/internal/infrastructure/logger"
	"github.com/gabrielksneiva/ChainSystemPro/internal/usecases"
	"go.uber.org/fx"
)

// UseCasesModule provides all use cases
var UseCasesModule = fx.Module("usecases",
	fx.Provide(
		func(registry ports.ChainRegistry, eventBus ports.EventPublisher, log *logger.ZapLogger) *usecases.GetBalanceUseCase {
			return usecases.NewGetBalanceUseCase(registry, eventBus, log)
		},
		func(registry ports.ChainRegistry, eventBus ports.EventPublisher, log *logger.ZapLogger) *usecases.CreateTransactionUseCase {
			return usecases.NewCreateTransactionUseCase(registry, eventBus, log)
		},
		func(registry ports.ChainRegistry, eventBus ports.EventPublisher, log *logger.ZapLogger) *usecases.SignTransactionUseCase {
			return usecases.NewSignTransactionUseCase(registry, eventBus, log)
		},
		func(registry ports.ChainRegistry, eventBus ports.EventPublisher, log *logger.ZapLogger) *usecases.BroadcastTransactionUseCase {
			return usecases.NewBroadcastTransactionUseCase(registry, eventBus, log)
		},
		func(registry ports.ChainRegistry, eventBus ports.EventPublisher, log *logger.ZapLogger) *usecases.EstimateFeeUseCase {
			return usecases.NewEstimateFeeUseCase(registry, eventBus, log)
		},
		func(registry ports.ChainRegistry, log *logger.ZapLogger) *usecases.GetTransactionStatusUseCase {
			return usecases.NewGetTransactionStatusUseCase(registry, log)
		},
	),
)
