# 05 — Limits Evaluation & Balance Reservation

## 1. Overview & Objective

`limits` manages credit limit evaluations and durable balance holds by calling Core Banking Ports (`BalanceHoldPort`). It creates idempotent balance reservations and handles compensation/release upon payment rejection or cancellation.

## 2. Domain Model (DDD & CQRS Architecture)

### 2.1 Aggregates & Value Objects

- **Aggregate**: `LimitReservation`
- **Value Objects**: `HoldStatus` (`RESERVED`, `RELEASED`, `CONSUMED`, `EXPIRED`), `HoldReference`

### 2.2 Commands & Domain Events

- **Commands**:
  - `ReserveLimitCommand`
  - `ReleaseLimitCommand`
- **Domain Events**:
  - `LimitReservedEvent`
  - `LimitReleasedEvent`
  - `LimitReservationFailedEvent`

## 3. Database Schema (`adapters/db/migration/0001_create_limit_reservations.up.sql`)

```sql
CREATE SCHEMA IF NOT EXISTS limits;

CREATE TABLE limits.limit_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID NOT NULL,
    account_id VARCHAR(34) NOT NULL,
    reserved_amount NUMERIC(18,4) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(32) NOT NULL,
    core_banking_ref VARCHAR(128),
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
```

## 4. API Specification (`limits/api/module/client/client.go`)

```go
package client

import "context"

type Client interface {
    ReserveLimit(ctx context.Context, req ReserveLimitReq) (*ReservationDTO, error)
    ReleaseLimit(ctx context.Context, paymentID string) error
}
```
