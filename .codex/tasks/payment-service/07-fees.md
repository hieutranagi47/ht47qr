# 07 — Fee Calculation Engine

## 1. Overview & Objective
`fees` calculates transaction fees based on versioned fee policies, determines charge bearers (`OUR`, `BEN`, `SHA`), applies currency rounding rules, and creates an immutable fee calculation snapshot.

## 2. Domain Model (DDD & CQRS Architecture)

### 2.1 Aggregates & Value Objects
- **Aggregate**: `FeeCalculation`
- **Value Objects**: `ChargeBearer`, `PolicyVersion`, `RoundingRule`

### 2.2 Commands & Events
- **Command**: `CalculateFeeCommand`
- **Domain Event**: `FeeCalculatedEvent`

## 3. Database Schema (`adapters/db/migration/0001_create_fee_calculations.up.sql`)
```sql
CREATE SCHEMA IF NOT EXISTS fees;

CREATE TABLE fees.fee_calculations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID NOT NULL,
    policy_version VARCHAR(32) NOT NULL,
    charge_bearer VARCHAR(8) NOT NULL,
    calculated_fee NUMERIC(18,4) NOT NULL,
    fee_currency VARCHAR(3) NOT NULL,
    rounding_rule VARCHAR(32),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## 4. In-Process Go Client Interface
```go
package client

import "context"

type Client interface {
    CalculateFee(ctx context.Context, req CalculateFeeRequest) (*FeeCalculationDTO, error)
}
```
