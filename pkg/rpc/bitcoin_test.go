package rpc

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetBalance(t *testing.T) {
	tests := []struct {
		name           string
		address        string
		serverResponse RPCResponse
		expectedAmount float64
		expectErr      bool
	}{
		{
			name:    "address with balance",
			address: "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			serverResponse: RPCResponse{
				Result: map[string]interface{}{
					"balance": float64(50.0),
				},
				Error: nil,
				ID:    "1",
			},
			expectedAmount: 50.0,
			expectErr:      false,
		},
		{
			name:    "address with zero balance",
			address: "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2",
			serverResponse: RPCResponse{
				Result: map[string]interface{}{
					"balance": float64(0),
				},
				Error: nil,
				ID:    "1",
			},
			expectedAmount: 0,
			expectErr:      false,
		},
		{
			name:    "invalid address",
			address: "invalid",
			serverResponse: RPCResponse{
				Result: nil,
				Error: &RPCError{
					Code:    -5,
					Message: "Invalid Bitcoin address",
				},
				ID: "1",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer server.Close()

			client, err := NewClient(server.URL, "user", "pass")
			require.NoError(t, err)

			amount, err := client.GetBalance(tt.address)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedAmount, amount)
			}
		})
	}
}

func TestClient_ListUnspent(t *testing.T) {
	tests := []struct {
		name           string
		addresses      []string
		serverResponse RPCResponse
		expectedUTXOs  int
		expectErr      bool
	}{
		{
			name:      "address with UTXOs",
			addresses: []string{"1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"},
			serverResponse: RPCResponse{
				Result: []interface{}{
					map[string]interface{}{
						"txid":          "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
						"vout":          float64(0),
						"address":       "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
						"scriptPubKey":  "76a91462e907b15cbf27d5425399ebf6f0fb50ebb88f1888ac",
						"amount":        float64(50.0),
						"confirmations": float64(100),
					},
					map[string]interface{}{
						"txid":          "0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098",
						"vout":          float64(1),
						"address":       "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
						"scriptPubKey":  "76a91462e907b15cbf27d5425399ebf6f0fb50ebb88f1888ac",
						"amount":        float64(25.5),
						"confirmations": float64(50),
					},
				},
				Error: nil,
				ID:    "1",
			},
			expectedUTXOs: 2,
			expectErr:     false,
		},
		{
			name:      "address with no UTXOs",
			addresses: []string{"1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2"},
			serverResponse: RPCResponse{
				Result: []interface{}{},
				Error:  nil,
				ID:     "1",
			},
			expectedUTXOs: 0,
			expectErr:     false,
		},
		{
			name:      "multiple addresses",
			addresses: []string{"1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2"},
			serverResponse: RPCResponse{
				Result: []interface{}{
					map[string]interface{}{
						"txid":          "abc123",
						"vout":          float64(0),
						"address":       "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
						"amount":        float64(10.0),
						"confirmations": float64(6),
					},
				},
				Error: nil,
				ID:    "1",
			},
			expectedUTXOs: 1,
			expectErr:     false,
		},
		{
			name:      "rpc error",
			addresses: []string{"invalid"},
			serverResponse: RPCResponse{
				Result: nil,
				Error: &RPCError{
					Code:    -5,
					Message: "Invalid address",
				},
				ID: "1",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer server.Close()

			client, err := NewClient(server.URL, "user", "pass")
			require.NoError(t, err)

			utxos, err := client.ListUnspent(tt.addresses...)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUTXOs, len(utxos))
				if len(utxos) > 0 {
					// Verify first UTXO structure
					assert.NotEmpty(t, utxos[0].TxID)
					assert.GreaterOrEqual(t, utxos[0].Vout, uint32(0))
					assert.NotEmpty(t, utxos[0].Address)
					assert.Greater(t, utxos[0].Amount, float64(0))
				}
			}
		})
	}
}

func TestClient_ListUnspent_MinConfirmations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse request to verify minconf parameter
		var req RPCRequest
		json.NewDecoder(r.Body).Decode(&req)

		// listunspent params: [minconf, maxconf, addresses]
		assert.GreaterOrEqual(t, len(req.Params), 1)
		minConf, ok := req.Params[0].(float64)
		assert.True(t, ok)
		assert.Equal(t, float64(6), minConf)

		resp := RPCResponse{
			Result: []interface{}{},
			Error:  nil,
			ID:     "1",
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "user", "pass")
	require.NoError(t, err)

	_, err = client.ListUnspent("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa")
	assert.NoError(t, err)
}

func TestUTXO_Value(t *testing.T) {
	utxo := UTXO{
		TxID:          "abc123",
		Vout:          0,
		Amount:        0.5,
		Confirmations: 10,
	}

	assert.Equal(t, "abc123", utxo.TxID)
	assert.Equal(t, uint32(0), utxo.Vout)
	assert.Equal(t, 0.5, utxo.Amount)
	assert.Equal(t, uint32(10), utxo.Confirmations)
}

func TestClient_GetBalance_UnexpectedFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := RPCResponse{
			Result: "not a map",
			Error:  nil,
			ID:     "1",
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "user", "pass")
	_, err := client.GetBalance("test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected response format")
}

func TestClient_GetBalance_NoBalanceKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := RPCResponse{
			Result: map[string]interface{}{
				"other": "field",
			},
			Error: nil,
			ID:    "1",
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "user", "pass")
	_, err := client.GetBalance("test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "balance not found")
}

func TestClient_GetBalance_TotalAmountFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := RPCResponse{
			Result: map[string]interface{}{
				"total_amount": float64(123.45),
			},
			Error: nil,
			ID:    "1",
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "user", "pass")
	balance, err := client.GetBalance("test")
	assert.NoError(t, err)
	assert.Equal(t, 123.45, balance)
}

func TestClient_ListUnspent_EmptyAddresses(t *testing.T) {
	client, _ := NewClient("http://localhost:8332", "user", "pass")
	utxos, err := client.ListUnspent()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(utxos))
}

func TestClient_ListUnspent_UnexpectedFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := RPCResponse{
			Result: "not an array",
			Error:  nil,
			ID:     "1",
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "user", "pass")
	_, err := client.ListUnspent("test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected response format")
}

func TestHelpers_InvalidTypes(t *testing.T) {
	m := map[string]interface{}{
		"string": 123,
		"float":  "not a float",
		"uint":   "not a uint",
	}

	// getString with non-string
	assert.Equal(t, "", getString(m, "string"))

	// getFloat64 with non-float
	assert.Equal(t, 0.0, getFloat64(m, "float"))

	// getUint32 with non-float
	assert.Equal(t, uint32(0), getUint32(m, "uint"))

	// Missing keys
	assert.Equal(t, "", getString(m, "missing"))
	assert.Equal(t, 0.0, getFloat64(m, "missing"))
	assert.Equal(t, uint32(0), getUint32(m, "missing"))
}

func TestClient_GetTransaction(t *testing.T) {
	tests := []struct {
		name           string
		txid           string
		serverResponse RPCResponse
		expectErr      bool
		expectedConfs  uint32
	}{
		{
			name: "valid transaction",
			txid: "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
			serverResponse: RPCResponse{
				Result: map[string]interface{}{
					"txid":          "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
					"hash":          "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
					"confirmations": float64(100),
					"blocktime":     float64(1231006505),
					"time":          float64(1231006505),
				},
				Error: nil,
				ID:    "1",
			},
			expectErr:     false,
			expectedConfs: 100,
		},
		{
			name: "transaction not found",
			txid: "invalid",
			serverResponse: RPCResponse{
				Result: nil,
				Error: &RPCError{
					Code:    -5,
					Message: "No such mempool or blockchain transaction",
				},
				ID: "1",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer server.Close()

			client, _ := NewClient(server.URL, "user", "pass")
			tx, err := client.GetTransaction(tt.txid)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.txid, tx.TxID)
				assert.Equal(t, tt.expectedConfs, tx.Confirmations)
			}
		})
	}
}

func TestClient_GetTransaction_UnexpectedFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := RPCResponse{
			Result: "not a map",
			Error:  nil,
			ID:     "1",
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "user", "pass")
	_, err := client.GetTransaction("test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected response format")
}

func TestClient_GetRawTransaction(t *testing.T) {
	tests := []struct {
		name           string
		txid           string
		serverResponse RPCResponse
		expectErr      bool
		expectedHex    string
	}{
		{
			name: "valid raw transaction",
			txid: "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
			serverResponse: RPCResponse{
				Result: "0100000001c997a5e56e104102fa209c6a852dd90660a20b2d9c352423edce25857fcd3704000000004847304402204e45e16932b8af514961a1d3a1a25fdf3f4f7732e9d624c6c61548ab5fb8cd410220181522ec8eca07de4860a4acdd12909d831cc56cbbac4622082221a8768d1d0901ffffffff0200ca9a3b00000000434104ae1a62fe09c5f51b13905f07f06b99a2f7159b2225f374cd378d71302fa28414e7aab37397f554a7df5f142c21c1b7303b8a0626f1baded5c72a704f7e6cd84cac00286bee0000000043410411db93e1dcdb8a016b49840f8c53bc1eb68a382e97b1482ecad7b148a6909a5cb2e0eaddfb84ccf9744464f82e160bfa9b8b64f9d4c03f999b8643f656b412a3ac00000000",
				Error:  nil,
				ID:     "1",
			},
			expectErr:   false,
			expectedHex: "0100000001c997a5e56e104102fa209c6a852dd90660a20b2d9c352423edce25857fcd3704000000004847304402204e45e16932b8af514961a1d3a1a25fdf3f4f7732e9d624c6c61548ab5fb8cd410220181522ec8eca07de4860a4acdd12909d831cc56cbbac4622082221a8768d1d0901ffffffff0200ca9a3b00000000434104ae1a62fe09c5f51b13905f07f06b99a2f7159b2225f374cd378d71302fa28414e7aab37397f554a7df5f142c21c1b7303b8a0626f1baded5c72a704f7e6cd84cac00286bee0000000043410411db93e1dcdb8a016b49840f8c53bc1eb68a382e97b1482ecad7b148a6909a5cb2e0eaddfb84ccf9744464f82e160bfa9b8b64f9d4c03f999b8643f656b412a3ac00000000",
		},
		{
			name: "transaction not found",
			txid: "invalid",
			serverResponse: RPCResponse{
				Result: nil,
				Error: &RPCError{
					Code:    -5,
					Message: "No such mempool or blockchain transaction",
				},
				ID: "1",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer server.Close()

			client, _ := NewClient(server.URL, "user", "pass")
			hexStr, err := client.GetRawTransaction(tt.txid)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedHex, hexStr)
			}
		})
	}
}

func TestClient_GetRawTransaction_UnexpectedFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := RPCResponse{
			Result: 123,
			Error:  nil,
			ID:     "1",
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "user", "pass")
	_, err := client.GetRawTransaction("test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected response format")
}

func TestTransaction_Value(t *testing.T) {
	tx := Transaction{
		TxID:          "abc123",
		Confirmations: 6,
		BlockTime:     1231006505,
	}

	assert.Equal(t, "abc123", tx.TxID)
	assert.Equal(t, uint32(6), tx.Confirmations)
	assert.Equal(t, uint64(1231006505), tx.BlockTime)
}
