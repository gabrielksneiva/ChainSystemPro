package rpc

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		user      string
		pass      string
		expectErr bool
	}{
		{
			name:      "valid client",
			url:       "http://localhost:8332",
			user:      "bitcoinrpc",
			pass:      "password",
			expectErr: false,
		},
		{
			name:      "empty url",
			url:       "",
			user:      "user",
			pass:      "pass",
			expectErr: true,
		},
		{
			name:      "invalid url",
			url:       "://invalid",
			user:      "user",
			pass:      "pass",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.url, tt.user, tt.pass)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestClient_CallRPC(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		params         []interface{}
		serverResponse RPCResponse
		serverStatus   int
		expectErr      bool
		expectedResult interface{}
	}{
		{
			name:   "successful call",
			method: "getblockchaininfo",
			params: []interface{}{},
			serverResponse: RPCResponse{
				Result: map[string]interface{}{
					"chain":  "main",
					"blocks": float64(800000),
				},
				Error: nil,
				ID:    "1",
			},
			serverStatus:   http.StatusOK,
			expectErr:      false,
			expectedResult: map[string]interface{}{"chain": "main", "blocks": float64(800000)},
		},
		{
			name:   "rpc error",
			method: "invalidmethod",
			params: []interface{}{},
			serverResponse: RPCResponse{
				Result: nil,
				Error: &RPCError{
					Code:    -32601,
					Message: "Method not found",
				},
				ID: "1",
			},
			serverStatus: http.StatusOK,
			expectErr:    true,
		},
		{
			name:         "http error",
			method:       "getblockchaininfo",
			params:       []interface{}{},
			serverStatus: http.StatusInternalServerError,
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify method
				assert.Equal(t, "POST", r.Method)

				// Verify authentication
				user, pass, ok := r.BasicAuth()
				assert.True(t, ok)
				assert.Equal(t, "testuser", user)
				assert.Equal(t, "testpass", pass)

				// Verify content type
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				// Parse request body
				var req RPCRequest
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)
				assert.Equal(t, tt.method, req.Method)
				assert.Equal(t, "1.0", req.JSONRPC)

				// Send response
				w.WriteHeader(tt.serverStatus)
				if tt.serverStatus == http.StatusOK {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			}))
			defer server.Close()

			// Create client
			client, err := NewClient(server.URL, "testuser", "testpass")
			require.NoError(t, err)

			// Call RPC
			result, err := client.CallRPC(tt.method, tt.params...)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestClient_CallRPC_WithRetry(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			// Simulate network error
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		// Success on third attempt
		resp := RPCResponse{
			Result: "success",
			Error:  nil,
			ID:     "1",
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "user", "pass")
	require.NoError(t, err)

	result, err := client.CallRPC("test")
	assert.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, 3, attemptCount)
}

func TestClient_CallRPC_MaxRetriesExceeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "user", "pass")
	require.NoError(t, err)

	// Set lower retry count for faster test
	client.maxRetries = 2
	client.retryDelay = 0

	_, err = client.CallRPC("test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max retries exceeded")
}

func TestRPCError_Error(t *testing.T) {
	rpcErr := &RPCError{
		Code:    -32600,
		Message: "Invalid Request",
	}
	assert.Equal(t, "RPC error -32600: Invalid Request", rpcErr.Error())
}

func TestClient_CallRPC_MarshalError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "user", "pass")
	require.NoError(t, err)

	// Pass unmarshalable type (channel)
	ch := make(chan int)
	_, err = client.CallRPC("test", ch)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal request")
}

func TestClient_CallRPC_InvalidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "user", "pass")
	require.NoError(t, err)

	_, err = client.CallRPC("test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse response")
}

func TestClient_CallRPC_NilParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req RPCRequest
		json.NewDecoder(r.Body).Decode(&req)

		// Verify params is empty array, not nil
		assert.NotNil(t, req.Params)
		assert.Equal(t, 0, len(req.Params))

		resp := RPCResponse{
			Result: "ok",
			ID:     "1",
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "user", "pass")
	require.NoError(t, err)

	result, err := client.CallRPC("test")
	assert.NoError(t, err)
	assert.Equal(t, "ok", result)
}
