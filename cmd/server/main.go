package main

import (
	"github.com/gabrielksneiva/ChainSystemPro/internal/modules"

	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		modules.LoggerModule,
		modules.EventBusModule,
		modules.RegistryModule,
		modules.AdaptersModule,
		modules.UseCasesModule,
		modules.APIModule,
	)

	app.Run()
}
