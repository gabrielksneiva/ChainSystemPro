package api

import (
	"bytes"
	"context"
	"encoding/json"
	"math/big"
	"net/http/httptest"
	"testing"

	"github.com/gabrielksneiva/ChainSystemPro/internal/adapters/evm/harness"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/entities"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/valueobjects"
	"github.com/gabrielksneiva/ChainSystemPro/internal/infrastructure/registry"
	"github.com/gabrielksneiva/ChainSystemPro/internal/mocks"
	"github.com/gabrielksneiva/ChainSystemPro/internal/usecases"
	"github.com/stretchr/testify/require"
)

func TestServerRoutes(t *testing.T) {
	t.Parallel()

	logger := mocks.NewMockLogger()
	reg := registry.NewChainRegistry(logger)
	h := harness.NewEVMHarness("evm-mainnet")
	_ = reg.Register("evm-mainnet", h)

	// minimal UCs with mocks
	eb := mocks.NewMockEventPublisher()
	gb := usecases.NewGetBalanceUseCase(reg, eb, logger)
	ct := usecases.NewCreateTransactionUseCase(reg, eb, logger)
	st := usecases.NewSignTransactionUseCase(reg, eb, logger)
	bt := usecases.NewBroadcastTransactionUseCase(reg, eb, logger)
	ef := usecases.NewEstimateFeeUseCase(reg, eb, logger)
	gs := usecases.NewGetTransactionStatusUseCase(reg, logger)

	srv := NewServer(reg, gb, ct, st, bt, ef, gs, logger)

	// list chains
	req := httptest.NewRequest("GET", "/v1/chains", nil)
	resp, err := srv.app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	// balance endpoint
	req = httptest.NewRequest("GET", "/v1/evm-mainnet/balance/0xabc", nil)
	resp, err = srv.app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	// create transaction
	body := map[string]interface{}{"from": "0xabc", "to": "0xdef", "value": "1000", "gas_limit": 21000}
	reqBody, _ := json.Marshal(body)
	req = httptest.NewRequest("POST", "/v1/evm-mainnet/transaction/create", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = srv.app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, 201, resp.StatusCode)

	// broadcast transaction
	broadcastBody := map[string]interface{}{"transaction_id": "tx123", "signed_data": "0xsigned"}
	reqBody, _ = json.Marshal(broadcastBody)
	req = httptest.NewRequest("POST", "/v1/evm-mainnet/transaction/send", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = srv.app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	// get transaction status
	h.SetBalance("0xabc", big.NewInt(100))
	from, _ := valueobjects.NewAddress("0xabc", "evm-mainnet")
	to, _ := valueobjects.NewAddress("0xdef", "evm-mainnet")
	testTx, _ := h.BuildTransaction(context.Background(), entities.TransactionParams{ChainID: "evm-mainnet", From: from, To: to, Value: big.NewInt(1)})
	_ = h.SignTransaction(context.Background(), testTx, []byte("key"))
	txHash, _ := h.BroadcastTransaction(context.Background(), testTx)
	req = httptest.NewRequest("GET", "/v1/evm-mainnet/transaction/"+txHash.HexWithoutPrefix(), nil)
	resp, err = srv.app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	// error cases
	req = httptest.NewRequest("GET", "/v1/unknown/balance/0xabc", nil)
	resp, err = srv.app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, 500, resp.StatusCode)

	req = httptest.NewRequest("POST", "/v1/evm-mainnet/transaction/create", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	resp, err = srv.app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)
}

func TestServerStartShutdown(t *testing.T) {
	t.Parallel()
	h := harness.NewEVMHarness("evm-mainnet")
	registry := mocks.NewMockChainRegistry()
	_ = registry.Register("evm-mainnet", h)
	publisher := mocks.NewMockEventPublisher()
	logger := mocks.NewMockLogger()

	getBalanceUC := usecases.NewGetBalanceUseCase(registry, publisher, logger)
	createTxUC := usecases.NewCreateTransactionUseCase(registry, publisher, logger)
	signTxUC := usecases.NewSignTransactionUseCase(registry, publisher, logger)
	broadcastTxUC := usecases.NewBroadcastTransactionUseCase(registry, publisher, logger)
	estimateFeeUC := usecases.NewEstimateFeeUseCase(registry, publisher, logger)
	getStatusUC := usecases.NewGetTransactionStatusUseCase(registry, logger)

	srv := NewServer(registry, getBalanceUC, createTxUC, signTxUC, broadcastTxUC, estimateFeeUC, getStatusUC, logger)

	go func() {
		_ = srv.Start("9999")
	}()
	require.NoError(t, srv.Shutdown())
}
