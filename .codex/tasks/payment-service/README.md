# Payment Service Implementation Tasks

Review and implement in numeric order. The governing design is
[SYSTEM_DESIGN.md](../../../services/payment/docs/SYSTEM_DESIGN.md); `AGENTS.md`
rules are mandatory.

| Order | Task | Depends on |
|---:|---|---|
| 00 | foundation and module skeleton | — |
| 00a | external core-banking fakes and HTTP stubs | 00 |
| 01 | durable messaging primitives | 00 |
| 02 | ISO 20022 module | 00, 01 |
| 03 | payment order and idempotent intake | 00, 01, 02 |
| 04–08 | validation, limits, screening, fees, routing | 03; 05–06 also require 00a |
| 09–14 | execution, status, exceptions, settlement, reconciliation, notifications | prior domain tasks; 09/12 require 00a |
| 15 | scale, operations and resilience | all relevant tasks |
