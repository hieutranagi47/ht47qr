# 04 — Semantic Business Validation

## 1. Overview & Objective

`paymentvalidation` performs semantic business field validation after structural XML/XSD validation passes. It validates account formats, currency ISO codes, amount bounds, value dates, and duplicate business references without mutating `paymentorder` state directly.

## 2. Domain Model (DDD & CQRS Architecture)

### 2.1 Value Objects & Rules

- `ValidationRule` (AccountFormatRule, CurrencyRule, BusinessReferenceRule)
- `ValidationViolation` (Field, Code, Message)
- `ValidationResult` (IsValid, Violations)

### 2.2 Commands & Events

- **Command**: `ValidatePaymentCommand`
- **Domain Events**:
  - `PaymentValidatedEvent`
  - `PaymentValidationFailedEvent`

## 3. Database Schema (`adapters/db/migration/0001_create_validation_results.up.sql`)

```sql
CREATE SCHEMA IF NOT EXISTS paymentvalidation;

CREATE TABLE paymentvalidation.validation_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID NOT NULL,
    is_valid BOOLEAN NOT NULL,
    error_code VARCHAR(64),
    error_details JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## 4. API Specification & In-Process Client Interface

```go
package client

import "context"

type Client interface {
    ValidatePayment(ctx context.Context, req ValidationRequest) (*ValidationResultDTO, error)
}
```
