# 00_final — Payment Module Foundation, Architecture Baseline & Core Banking Stubs

## 1. Overview & Architectural Baseline
This specification establishes the foundational structure and external integration ports for the Go-based Modular Monolith payment service (`services/payment`).

The system comprises 13 isolated Bounded Contexts (modules). Each module owns its database schema, migrations, domain models, and public interfaces, while strictly preventing cross-module database writes or package import bypasses.

---

## 2. Bounded Context Structure & Standard Package Layout
Every module must conform strictly to the standard package layout:

```text
services/payment/modules/<module_name>/
├── module.go                     # Module composition root & RegisterContracts()
├── api/
│   └── module/
│       ├── client/
│       │   └── client.go         # Consumer-facing Go interfaces
│       └── dto/
│           └── dto.go            # Transport Data Transfer Objects
├── app/                          # CQRS Command/Query Handlers & Application Services
├── domain/                       # Pure Aggregates, Entities, Value Objects, Domain Invariants
└── adapters/                     # Secondary Adapters
    ├── db/
    │   ├── migration/            # Embedded SQL migration scripts (*.up.sql)
    │   └── repo/                 # PostgreSQL repositories
    ├── kafka/                    # Event producers & consumers
    └── http/                     # Outbound REST clients or Inbound REST handlers
```

---

## 3. Module Registration & Isolation Directives
1. **No Direct Package Imports**: Module A MUST NOT import internal packages (`app`, `domain`, `adapters`) of Module B. All inter-module communication occurs exclusively via `<module>/api/module/client`.
2. **Central Contract Registry**: `common/module/contracts.Contracts` holds references to all module clients for in-process synchronous invocation.
3. **Database Isolation**: Each module executes its own embedded migrations using separate PostgreSQL schemas (`paymentorder`, `limits`, `screening`, `fees`, `paymentrouting`, `paymentexecution`, `paymentstatus`, `exceptions`, `settlement`, `reconciliation`, `notifications`, `identity`, `otp`).

---

## 4. Core Banking Capability Ports (Go Interfaces)

To isolate Core Banking complexities, narrow consumer-owned Go interfaces (Ports) are declared under `common/ports/corebanking/`:

```go
package corebanking

import "context"

type AccountStatusPort interface {
    GetAccountStatus(ctx context.Context, accountID string) (*AccountStatusDTO, error)
}

type BalanceHoldPort interface {
    ReserveBalance(ctx context.Context, req ReserveBalanceRequest) (*ReserveBalanceResponse, error)
    ReleaseBalance(ctx context.Context, reservationID string) error
}

type LedgerPostingPort interface {
    PostLedger(ctx context.Context, req PostLedgerRequest) (*PostLedgerResponse, error)
    PostSettlementLedger(ctx context.Context, req PostSettlementLedgerRequest) (*PostSettlementLedgerResponse, error)
}

type CustomerKYCPort interface {
    GetCustomerKYC(ctx context.Context, customerID string) (*CustomerKYCDTO, error)
}

type ParticipantDirectoryPort interface {
    VerifyBIC(ctx context.Context, bic string) (bool, error)
}
```

---

## 5. Deterministic Test HTTP Stubs (`/stubs/core-banking/*`)

For local testing, CI/CD, and Sandbox environments, mock HTTP stub servers simulate Core Banking behaviors across control scenarios:

### 5.1 Endpoints
* `POST /stubs/core-banking/hold`: Simulates account balance reservation.
* `POST /stubs/core-banking/post`: Simulates ledger entries (Debit/Credit).
* `GET /stubs/core-banking/accounts/{id}`: Simulates account status lookup.

### 5.2 Control Scenarios via HTTP Headers
- `X-Stub-Scenario: APPROVED`: Returns `200 OK` with valid transaction/hold ID.
- `X-Stub-Scenario: DECLINED`: Returns `422 Unprocessable` with error code `INSUFFICIENT_FUNDS` or `ACCOUNT_FROZEN`.
- `X-Stub-Scenario: TIMEOUT`: Delay response by 30 seconds to test gateway timeout handling.
- `X-Stub-Scenario: DUPLICATE_REQUEST`: Returns previous result matching idempotency reference.

---

## 6. Implementation Checklist for AI Agent
- [ ] Implement composition roots (`module.go`) for all 13 modules.
- [ ] Declare public client interfaces under `modules/<name>/api/module/client/client.go`.
- [ ] Set up `common/module/contracts.Contracts` struct for in-process DI container.
- [ ] Implement Core Banking Port interfaces and HTTP Stubs with scenario header controls.
- [ ] Write startup verification test confirming all 13 module schemas migrate successfully without circular dependencies.
