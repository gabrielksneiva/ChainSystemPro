-- Drop triggers
DROP TRIGGER IF EXISTS update_withdrawals_updated_at ON withdrawals;
DROP TRIGGER IF EXISTS update_deposits_updated_at ON deposits;
DROP TRIGGER IF EXISTS update_transactions_updated_at ON transactions;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables
DROP TABLE IF EXISTS event_bus_messages;
DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS withdrawals;
DROP TABLE IF EXISTS deposits;
DROP TABLE IF EXISTS balance_snapshots;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS ledger_entries;

-- Drop extension
DROP EXTENSION IF EXISTS "uuid-ossp";
