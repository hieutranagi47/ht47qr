# 03_final — Payment Order Aggregate, Two-Phase Intake, Inbound & Outbound ISO 20022 Flows Specification

## 1. Overview & Single State Authority Governance
`paymentorder` is the **Sole State Authority** for all payment orders in the system. It manages payment intake, owns state machine transitions, enforces idempotency, and persists immutable audit logs in a single atomic PostgreSQL transaction.

---

## 2. Domain Model & State Machine

### 2.1 State Transitions
Allowed state sequence:
`received` $\rightarrow$ `structurally_validated` $\rightarrow$ `business_validated` $\rightarrow$ `pending_controls` $\rightarrow$ `approved` $\rightarrow$ `submission_pending` $\rightarrow$ `submitted` $\rightarrow$ `accepted` / `rejected` $\rightarrow$ `settled` / `returned` / `failed`.

### 2.2 Governance Rule
Outside modules propose state changes or submit result events via module client interfaces (`api/module/client`). Direct DB writes to `paymentorder` schema from external modules are strictly prohibited.

---

## 3. Two-Phase Biometric Intake Pipeline (`03a`)

```text
[Mobile App]                   [Payment Gateway / 03a]           [Redis]           [otp / identity]          [paymentorder 03]
     │                                │                             │                     │                            │
     │ 1. POST /initiate              │                             │                     │                            │
     ├───────────────────────────────►│ 2. Check balance & limits   │                     │                            │
     │                                │ 3. Store draft payload 120s ├────────────────────►│                            │
     │                                │ 4. Generate Challenge Key   │                     ├───────────────────────────►│
     │ 5. Return Trans_ID + Challenge │                             │                     │                            │
     │◄───────────────────────────────┤                             │                     │                            │
  [Biometric Sign at TEE]             │                             │                     │                            │
     │                                │                             │                     │                            │
     │ 6. POST /execute               │                             │                     │                            │
     ├───────────────────────────────►│ 7. Fetch draft & verify key ├────────────────────►│                            │
     │                                │ 8. Verify digital signature │                     ├───────────────────────────►│
     │                                │ 9. (If >20M) Verify TOTP    │                     ├───────────────────────────►│
     │                                │ 10. Atomic Commit PostgreSQL│                     │                            │
     │                                ├─────────────────────────────┼─────────────────────┼───────────────────────────►│
     │ 11. HTTP 200/202 Response      │                             │                     │   (Create Payment Order)   │
     │◄───────────────────────────────┤                             │                     │                            │
```

### 3.1 Phase 1 Initiation (`POST /api/v1/payments/initiate`)
* Validates preliminary limits (`limits` module).
* Stores draft payload in Redis `payment:draft:{Trans_ID}` with TTL = 120s.
* Generates `Challenge_Key` via `otp` module.
* Returns `HTTP 200 OK` containing `transaction_id`, `challenge_key`, and `expires_in_seconds: 120`.

### 3.2 Phase 2 Execution (`POST /api/v1/payments/execute`)
* Idempotency check against `(tenant_id, operation, idempotency_key)` and Request Hash.
* Fetches draft payload, consumes and deletes `Challenge_Key` from Redis.
* Verifies cryptographic signature via `identity` module.
* If amount $\ge 20,000,000\text{ VND}$ (Decision 2345/QD-NHNN), verifies 6-digit Smart OTP code via `otp` module.
* **Atomic PostgreSQL Transaction**: Persists `payment_orders` (`state: 'received'`), `idempotency_keys`, `payment_audit_logs`, and `watermill_events_to_forward` (Outbox event: `PaymentReceivedEvent`).

---

## 4. Inbound ISO 20022 XML Payment Flow (`03b`)

Processing pipeline for **Inbound Credit Transfers (`pacs.008.001.08`)** received from external banks via clearing rails (CITAD, NAPAS 24/7, SWIFT):

1. **Inbound Endpoint**: `POST /api/v1/rails/{rail_type}/inbound-payments` receives raw XML.
2. **Audit Storage**: Saves raw XML payload to S3/MinIO bucket.
3. **Idempotency**: Deduplicates against `(rail_type, EndToEndId)` in `paymentorder.idempotency_keys`.
4. **Beneficiary Account Check (`04`)**: Verifies `CreditorAccount` active status in Core Banking.
5. **AML Screening (`06`)**: Screens recipient name/account against sanction lists.
6. **Ledger Posting (`00a` / `12`)**: Calls `CoreBankingClient.PostLedger` with `CREDIT` entry.
7. **Status Report (`10`)**: Returns synchronous or asynchronous `pacs.002.001.10` XML with status `ACCP` (Accepted) or `RJCT` (Rejected with ISO reason code: `AC01` - Invalid Account, `LEGL` - Sanctions Block).
8. **Notification (`14`)**: Dispatches async Push Notification / SMS to beneficiary.

---

## 5. Outbound ISO 20022 Payment Flow & Settlement Methods (`03c`)

Pipeline for **Outbound Credit Transfers (`pacs.008.001.08` & `pacs.009.001.08 COVE`)**:

### 5.1 Settlement Method Selection (`08 paymentrouting`)
1. **Direct Clearing (`CLRG`)**: Format single `pacs.008` XML for national clearing systems (CITAD/NAPAS), settled via Central Bank accounts.
2. **Serial Method (`IND`)**: Send `pacs.008` directly through a chain of correspondent banks holding direct Nostro/Vostro accounts.
3. **Cover Payment Method (`COVE`)**: Generate two parallel messages sharing the SAME `UETR`:
   - `pacs.008` sent directly to Beneficiary Bank containing customer remittance info.
   - `pacs.009 COVE` sent to Intermediary/Reimbursement Bank for Nostro/Vostro liquidity settlement.

### 5.2 Nostro, Vostro, Loro Account Accounting
* **Nostro Account**: Our foreign currency account held at a foreign correspondent bank.
* **Vostro Account**: Foreign bank's account held at our bank in local/foreign currency.
* **Loro Account**: Third-party bank account held at another correspondent bank used for liquidity reference.
* **Outbound Ledger Entry**: Debit Customer Account $\rightarrow$ Credit Nostro/Vostro Account.

### 5.3 Inbound Cover Matching Engine
* Correlates incoming `pacs.008` and `pacs.009 COVE` by `UETR`.
* If `pacs.008` arrives without matching Cover, hold in `pending_settlement_cover` for up to 48 hours before triggering investigation (`camt.029` / `11 exceptions`).

---

## 6. Consolidated Database Schemas

```sql
-- Core Payment Order Table
CREATE TABLE paymentorder.payment_orders (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  payment_reference VARCHAR NOT NULL UNIQUE,
  tenant_id VARCHAR NOT NULL,
  direction VARCHAR(10) NOT NULL DEFAULT 'OUTBOUND', -- 'INBOUND' or 'OUTBOUND'
  debtor_account VARCHAR NOT NULL,
  creditor_account VARCHAR NOT NULL,
  amount DECIMAL(18,4) NOT NULL,
  currency VARCHAR(3) NOT NULL,
  settlement_method VARCHAR(10) NOT NULL DEFAULT 'CLRG', -- 'CLRG', 'IND', 'COVE'
  state VARCHAR NOT NULL,
  version INT NOT NULL DEFAULT 1,
  external_end_to_end_id VARCHAR(100),
  uetr VARCHAR(36),
  debtor_agent_bic VARCHAR(11),
  creditor_agent_bic VARCHAR(11),
  remittance_information TEXT,
  raw_instruction_s3_ref VARCHAR,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  updated_at TIMESTAMP NOT NULL DEFAULT now(),
  deleted_at TIMESTAMP
);

-- Idempotency Keys Table
CREATE TABLE paymentorder.idempotency_keys (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id VARCHAR NOT NULL,
  operation VARCHAR NOT NULL,
  idempotency_key VARCHAR NOT NULL,
  request_hash VARCHAR NOT NULL,
  payment_id UUID NOT NULL,
  response_code INT,
  response_payload JSONB,
  status VARCHAR NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  updated_at TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT uq_idempotency UNIQUE (tenant_id, operation, idempotency_key)
);

-- Payment Audit Logs
CREATE TABLE paymentorder.payment_audit_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  payment_id UUID NOT NULL,
  from_state VARCHAR,
  to_state VARCHAR NOT NULL,
  reason TEXT,
  actor_id VARCHAR,
  causation_id UUID,
  correlation_id UUID,
  created_at TIMESTAMP NOT NULL DEFAULT now()
);

-- Outbox Table
CREATE TABLE paymentorder.watermill_events_to_forward (
  transaction_id BIGINT NOT NULL,
  offset BIGINT NOT NULL,
  uuid VARCHAR NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  payload JSONB,
  metadata JSONB,
  PRIMARY KEY (transaction_id, offset)
);
```

---

## 7. Implementation Checklist & Test Matrix for AI Agent
- [x] Implement `paymentorder` aggregate root with strict state machine transition validation.
- [x] Implement Two-Phase Intake API endpoints: `POST /initiate` and `POST /execute`.
- [x] Implement Inbound ISO 20022 XML Handler `POST /rails/{rail}/inbound-payments` with `pacs.002` response.
- [x] Implement Outbound Routing & Settlement Selector (`CLRG`, `IND`, `COVE` with matching `UETR`).
- [x] Implement Cover Matching Engine correlating `pacs.008` and `pacs.009 COVE`.
- [x] Write integration test cases: `TS-ORD-01`, `TS-ORD-02`, `TS-ISO-01`, `TS-ISO-02`, `TS-ISO-03`.


## 8. Implemented runtime contracts and verification

`paymentorder` stores the immutable instruction snapshot in JSONB, with separate
state, version, reference and direction columns. Creation, the globally unique
idempotency tuple, initial audit entry and Watermill outbox envelope commit in one
PostgreSQL transaction. State changes use explicit aggregate methods under a row
lock and an expected version, with audit and outbox writes in the same transaction.
Early validation/control rejection is legal; rejected and failed orders cannot be
settled. A settled order may be returned. Currency and precision use the checked-in
`common/shared.Currency` catalogue.

Both JSON endpoints require `Idempotency-Key` and the `payments:authenticate`
authorization scope. Initiation returns the exact `signing_payload` snapshot in
addition to the transaction/challenge/expiry fields. The device signs
`challenge_key || CanonicalJSON(signing_payload)` using the existing identity
protocol and sends the signature as standard Base64. Initiation replay uses the
same draft/challenge during its TTL and reports the remaining expiry. Execution
returns HTTP 202 with the stable received result. A committed replay is resolved
from PostgreSQL before touching Redis; different payloads under one key conflict.
A concurrent execution returns HTTP 409, slug `execute_pending`. Expired drafts
return HTTP 410. Redis leases are best-effort; they do not establish authoritative
payment idempotency. A consumed challenge cannot be retried after an authentication
failure: initiate a new draft. Smart OTP is required at **>= 20,000,000 VND**.

Rail intake requires `payments:rail`, a supported rail, a valid UETR and a unique
EndToEndId; this increment accepts one transfer per XML message. Customer transfers
are deduplicated globally per rail, independently of the authenticated rail client.
Original XML is stored in a private S3/MinIO bucket and only its internal object
reference enters the immutable instruction. The API returns schema-valid
pacs.002.001.10 ACCP or RJCT (AC01/LEGL). Temporary account, screening and ledger
failures are retried without converting them into permanent rejections.

Cover legs are settlement records and cannot create a second customer credit.
Either arrival order is supported; amount, currency and both accounts must match
as well as rail/UETR. Customer orders remain approved while settlement records
represent `pending_settlement_cover`; HTTP 409 documents that pending result.
The first leg starts the 48-hour deadline. An expired leg cannot automatically
resume a credit. The settlement worker publishes a durable expiry event; exceptions
opens an idempotent investigation case with `requested_message = camt.029`.
This case queues the investigation workflow; it does not transmit a camt.029 message
to an external bank. A screening review can propose approval through paymentorder;
the approval event resumes execution.

Outbound CLRG and IND generate one customer transfer; application method IND maps
to ISO settlement code INDA. COVE generates a customer transfer plus a cover
transfer to the configured reimbursement bank, preserving UETR. Liquidity is
checked before debit/dispatch. The core-banking debit request carries the configured
Nostro counter-account for an atomic ledger transfer at the authoritative provider.
Ledger and rail intent/reference/payload are persisted before each external call.
Retries reuse the exact payload and idempotency reference, including when a
previous rail call succeeded but its acknowledgement was lost. Outbound intake
finishes at submitted; rail status proposals through `paymentstatus.Apply` request
accepted/rejected/settled/returned/failed transitions from paymentorder.

Execution retries are bounded to ten failures; exhausted work remains in
`paymentexecution.processing_failures` with `status = dead` and is logged for
restricted investigation. After resolving the cause, an authorized in-process
operator can call `PaymentExecution.Process(payment_id)` to replay the same durable
references. Beneficiary notifications have their own durable inbox/delivery queue,
bounded retries and visible dead records; delivery failure never changes payment
state. Payment, rail and authentication request/response bodies are excluded from
HTTP logs, and endpoint body limits run before body reads.

Real adapters are configured by the entries in `services/payment/example.env`:
core banking, screening, rail connectors, private audit S3 credentials/bucket and
notification delivery. The deployment must preprovision that private bucket with
its required retention policy. Core banking provides GET `/accounts/{id}`, POST
`/balance/check`, and POST `/post`; posting requires X-Client-ID/idempotency references
and returns a durable POSTED result. No stub is a production fallback.

Verification lives in `services/payment/payment_flows_integration_test.go`, the
aggregate/XSD tests, adapter tests, and generated-error-model tests. The PostgreSQL
suite covers TS-ORD-01/02 and TS-ISO-01/02/03, authenticated HTTP routes, concurrent
idempotency, rollback on outbox failure, cover arrival order/mismatch/expiry, single
credit on replay, and stable rail references after a partial cover failure. CI runs
`GOWORK=off go test -race ./...` with `PAYMENT_TEST_POSTGRES_URL` pointing to its
dedicated PostgreSQL service. Optional broker integration suites retain their own
environment settings. Five approved native XSD roots are compiled at startup;
benchmark methodology and results are recorded beside the schema catalogue.
Execution session locks use a separate bounded connection pool so concurrent
workers cannot exhaust the write pool while holding locks; this is tested with
a one-connection write pool.
