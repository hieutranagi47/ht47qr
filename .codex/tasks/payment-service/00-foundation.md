# 00 — Payment Module Foundation & Architecture Baseline

## 1. Overview & Objective
This specification defines the foundational structure for the Go-based Modular Monolith payment service (`services/payment`). The system comprises 13 isolated Bounded Contexts (modules). Each module owns its database schema, migrations, domain models, and public interfaces, while strictly preventing cross-module database writes or package import bypasses.

## 2. Bounded Context Structure & Package Layout
Every module must conform strictly to the standard package layout:

```text
services/payment/modules/<module_name>/
├── module.go                     # Module composition root & RegisterContracts()
├── api/
│   └── module/
│       ├── client/
│       │   └── client.go         # Consumer-facing Go interfaces
│       └── dto/
│           └── dto.go            # Transport-agnostic Data Transfer Objects
├── app/
│   ├── command/                  # Command handlers (CQRS - Write side)
│   ├── query/                    # Query handlers (CQRS - Read side)
│   └── service/                  # Application orchestration services
├── domain/
│   ├── aggregate.go              # Aggregates & Domain Entities
│   ├── value_objects.go          # Value Objects & Invariants
│   ├── events.go                 # Domain Events
│   └── repository.go             # Repository interfaces
└── adapters/
    ├── db/
    │   ├── migration/            # SQL migrations (*.up.sql, *.down.sql)
    │   └── repository.go         # PostgreSQL implementation
    └── http/ or grpc/            # Inbound/Outbound HTTP handlers & clients
```

## 3. Composition Root & Contracts Registration
Modules register their public capability interfaces in `common/module/contracts.Contracts` during system bootstrapping in `cmd/server/main.go`.

### Go Interface: Contract Registration (`common/module/contracts.go`)
```go
package contracts

type Module interface {
    Name() string
    RegisterContracts(registry *Registry) error
}

type Registry struct {
    clients map[string]any
}

func (r *Registry) Register(clientName string, clientImpl any) {
    r.clients[clientName] = clientImpl
}
```

## 4. Database Isolation & Schema Management
Each module owns a dedicated PostgreSQL schema. Migrations are executed using embedded Go `embed.FS` during service startup.

### SQL Migration: `0001_init_schema.up.sql`
```sql
CREATE SCHEMA IF NOT EXISTS paymentorder;
CREATE SCHEMA IF NOT EXISTS common_infra;

CREATE TABLE IF NOT EXISTS common_infra.system_health (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    module_name VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL,
    checked_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## 5. Domain & Application Layer Guidelines (CQRS & DDD)
- **Domain Layer (`domain/`)**: Pure Go code containing business logic, aggregate invariants, and value objects. No external database, HTTP, or framework imports allowed.
- **Application Layer (`app/`)**: Orchestrates workflows, manages database transactions, handles CQRS Commands and Queries, and calls external ports or module clients.

## 6. Verification & Acceptance Criteria
- Service starts cleanly with all 13 modules initialized.
- Dependency graph check confirms zero circular dependencies or import bypasses.
- PostgreSQL migrations execute successfully for all isolated schemas.
