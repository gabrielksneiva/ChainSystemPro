-- Create extension for UUIDs
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Ledger entries table (immutable, append-only)
CREATE TABLE ledger_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entry_type VARCHAR(50) NOT NULL, -- deposit, withdrawal, transfer, mint, burn, fee
    chain_id VARCHAR(50) NOT NULL,
    address VARCHAR(255) NOT NULL,
    amount NUMERIC(78, 0) NOT NULL, -- Support up to 256-bit integers
    asset VARCHAR(100) NOT NULL,
    tx_hash VARCHAR(255),
    event_id UUID NOT NULL,
    metadata JSONB DEFAULT '{}',
    balance_after NUMERIC(78, 0),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT ledger_entries_amount_check CHECK (amount >= 0)
);

-- Indexes for ledger
CREATE INDEX idx_ledger_entries_chain_address ON ledger_entries(chain_id, address);
CREATE INDEX idx_ledger_entries_tx_hash ON ledger_entries(tx_hash);
CREATE INDEX idx_ledger_entries_event_id ON ledger_entries(event_id);
CREATE INDEX idx_ledger_entries_created_at ON ledger_entries(created_at);
CREATE INDEX idx_ledger_entries_entry_type ON ledger_entries(entry_type);

-- Transactions table
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chain_id VARCHAR(50) NOT NULL,
    tx_hash VARCHAR(255) UNIQUE,
    from_address VARCHAR(255) NOT NULL,
    to_address VARCHAR(255) NOT NULL,
    value NUMERIC(78, 0) NOT NULL,
    data BYTEA,
    nonce BIGINT,
    gas_limit BIGINT,
    gas_price NUMERIC(78, 0),
    max_fee_per_gas NUMERIC(78, 0),
    max_priority_fee NUMERIC(78, 0),
    signature BYTEA,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, confirmed, failed, dropped
    block_number BIGINT,
    confirmations BIGINT DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT transactions_value_check CHECK (value >= 0)
);

-- Indexes for transactions
CREATE INDEX idx_transactions_chain_id ON transactions(chain_id);
CREATE INDEX idx_transactions_tx_hash ON transactions(tx_hash);
CREATE INDEX idx_transactions_from_address ON transactions(from_address);
CREATE INDEX idx_transactions_to_address ON transactions(to_address);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);

-- Events table for event sourcing
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_type VARCHAR(100) NOT NULL,
    aggregate_id VARCHAR(255) NOT NULL,
    aggregate_type VARCHAR(100) NOT NULL,
    version INTEGER NOT NULL,
    payload JSONB NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT events_version_check CHECK (version > 0)
);

-- Indexes for events
CREATE INDEX idx_events_aggregate ON events(aggregate_id, aggregate_type, version);
CREATE INDEX idx_events_event_type ON events(event_type);
CREATE INDEX idx_events_created_at ON events(created_at);
CREATE UNIQUE INDEX idx_events_aggregate_version ON events(aggregate_id, aggregate_type, version);

-- Balances snapshot table (for performance optimization)
CREATE TABLE balance_snapshots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chain_id VARCHAR(50) NOT NULL,
    address VARCHAR(255) NOT NULL,
    asset VARCHAR(100) NOT NULL,
    balance NUMERIC(78, 0) NOT NULL,
    last_ledger_entry_id UUID NOT NULL,
    snapshot_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT balance_snapshots_balance_check CHECK (balance >= 0),
    UNIQUE(chain_id, address, asset)
);

-- Indexes for balance snapshots
CREATE INDEX idx_balance_snapshots_chain_address ON balance_snapshots(chain_id, address);
CREATE INDEX idx_balance_snapshots_snapshot_at ON balance_snapshots(snapshot_at);

-- Deposits table
CREATE TABLE deposits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chain_id VARCHAR(50) NOT NULL,
    address VARCHAR(255) NOT NULL,
    amount NUMERIC(78, 0) NOT NULL,
    asset VARCHAR(100) NOT NULL,
    tx_hash VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, processing, completed, failed
    fiat_amount NUMERIC(20, 2),
    fiat_currency VARCHAR(10),
    gateway_reference VARCHAR(255),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT deposits_amount_check CHECK (amount >= 0)
);

-- Indexes for deposits
CREATE INDEX idx_deposits_chain_address ON deposits(chain_id, address);
CREATE INDEX idx_deposits_status ON deposits(status);
CREATE INDEX idx_deposits_tx_hash ON deposits(tx_hash);
CREATE INDEX idx_deposits_created_at ON deposits(created_at);

-- Withdrawals table
CREATE TABLE withdrawals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chain_id VARCHAR(50) NOT NULL,
    from_address VARCHAR(255) NOT NULL,
    to_address VARCHAR(255) NOT NULL,
    amount NUMERIC(78, 0) NOT NULL,
    asset VARCHAR(100) NOT NULL,
    tx_hash VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, processing, completed, failed, cancelled
    fiat_amount NUMERIC(20, 2),
    fiat_currency VARCHAR(10),
    gateway_reference VARCHAR(255),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT withdrawals_amount_check CHECK (amount >= 0)
);

-- Indexes for withdrawals
CREATE INDEX idx_withdrawals_chain_from ON withdrawals(chain_id, from_address);
CREATE INDEX idx_withdrawals_status ON withdrawals(status);
CREATE INDEX idx_withdrawals_tx_hash ON withdrawals(tx_hash);
CREATE INDEX idx_withdrawals_created_at ON withdrawals(created_at);

-- Idempotency keys table for deduplication
CREATE TABLE idempotency_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(255) NOT NULL UNIQUE,
    request_hash VARCHAR(64) NOT NULL,
    response JSONB,
    status VARCHAR(20) NOT NULL, -- processing, completed, failed
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Indexes for idempotency
CREATE INDEX idx_idempotency_keys_key ON idempotency_keys(key);
CREATE INDEX idx_idempotency_keys_expires_at ON idempotency_keys(expires_at);

-- Event bus messages table (for Redis fallback/persistence)
CREATE TABLE event_bus_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stream VARCHAR(255) NOT NULL,
    event_id UUID NOT NULL UNIQUE,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    metadata JSONB DEFAULT '{}',
    published_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    retry_count INTEGER DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' -- pending, processing, processed, failed
);

-- Indexes for event bus messages
CREATE INDEX idx_event_bus_messages_stream ON event_bus_messages(stream);
CREATE INDEX idx_event_bus_messages_status ON event_bus_messages(status);
CREATE INDEX idx_event_bus_messages_published_at ON event_bus_messages(published_at);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for updated_at
CREATE TRIGGER update_transactions_updated_at BEFORE UPDATE ON transactions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_deposits_updated_at BEFORE UPDATE ON deposits
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_withdrawals_updated_at BEFORE UPDATE ON withdrawals
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
