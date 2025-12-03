package rpc

import "fmt"

// UTXO represents an unspent transaction output
type UTXO struct {
	TxID          string  `json:"txid"`
	Vout          uint32  `json:"vout"`
	Address       string  `json:"address"`
	ScriptPubKey  string  `json:"scriptPubKey"`
	Amount        float64 `json:"amount"`
	Confirmations uint32  `json:"confirmations"`
}

// Transaction represents a Bitcoin transaction
type Transaction struct {
	TxID          string `json:"txid"`
	Hash          string `json:"hash"`
	Confirmations uint32 `json:"confirmations"`
	BlockTime     uint64 `json:"blocktime"`
	Time          uint64 `json:"time"`
}

// GetBalance retrieves the balance for a given Bitcoin address
func (c *Client) GetBalance(address string) (float64, error) {
	// Bitcoin Core doesn't have a direct "getbalance" for arbitrary addresses
	// We use scantxoutset to get the balance
	result, err := c.CallRPC("scantxoutset", "start", []string{"addr(" + address + ")"})
	if err != nil {
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}

	// Parse result
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("unexpected response format")
	}

	balance, ok := resultMap["balance"].(float64)
	if !ok {
		// If no balance key, it might be total_amount
		balance, ok = resultMap["total_amount"].(float64)
		if !ok {
			return 0, fmt.Errorf("balance not found in response")
		}
	}

	return balance, nil
}

// ListUnspent retrieves all unspent transaction outputs for given addresses
func (c *Client) ListUnspent(addresses ...string) ([]UTXO, error) {
	if len(addresses) == 0 {
		return []UTXO{}, nil
	}

	// Call listunspent with minconf=1, maxconf=9999999, addresses
	result, err := c.CallRPC("listunspent", 6, 9999999, addresses)
	if err != nil {
		return nil, fmt.Errorf("failed to list unspent: %w", err)
	}

	// Parse result
	resultArray, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	utxos := make([]UTXO, 0, len(resultArray))
	for _, item := range resultArray {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		utxo := UTXO{
			TxID:          getString(itemMap, "txid"),
			Vout:          getUint32(itemMap, "vout"),
			Address:       getString(itemMap, "address"),
			ScriptPubKey:  getString(itemMap, "scriptPubKey"),
			Amount:        getFloat64(itemMap, "amount"),
			Confirmations: getUint32(itemMap, "confirmations"),
		}

		utxos = append(utxos, utxo)
	}

	return utxos, nil
}

// Helper functions to safely extract values from maps
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getFloat64(m map[string]interface{}, key string) float64 {
	if val, ok := m[key].(float64); ok {
		return val
	}
	return 0
}

func getUint64(m map[string]interface{}, key string) uint64 {
	if val, ok := m[key].(float64); ok {
		return uint64(val)
	}
	return 0
}

// GetTransaction retrieves transaction details by TXID
func (c *Client) GetTransaction(txid string) (*Transaction, error) {
	result, err := c.CallRPC("gettransaction", txid)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	tx := &Transaction{
		TxID:          getString(resultMap, "txid"),
		Hash:          getString(resultMap, "hash"),
		Confirmations: getUint32(resultMap, "confirmations"),
		BlockTime:     getUint64(resultMap, "blocktime"),
		Time:          getUint64(resultMap, "time"),
	}

	return tx, nil
}

// GetRawTransaction retrieves the raw transaction hex by TXID
func (c *Client) GetRawTransaction(txid string) (string, error) {
	// Call with verbose=false to get hex string
	result, err := c.CallRPC("getrawtransaction", txid, false)
	if err != nil {
		return "", fmt.Errorf("failed to get raw transaction: %w", err)
	}

	hexStr, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}

	return hexStr, nil
}

func getUint32(m map[string]interface{}, key string) uint32 {
	if val, ok := m[key].(float64); ok {
		return uint32(val)
	}
	return 0
}
