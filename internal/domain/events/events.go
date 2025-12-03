package events

import (
	"math/big"
	"time"

	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/entities"
	"github.com/gabrielksneiva/ChainSystemPro/internal/domain/valueobjects"
	"github.com/google/uuid"
)

// EventType represents the type of domain event
type EventType string

const (
	EventTypeTransactionCreated     EventType = "transaction.created"
	EventTypeTransactionSigned      EventType = "transaction.signed"
	EventTypeTransactionBroadcasted EventType = "transaction.broadcasted"
	EventTypeTransactionConfirmed   EventType = "transaction.confirmed"
	EventTypeTransactionFailed      EventType = "transaction.failed"
	EventTypeBalanceQueried         EventType = "balance.queried"
	EventTypeFeeEstimated           EventType = "fee.estimated"
)

// BaseEvent contains common event fields
type BaseEvent struct {
	ID        string    `json:"id"`
	Type      EventType `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	ChainID   string    `json:"chain_id"`
}

// NewBaseEvent creates a new base event
func NewBaseEvent(eventType EventType, chainID string) BaseEvent {
	return BaseEvent{
		ID:        uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now(),
		ChainID:   chainID,
	}
}

// TransactionCreatedEvent is published when a transaction is created
type TransactionCreatedEvent struct {
	BaseEvent
	TransactionID string `json:"transaction_id"`
	From          string `json:"from"`
	To            string `json:"to"`
	Value         string `json:"value"`
	Data          []byte `json:"data,omitempty"`
	Nonce         uint64 `json:"nonce"`
	GasLimit      uint64 `json:"gas_limit"`
	GasPrice      string `json:"gas_price,omitempty"`
}

// NewTransactionCreatedEvent creates a new transaction created event
func NewTransactionCreatedEvent(tx *entities.Transaction) *TransactionCreatedEvent {
	event := &TransactionCreatedEvent{
		BaseEvent:     NewBaseEvent(EventTypeTransactionCreated, tx.ChainID()),
		TransactionID: tx.ID(),
		From:          tx.From().String(),
		To:            tx.To().String(),
		Value:         tx.Value().String(),
		Data:          tx.Data(),
		GasLimit:      tx.GasLimit(),
	}

	if tx.Nonce() != nil {
		event.Nonce = tx.Nonce().Value()
	}

	if tx.GasPrice() != nil {
		event.GasPrice = tx.GasPrice().String()
	}

	return event
}

// TransactionSignedEvent is published when a transaction is signed
type TransactionSignedEvent struct {
	BaseEvent
	TransactionID string `json:"transaction_id"`
	Hash          string `json:"hash"`
	Signature     string `json:"signature"`
}

// NewTransactionSignedEvent creates a new transaction signed event
func NewTransactionSignedEvent(tx *entities.Transaction) *TransactionSignedEvent {
	event := &TransactionSignedEvent{
		BaseEvent:     NewBaseEvent(EventTypeTransactionSigned, tx.ChainID()),
		TransactionID: tx.ID(),
	}

	if tx.Hash() != nil {
		event.Hash = tx.Hash().Hex()
	}

	if tx.Signature() != nil {
		event.Signature = tx.Signature().Hex()
	}

	return event
}

// TransactionBroadcastedEvent is published when a transaction is broadcasted
type TransactionBroadcastedEvent struct {
	BaseEvent
	TransactionID string `json:"transaction_id"`
	Hash          string `json:"hash"`
}

// NewTransactionBroadcastedEvent creates a new transaction broadcasted event
func NewTransactionBroadcastedEvent(chainID, txID string, hash *valueobjects.Hash) *TransactionBroadcastedEvent {
	return &TransactionBroadcastedEvent{
		BaseEvent:     NewBaseEvent(EventTypeTransactionBroadcasted, chainID),
		TransactionID: txID,
		Hash:          hash.Hex(),
	}
}

// TransactionConfirmedEvent is published when a transaction is confirmed
type TransactionConfirmedEvent struct {
	BaseEvent
	TransactionID string `json:"transaction_id"`
	Hash          string `json:"hash"`
	BlockNumber   uint64 `json:"block_number"`
	Confirmations uint64 `json:"confirmations"`
}

// NewTransactionConfirmedEvent creates a new transaction confirmed event
func NewTransactionConfirmedEvent(tx *entities.Transaction) *TransactionConfirmedEvent {
	event := &TransactionConfirmedEvent{
		BaseEvent:     NewBaseEvent(EventTypeTransactionConfirmed, tx.ChainID()),
		TransactionID: tx.ID(),
		BlockNumber:   tx.BlockNumber(),
		Confirmations: tx.Confirmations(),
	}

	if tx.Hash() != nil {
		event.Hash = tx.Hash().Hex()
	}

	return event
}

// TransactionFailedEvent is published when a transaction fails
type TransactionFailedEvent struct {
	BaseEvent
	TransactionID string `json:"transaction_id"`
	Hash          string `json:"hash,omitempty"`
	Reason        string `json:"reason"`
	ErrorCode     string `json:"error_code,omitempty"`
}

// NewTransactionFailedEvent creates a new transaction failed event
func NewTransactionFailedEvent(chainID, txID, hash, reason, errorCode string) *TransactionFailedEvent {
	return &TransactionFailedEvent{
		BaseEvent:     NewBaseEvent(EventTypeTransactionFailed, chainID),
		TransactionID: txID,
		Hash:          hash,
		Reason:        reason,
		ErrorCode:     errorCode,
	}
}

// BalanceQueriedEvent is published when a balance is queried
type BalanceQueriedEvent struct {
	BaseEvent
	Address      string `json:"address"`
	Balance      string `json:"balance"`
	TokenAddress string `json:"token_address,omitempty"`
}

// NewBalanceQueriedEvent creates a new balance queried event
func NewBalanceQueriedEvent(chainID string, address *valueobjects.Address, balance *big.Int, tokenAddress *valueobjects.Address) *BalanceQueriedEvent {
	event := &BalanceQueriedEvent{
		BaseEvent: NewBaseEvent(EventTypeBalanceQueried, chainID),
		Address:   address.String(),
		Balance:   balance.String(),
	}

	if tokenAddress != nil {
		event.TokenAddress = tokenAddress.String()
	}

	return event
}

// FeeEstimatedEvent is published when a fee is estimated
type FeeEstimatedEvent struct {
	BaseEvent
	TransactionID  string `json:"transaction_id,omitempty"`
	GasLimit       uint64 `json:"gas_limit"`
	GasPrice       string `json:"gas_price,omitempty"`
	MaxFeePerGas   string `json:"max_fee_per_gas,omitempty"`
	MaxPriorityFee string `json:"max_priority_fee,omitempty"`
	Total          string `json:"total"`
	Currency       string `json:"currency"`
}

// NewFeeEstimatedEvent creates a new fee estimated event
func NewFeeEstimatedEvent(chainID string, fee *entities.Fee, txID string) *FeeEstimatedEvent {
	event := &FeeEstimatedEvent{
		BaseEvent:     NewBaseEvent(EventTypeFeeEstimated, chainID),
		TransactionID: txID,
		GasLimit:      fee.GasLimit(),
		Total:         fee.Total().String(),
		Currency:      fee.Currency(),
	}

	if fee.GasPrice() != nil {
		event.GasPrice = fee.GasPrice().String()
	}

	if fee.MaxFeePerGas() != nil {
		event.MaxFeePerGas = fee.MaxFeePerGas().String()
	}

	if fee.MaxPriorityFee() != nil {
		event.MaxPriorityFee = fee.MaxPriorityFee().String()
	}

	return event
}
