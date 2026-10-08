# 10 — Status Reports & ACK Ingestion

## 1. Overview & Objective
`paymentstatus` ingests asynchronous rail feedback (ACK/NACK, camt.054, pacs.002). It correlates reports with `payment_id`, deduplicates report payloads, and proposes legal state transitions to `paymentorder`.

## 2. Domain Model (DDD & CQRS Architecture)

### 2.1 Aggregates & Value Objects
- **Aggregate**: `StatusReport`
- **Value Objects**: `ExternalStatusCode` (e.g. `ACCP`, `RJCT`, `PDNG`), `ReportCorrelation`

### 2.2 Commands & Domain Events
- **Command**: `IngestStatusReportCommand`
- **Domain Event**: `StatusReportProcessedEvent`

## 3. Database Schema (`adapters/db/migration/0001_create_status_reports.up.sql`)
```sql
CREATE SCHEMA IF NOT EXISTS paymentstatus;

CREATE TABLE paymentstatus.status_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID NOT NULL,
    execution_id UUID,
    external_status_code VARCHAR(32) NOT NULL,
    raw_report_s3_ref VARCHAR(255),
    processed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## 4. Webhook Intake API
- `POST /api/v1/status-reports/ingest`
