package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// RPCRequest represents a JSON-RPC request
type RPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      string        `json:"id"`
}

// RPCResponse represents a JSON-RPC response
type RPCResponse struct {
	Result interface{} `json:"result"`
	Error  *RPCError   `json:"error"`
	ID     string      `json:"id"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("RPC error %d: %s", e.Code, e.Message)
}

// Client represents a Bitcoin RPC client
type Client struct {
	url        string
	user       string
	pass       string
	httpClient *http.Client
	maxRetries int
	retryDelay time.Duration
}

// NewClient creates a new Bitcoin RPC client
func NewClient(rpcURL, user, pass string) (*Client, error) {
	if rpcURL == "" {
		return nil, fmt.Errorf("rpc url cannot be empty")
	}

	// Validate URL
	parsedURL, err := url.Parse(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("invalid rpc url: %w", err)
	}
	if parsedURL.Scheme == "" {
		return nil, fmt.Errorf("invalid rpc url: missing scheme")
	}

	return &Client{
		url:  rpcURL,
		user: user,
		pass: pass,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		maxRetries: 3,
		retryDelay: 1 * time.Second,
	}, nil
}

// CallRPC executes a JSON-RPC call to the Bitcoin node
func (c *Client) CallRPC(method string, params ...interface{}) (interface{}, error) {
	if params == nil {
		params = []interface{}{}
	}

	req := RPCRequest{
		JSONRPC: "1.0",
		Method:  method,
		Params:  params,
		ID:      "1",
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 && c.retryDelay > 0 {
			time.Sleep(c.retryDelay)
		}

		result, err := c.doRequest(req)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Don't retry RPC errors (method not found, invalid params, etc)
		if _, ok := err.(*RPCError); ok {
			return nil, err
		}
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (c *Client) doRequest(req RPCRequest) (interface{}, error) {
	// Marshal request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", c.url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.SetBasicAuth(c.user, c.pass)

	// Execute request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http error %d: %s", resp.StatusCode, string(body))
	}

	// Parse JSON-RPC response
	var rpcResp RPCResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for RPC error
	if rpcResp.Error != nil {
		return nil, rpcResp.Error
	}

	return rpcResp.Result, nil
}
