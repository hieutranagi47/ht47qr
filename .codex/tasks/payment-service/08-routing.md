# 08 — Payment Routing Engine

## 1. Overview & Objective
`paymentrouting` evaluates payment attributes (destination bank, amount, currency, time against channel Cut-off time) to deterministically select payment channels (`NAPAS`, `CITAD`, `SWIFT`, `INTERNAL`).

## 2. Domain Model (DDD & CQRS Architecture)

### 2.1 Value Objects & Entities
- **Aggregate**: `RoutingDecision`
- **Value Objects**: `PaymentRail` (`NAPAS_247`, `CITAD_RTGS`, `SWIFT_FIN`, `INTERNAL_TRANSFER`), `Priority` (`NORMAL`, `HIGH`, `CRITICAL`)

### 2.2 Commands & Events
- **Command**: `DetermineRouteCommand`
- **Domain Events**:
  - `PaymentRoutedEvent`
  - `RoutingFailedEvent`

## 3. Database Schema (`adapters/db/migration/0001_create_routing_decisions.up.sql`)
```sql
CREATE SCHEMA IF NOT EXISTS paymentrouting;

CREATE TABLE paymentrouting.routing_decisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID NOT NULL,
    selected_rail VARCHAR(32) NOT NULL,
    participant_id VARCHAR(64),
    priority VARCHAR(16) NOT NULL,
    policy_version VARCHAR(32) NOT NULL,
    cutoff_applied BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## 4. In-Process Go Client Interface
```go
package client

import "context"

type Client interface {
    DetermineRoute(ctx context.Context, req RouteRequest) (*RoutingDecisionDTO, error)
}
```
