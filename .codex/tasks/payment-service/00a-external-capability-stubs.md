# 00a — External Core-Banking Capability Ports & HTTP Stubs

## 1. Overview & Objective
Isolates core banking complexities by defining narrow, consumer-owned Go interfaces (Ports) for balance holds, limit reservations, ledger posting, KYC lookup, and participant directory verification. Provides deterministic in-memory fakes and HTTP stub servers for local and Sandbox testing without external infrastructure.

## 2. Ports Definition (Go Interfaces)

```go
package corebanking

import "context"

type AccountStatusPort interface {
    GetAccountStatus(ctx context.Context, accountID string) (*AccountStatusDTO, error)
}

type BalanceHoldPort interface {
    ReserveBalance(ctx context.Context, req ReserveBalanceRequest) (*ReserveBalanceResponse, error)
    ReleaseBalance(ctx context.Context, req ReleaseBalanceRequest) error
}

type LedgerPostingPort interface {
    PostLedger(ctx context.Context, req LedgerPostRequest) (*LedgerPostResponse, error)
}
```

## 3. CQRS & Domain Contract DTOs
```go
type ReserveBalanceRequest struct {
    PaymentID              string `json:"payment_id"`
    AccountID              string `json:"account_id"`
    Amount                 int64  `json:"amount"` // in smallest currency unit (cents/vnd)
    Currency               string `json:"currency"`
    ExternalIdempotencyRef string `json:"external_idempotency_ref"`
}

type ReserveBalanceResponse struct {
    ReservationID string `json:"reservation_id"`
    Status        string `json:"status"` // APPROVED, DECLINED
    FailureReason string `json:"failure_reason,omitempty"`
}
```

## 4. Test HTTP Stub Endpoints & Scenario Controls
A test-only HTTP server handles stub requests and provides scenario injection headers/APIs.

### API Specifications (HTTP Stub Server)
- `POST /stubs/core-banking/accounts/hold`
- `POST /stubs/core-banking/ledger/post`
- `POST /stubs/control/scenario` (Set scenario state for tests)

### Supported Scenario Controls:
1. `APPROVED`: Returns success response with valid transaction reference.
2. `DECLINED`: Returns business decline (e.g., Insufficient Funds).
3. `NOT_FOUND`: Returns account not found.
4. `TIMEOUT`: Delays response past client HTTP timeout.
5. `TEMPORARY_FAILURE`: Returns HTTP 503 Service Unavailable.
6. `DUPLICATE_REQUEST`: Returns previous response using `ExternalIdempotencyRef`.

## 5. Security & Isolation Invariants
- **Strict Environment Guard**: HTTP Stubs must NEVER be loaded in production (`//build !production` or runtime environment checks).
- Reuses external idempotency references on retry calls to verify double-payment prevention.
