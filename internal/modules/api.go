package modules

import (
	"context"
	"os"

	"github.com/gabrielksneiva/ChainSystemPro/internal/api"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/ports"
	"github.com/gabrielksneiva/ChainSystemPro/internal/infrastructure/logger"
	"github.com/gabrielksneiva/ChainSystemPro/internal/usecases"
	"go.uber.org/fx"
)

// APIModule provides API server
var APIModule = fx.Module("api",
	fx.Provide(
		func(
			registry ports.ChainRegistry,
			getBalanceUC *usecases.GetBalanceUseCase,
			createTransactionUC *usecases.CreateTransactionUseCase,
			signTransactionUC *usecases.SignTransactionUseCase,
			broadcastTransactionUC *usecases.BroadcastTransactionUseCase,
			estimateFeeUC *usecases.EstimateFeeUseCase,
			getTransactionStatusUC *usecases.GetTransactionStatusUseCase,
			log *logger.ZapLogger,
		) *api.Server {
			return api.NewServer(
				registry,
				getBalanceUC,
				createTransactionUC,
				signTransactionUC,
				broadcastTransactionUC,
				estimateFeeUC,
				getTransactionStatusUC,
				log,
			)
		},
	),
	fx.Invoke(func(server *api.Server, lifecycle fx.Lifecycle, log *logger.ZapLogger) {
		lifecycle.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				port := os.Getenv("PORT")
				if port == "" {
					port = "8080"
				}

				go func() {
					if err := server.Start(port); err != nil {
						log.Error("server error", err, nil)
					}
				}()

				log.Info("server started", map[string]interface{}{
					"port": port,
				})

				return nil
			},
			OnStop: func(ctx context.Context) error {
				log.Info("shutting down server...", nil)
				return server.Shutdown()
			},
		})
	}),
)
