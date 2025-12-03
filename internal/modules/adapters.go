package modules

import (
	"math/big"

	"github.com/gabrielksneiva/ChainSystemPro/internal/adapters/evm/harness"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/ports"
	"go.uber.org/fx"
)

// AdaptersModule provides blockchain adapters
var AdaptersModule = fx.Module("adapters",
	fx.Provide(
		fx.Annotate(
			func() ports.ChainAdapter {
				h := harness.NewEVMHarness("ethereum")
				h.SetBalance("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb", big.NewInt(1000000000000000000))
				return h
			},
			fx.ResultTags(`name:"ethereum"`),
		),
		fx.Annotate(
			func() ports.ChainAdapter {
				h := harness.NewEVMHarness("polygon")
				h.SetBalance("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb", big.NewInt(5000000000000000000))
				return h
			},
			fx.ResultTags(`name:"polygon"`),
		),
		fx.Annotate(
			func() ports.ChainAdapter {
				h := harness.NewEVMHarness("tron")
				h.SetBalance("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb", big.NewInt(2000000000000000000))
				return h
			},
			fx.ResultTags(`name:"tron"`),
		),
	),
	fx.Invoke(registerAdapters),
)

type AdapterParams struct {
	fx.In

	Registry ports.ChainRegistry
	Ethereum ports.ChainAdapter `name:"ethereum"`
	Polygon  ports.ChainAdapter `name:"polygon"`
	Tron     ports.ChainAdapter `name:"tron"`
}

func registerAdapters(params AdapterParams) error {
	if err := params.Registry.Register("ethereum", params.Ethereum); err != nil {
		return err
	}
	if err := params.Registry.Register("polygon", params.Polygon); err != nil {
		return err
	}
	if err := params.Registry.Register("tron", params.Tron); err != nil {
		return err
	}
	return nil
}
