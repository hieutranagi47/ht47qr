# 11 — Exceptions and Investigations

## 1. Overview & Objective
`exceptions` manages payment repair, recall, return, cancellation, and investigation workflows using a strict **Maker-Checker** operational process and immutable case audit timeline.

## 2. Domain Model (DDD & CQRS Architecture)

### 2.1 Aggregates & Value Objects
- **Aggregate**: `ExceptionCase`
- **Value Objects**:
  - `CaseType` (`REPAIR`, `RECALL`, `RETURN`, `CANCELLATION`)
  - `CaseStatus` (`OPEN`, `PENDING_CHECKER`, `APPROVED`, `REJECTED`, `CLOSED`)

### 2.2 Commands & Domain Events
- **Commands**:
  - `CreateExceptionCaseCommand` (Maker)
  - `ApproveExceptionCaseCommand` (Checker)
  - `RejectExceptionCaseCommand` (Checker)
- **Domain Events**:
  - `ExceptionCaseCreatedEvent`
  - `ExceptionCaseApprovedEvent`
  - `ExceptionCaseRejectedEvent`

## 3. Database Schema (`adapters/db/migration/0001_create_exception_cases.up.sql`)
```sql
CREATE SCHEMA IF NOT EXISTS exceptions;

CREATE TABLE exceptions.exception_cases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID NOT NULL,
    case_type VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL,
    maker_id VARCHAR(64) NOT NULL,
    checker_id VARCHAR(64),
    reason_code VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
```

## 4. REST Operations APIs
- `POST /api/v1/exceptions/cases`
- `POST /api/v1/exceptions/cases/{id}/approve`
- `POST /api/v1/exceptions/cases/{id}/reject`
