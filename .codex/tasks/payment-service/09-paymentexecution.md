# 09 — Rail Submission & Execution Intents

## 1. Overview & Objective
`paymentexecution` submits approved payments to external clearing rails via rail adapters. It persists **Durable Submission Intents** before network I/O to guarantee that retries reuse the same external idempotency reference and prevent double-payment.

## 2. Domain Model (DDD & CQRS Architecture)

### 2.1 Aggregates & Value Objects
- **Aggregate**: `SubmissionIntent`
- **Value Objects**: `RailType`, `AttemptNumber`, `RailIdempotencyRef`, `SubmissionStatus` (`PENDING`, `SUBMITTED`, `ACKNOWLEDGED`, `REJECTED`, `FAILED`)

### 2.2 Commands & Domain Events
- **Commands**:
  - `ExecuteSubmissionCommand`
  - `RetrySubmissionCommand`
- **Domain Events**:
  - `SubmissionAttemptedEvent`
  - `SubmissionAcceptedEvent`
  - `SubmissionRejectedEvent`

## 3. Database Schema (`adapters/db/migration/0001_create_submission_intents.up.sql`)
```sql
CREATE SCHEMA IF NOT EXISTS paymentexecution;

CREATE TABLE paymentexecution.submission_intents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID NOT NULL,
    attempt_number INT NOT NULL DEFAULT 1,
    rail_type VARCHAR(32) NOT NULL,
    rail_idempotency_ref VARCHAR(128) NOT NULL UNIQUE,
    status VARCHAR(32) NOT NULL,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## 4. Rail Adapter Interface (Egress Port)
```go
package adapter

import "context"

type RailAdapter interface {
    RailType() string
    Submit(ctx context.Context, intent *SubmissionIntent) (*RailResponseDTO, error)
}
```
