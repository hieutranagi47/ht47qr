# 06 — AML, Sanctions, and Fraud Screening

## 1. Overview & Objective
`screening` integrates compliance checks for Anti-Money Laundering (AML), Sanctions list screening, and Fraud scoring via vendor adapters. It manages compliance cases requiring manual review (Maker-Checker).

## 2. Domain Model (DDD & CQRS Architecture)

### 2.1 Aggregates & Value Objects
- **Aggregate**: `ScreeningCase`
- **Value Objects**:
  - `ScreeningType` (`AML`, `SANCTIONS`, `FRAUD`)
  - `ScreeningStatus` (`CLEAR`, `HIT`, `REVIEW_PENDING`, `UNAVAILABLE`)
  - `MatchScore`

### 2.2 Commands & Events
- **Commands**:
  - `InitiateScreeningCommand`
  - `ReviewScreeningCaseCommand` (Maker-Checker approval)
- **Domain Events**:
  - `ScreeningPassedEvent`
  - `ScreeningHitEvent`
  - `ScreeningFailedEvent`

## 3. Database Schema (`adapters/db/migration/0001_create_screening_cases.up.sql`)
```sql
CREATE SCHEMA IF NOT EXISTS screening;

CREATE TABLE screening.screening_cases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID NOT NULL,
    screening_type VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL,
    vendor_name VARCHAR(64),
    vendor_reference VARCHAR(128),
    match_score NUMERIC(5,2),
    decision_by VARCHAR(64),
    decision_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
```

## 4. REST API for Compliance Operations
- `POST /api/v1/screening/cases/{id}/review`
  - **Body**: `{"decision": "CLEAR" | "REJECT", "reason": "Verified identity"}`
  - **Requires Role**: `COMPLIANCE_OFFICER`
