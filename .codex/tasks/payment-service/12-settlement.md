# 12 — Settlement and Prefunding Operations

## 1. Overview & Objective
`settlement` manages clearing cycle closures, multilateral position netting, prefunding checks, and posts settlement ledger entries to Core Banking via `LedgerPostingPort`.

## 2. Domain Model (DDD & CQRS Architecture)

### 2.1 Aggregates & Value Objects
- **Aggregates**: `SettlementCycle`, `SettlementEntry`
- **Value Objects**: `CycleStatus` (`OPEN`, `CLOSING`, `SETTLED`, `FAILED`), `NetPosition`

### 2.2 Commands & Domain Events
- **Commands**:
  - `OpenSettlementCycleCommand`
  - `CloseSettlementCycleCommand`
- **Domain Event**: `SettlementCycleClosedEvent`

## 3. Database Schemas (`adapters/db/migration/0001_create_settlement_tables.up.sql`)
```sql
CREATE SCHEMA IF NOT EXISTS settlement;

CREATE TABLE settlement.settlement_cycles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cycle_reference VARCHAR(64) NOT NULL UNIQUE,
    rail_type VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL,
    total_debit NUMERIC(18,4) DEFAULT 0,
    total_credit NUMERIC(18,4) DEFAULT 0,
    net_position NUMERIC(18,4) DEFAULT 0,
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE settlement.settlement_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cycle_id UUID NOT NULL,
    payment_id UUID NOT NULL,
    amount NUMERIC(18,4) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    entry_type VARCHAR(8) NOT NULL, -- DEBIT, CREDIT
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## 4. REST Admin APIs
- `POST /api/v1/settlements/close-cycle`
