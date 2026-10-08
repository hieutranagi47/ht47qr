# 02_final — ISO 20022 Schema & Semantic Validation, Identity, OTP, and Security Audit Specification

## 1. Executive Summary & Subsystem Architecture
This consolidated specification defines the **Identity**, **OTP (One-Time Password)**, **Security Audit & Anomaly Detection**, and **ISO 20022 Message Validation** subsystems.

It integrates:
1. **Device Public Key Registration & Biometric Cryptographic Verification** (`identity` module).
2. **RFC 6238 TOTP (Smart OTP) & CSPRNG Challenge Key Management** (`otp` module).
3. **Bot Velocity Anomaly Detection & 5-Strike Temporary Account Lock** (`securityaudit` module).
4. **ISO 20022 Phase 1 XSD Schema Parsing & Phase 2 Advanced Semantic Business Validation** (`iso20022` & `paymentvalidation` modules).

---

## 2. Identity, OTP, and Biometric Security Subsystems

### 2.1 `otp` Module
* **TOTP Algorithm**: Implements RFC 6238 TOTP using HMAC-SHA1 (based on `generating-otp.go`):
  1. Base32 decode shared secret $K$.
  2. Compute time step $T = \text{floor}(\text{UnixTime} / 30)$.
  3. Dynamic Truncation converts HMAC-SHA1 hash into a 6-digit numeric OTP code.
  4. Accepts clock drift of $\pm 1$ time window (30 seconds).
  5. Single-use enforcement: verified codes are instantly invalidated in Redis/DB.
* **Challenge Key Lifecycle**: CSPRNG 256-bit random hex string generated upon transaction initiation (`POST /initiate`), stored in Redis `challenge:{Trans_ID}` with TTL = 120s, consumed and deleted atomically during execution (`POST /execute`).

### 2.2 `identity` Module
* **Device Key Registration**: `POST /api/v1/identity/public-key` accepts PEM-encoded Hardware Public Keys (ECDSA secp256r1 / Ed25519) generated inside mobile TEE/Secure Enclave.
* **Signature Verification**: Verifies digital signature over `[Challenge_Key + CanonicalJSON(Payload)]` using stored user device Public Key.

---

## 3. Security Audit, Anomaly Detection & Account Lock (`securityaudit`)

### 3.1 Impossible Human Speed Detection
* **Rule**: Measures interval $\Delta t = t_{execute} - t_{challenge}$.
* **Threshold**: Human reading and submitting OTP/Biometrics takes 5–30 seconds. If $\Delta t < 5.0\text{ seconds}$, flag as `AUTOMATED_BOT_SPEED_ANOMALY`.

### 3.2 5-Strike Lockout & 10-Minute Temporary Lock
* **Sliding Window**: Tracks failed authentication/signature attempts in Redis `failed_auth_count:{user_id}`.
* **Lock Trigger**: Upon hitting 5 failures within 15 minutes, set Redis key `account_temp_lock:{user_id}` with TTL = 600s (10 minutes).
* **HTTP Response**: Any new `/initiate` or `/execute` request during lock period returns `HTTP 423 Locked` (`ACCOUNT_TEMPORARILY_LOCKED`).
* **Multi-Channel Alerting**: Emits `AccountTemporarilyLockedEvent` to Kafka; module `14 notifications` sends Push Notification, SMS, and Email with 1-click permanent lock option.
* **Automatic Unlock**: Redis key expires after 10 minutes, restoring account payment access.

---

## 4. ISO 20022 Message Validation Engine

Validation occurs in two distinct phases:

```text
[Incoming XML] ──► Phase 1: XSD Schema Validation ──► Phase 2: Semantic Business Rules ──► [Valid DTO]
                         │ (Malformed / Bad XML)             │ (Business Rule Failure)
                         ▼                                   ▼
                   HTTP 400 Bad Request               pacs.002.001.10 XML (RJCT)
```

### 4.1 Phase 1: XSD Schema & XXE Protection (`iso20022`)
* Validates XML against native XSD schema catalog (`iso20022.schema_catalog`).
* Rejects messages > 10MB or containing XML External Entity (XXE) references.

### 4.2 Phase 2: 8 Semantic Business Validation Rules (`paymentvalidation`)
1. **External Code Sets Lookup**: Validate `ExternalCategoryPurpose1Code` (`SALA`, `SUPP`, `CASH`), `ExternalPurpose1Code`, and `ExternalServiceLevel1Code` against cached ISO code sets. *ISO Reason Code: `FF01`*.
2. **ISO 4217 Currency Precision**: Enforce decimal places (`VND`, `JPY` = 0; `USD`, `EUR` = max 2; `BHD`, `KWD` = max 3). *ISO Reason Code: `AM09`*.
3. **Execution Date & Cut-Off Controls**: Reject past dates, dates > 30 days future, or cut-off time violations. *ISO Reason Code: `DT01` / `TM01`*.
4. **Account & IBAN Checksum**: Modulo 97 IBAN checksum (ISO 13616) & local BBAN regex. *ISO Reason Code: `AC01`*.
5. **BIC / SWIFT Directory Lookup**: Validate BIC format (ISO 9362) and presence in Bank Directory. *ISO Reason Code: `RC01`*.
6. **Intra-Batch Deduplication**: Check uniqueness of `EndToEndId` and `UETR` across all `CdtTrfTxInf` elements within the same XML batch payload. *ISO Reason Code: `DUPL`*.
7. **Charge Bearer Rule Compliance**: Instant payment rails (e.g., NAPAS 24/7) must only accept `SHAR` or `SLEV` (reject unsupported `OUR`/`BEN`). *ISO Reason Code: `CHG1`*.
8. **Single Transaction Rail Limits**: Amount must not exceed rail single limit (e.g., NAPAS <= 500M VND). *ISO Reason Code: `AM02`*.

---

## 5. ISO 20022 Status Report Builder (`pacs.002.001.10`)
When validation fails, system constructs a structured `pacs.002.001.10` XML rejection payload:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<Document xmlns="urn:iso:std:iso:20022:tech:xsd:pacs.002.001.10">
  <FIToFIPmtStsRpt>
    <GrpHdr>
      <MsgId>RPT20260927-000123</MsgId>
      <CreDtTm>2026-09-27T07:45:00Z</CreDtTm>
    </GrpHdr>
    <TxInfAndSts>
      <OrgnlEndToEndId>E2E-20260927-8899</OrgnlEndToEndId>
      <TxSts>RJCT</TxSts>
      <StsRsnInf>
        <Rsn><Cd>AM09</Cd></Rsn>
        <AddtlInf>Invalid currency precision for VND: fractional amounts not allowed</AddtlInf>
      </StsRsnInf>
    </TxInfAndSts>
  </FIToFIPmtStsRpt>
</Document>
```

---

## 6. Database Schemas

### 6.1 `identity.user_public_keys`
```sql
CREATE TABLE identity.user_public_keys (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id VARCHAR NOT NULL,
  device_id VARCHAR NOT NULL,
  public_key_pem TEXT NOT NULL,
  algorithm VARCHAR NOT NULL DEFAULT 'ECDSA_P256',
  is_revoked BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  updated_at TIMESTAMP NOT NULL DEFAULT now(),
  deleted_at TIMESTAMP,
  CONSTRAINT uq_user_device UNIQUE (user_id, device_id)
);
```

### 6.2 `otp.user_secrets`
```sql
CREATE TABLE otp.user_secrets (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id VARCHAR NOT NULL UNIQUE,
  encrypted_secret_key VARCHAR NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  updated_at TIMESTAMP NOT NULL DEFAULT now()
);
```

### 6.3 `security.account_locks`
```sql
CREATE TABLE security.account_locks (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id VARCHAR NOT NULL,
  lock_reason VARCHAR NOT NULL,
  failed_attempts_count INT NOT NULL,
  locked_at TIMESTAMP NOT NULL DEFAULT now(),
  expires_at TIMESTAMP NOT NULL,
  unlocked_at TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT now()
);
```

---

## 7. Implementation Checklist & Test Matrix for AI Agent
- [ ] Implement `identity` public key registration and ECDSA/Ed25519 signature verification.
- [ ] Implement `otp` TOTP RFC 6238 generation/verification and Redis Challenge Key manager.
- [ ] Implement `securityaudit` velocity check ($\Delta t < 5s$) and 5-strike Redis lockout (`HTTP 423`).
- [ ] Implement Phase 1 XSD validator and Phase 2 Semantic Business Validator pipeline (Rules 1-8).
- [ ] Implement `pacs.002.001.10` XML generator.
- [ ] Write test cases: `TS-SEC-01` to `TS-SEC-06` and `TS-VAL-XML-01` to `TS-VAL-XML-08`.
