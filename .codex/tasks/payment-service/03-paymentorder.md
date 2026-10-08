# 03 — Payment Order, Idempotent Intake, and Audit

## 1. Overview & Objective

`paymentorder` is the **Sole State Authority** for all payment orders. It manages payment intake, enforces idempotency, owns the state machine transitions, and persists immutable audit logs in a single atomic PostgreSQL transaction.

## 2. Domain Model (DDD & CQRS Architecture)

### 2.1 Aggregates & Value Objects

- **Aggregate Root**: `PaymentOrder`
- **Value Objects**:
  - `Money` (Amount, Currency)
  - `AccountReference` (AccountNo, BSB/BIC, AccountName)
  - `PaymentState` (`received`, `structurally_validated`, `business_validated`, `pending_controls`, `approved`, `submission_pending`, `submitted`, `accepted`, `rejected`, `settled`, `returned`, `cancelled`, `failed`)
  - `IdempotencyKey` (TenantID, Operation, Key, RequestHash)

### 2.2 Domain Invariants & Rules

1. **State Machine Invariant**: Transitions are explicit methods on `PaymentOrder` (e.g. `TransitionToValidated()`, `TransitionToApproved()`). Forbidden transitions return `ErrInvalidStateTransition`.
2. **Optimistic Locking**: Every mutation checks and increments `version`.
3. **Atomicity**: Payment creation + Idempotency record + Audit Log + Outbox Event MUST execute in 1 PostgreSQL transaction.

### 2.3 Commands & Queries (CQRS)

- **Commands**:
  - `CreatePaymentOrderCommand`
  - `TransitionPaymentStateCommand`
- **Queries**:
  - `GetPaymentOrderByIDQuery`
  - `GetPaymentAuditHistoryQuery`

### 2.4 Domain Events (Outbox)

- `PaymentReceivedEvent`
- `PaymentStateChangedEvent`

## 3. Database Schemas (`adapters/db/migration/`)

### SQL Migration: `0001_create_paymentorder_tables.up.sql`

```sql
CREATE SCHEMA IF NOT EXISTS paymentorder;

CREATE TABLE paymentorder.payment_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_reference VARCHAR(64) NOT NULL UNIQUE,
    tenant_id VARCHAR(64) NOT NULL,
    debtor_account VARCHAR(34) NOT NULL,
    creditor_account VARCHAR(34) NOT NULL,
    amount NUMERIC(18,4) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    state VARCHAR(32) NOT NULL,
    version INT NOT NULL DEFAULT 1,
    raw_instruction_s3_ref VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE paymentorder.idempotency_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(64) NOT NULL,
    operation VARCHAR(64) NOT NULL,
    idempotency_key VARCHAR(128) NOT NULL,
    request_hash VARCHAR(64) NOT NULL,
    payment_id UUID NOT NULL,
    response_code INT,
    response_payload JSONB,
    status VARCHAR(32) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_idempotency UNIQUE (tenant_id, operation, idempotency_key)
);

CREATE TABLE paymentorder.payment_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID NOT NULL,
    from_state VARCHAR(32),
    to_state VARCHAR(32) NOT NULL,
    reason TEXT,
    actor_id VARCHAR(64),
    causation_id UUID,
    correlation_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## 4. API Specifications

### REST HTTP API (OpenAPI)

- `POST /api/v1/payments`
  - **Headers**: `X-Idempotency-Key` (required), `X-Tenant-ID` (required)
  - **Status 201**: Returns created payment order DTO
  - **Status 409 Conflict**: Returned when same idempotency key is used with a different request hash.

### In-Process Go Client (`paymentorder/api/module/client/client.go`)

```go
package client

import "context"

type Client interface {
    CreatePaymentOrder(ctx context.Context, cmd CreatePaymentCommand) (*PaymentOrderDTO, error)
    TransitionState(ctx context.Context, cmd TransitionStateCommand) error
}
```
