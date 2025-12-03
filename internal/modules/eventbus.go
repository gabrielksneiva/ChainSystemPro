package modules

import (
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/ports"
	"github.com/gabrielksneiva/ChainSystemPro/internal/infrastructure/eventbus"
	"github.com/gabrielksneiva/ChainSystemPro/internal/infrastructure/logger"
	"go.uber.org/fx"
)

// EventBusModule provides event bus dependency
var EventBusModule = fx.Module("eventbus",
	fx.Provide(
		func(log *logger.ZapLogger) ports.EventBus {
			return eventbus.NewInMemoryEventBus(log)
		},
		// Provide EventPublisher interface from EventBus
		func(bus ports.EventBus) ports.EventPublisher {
			return bus
		},
	),
	fx.Invoke(func(bus ports.EventBus, lifecycle fx.Lifecycle) {
		lifecycle.Append(fx.Hook{
			OnStart: bus.Start,
			OnStop:  bus.Stop,
		})
	}),
)
