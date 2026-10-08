# 13 — Statement Reconciliation & Break Management

## 1. Overview & Objective
`reconciliation` ingests MT940 and camt.053 clearing bank statements, enforces duplicate file check via file SHA256 checksum, matches internal transactions against external statements, and flags unmatched entries as observable `ReconciliationBreak`s.

## 2. Domain Model (DDD & CQRS Architecture)

### 2.1 Aggregates & Value Objects
- **Aggregates**: `ReconciliationStatement`, `ReconciliationBreak`
- **Value Objects**: `BreakType` (`UNMATCHED_INTERNAL`, `UNMATCHED_EXTERNAL`, `AMOUNT_MISMATCH`), `BreakStatus` (`OPEN`, `INVESTIGATING`, `RESOLVED`)

### 2.2 Commands & Events
- **Commands**:
  - `IngestStatementCommand`
  - `ResolveBreakCommand`
- **Domain Events**:
  - `StatementIngestedEvent`
  - `BreakDetectedEvent`

## 3. Database Schemas (`adapters/db/migration/0001_create_reconciliation_tables.up.sql`)
```sql
CREATE SCHEMA IF NOT EXISTS reconciliation;

CREATE TABLE reconciliation.reconciliation_statements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_checksum VARCHAR(64) NOT NULL UNIQUE,
    statement_reference VARCHAR(64) NOT NULL,
    rail_type VARCHAR(32) NOT NULL,
    raw_file_s3_ref VARCHAR(255),
    status VARCHAR(32) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE reconciliation.reconciliation_breaks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    statement_id UUID NOT NULL,
    payment_id UUID NOT NULL,
    break_type VARCHAR(32) NOT NULL,
    amount_difference NUMERIC(18,4),
    status VARCHAR(32) NOT NULL,
    resolved_by VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
```

## 4. Operations REST APIs
- `POST /api/v1/reconciliation/statements`
- `GET /api/v1/reconciliation/breaks`
