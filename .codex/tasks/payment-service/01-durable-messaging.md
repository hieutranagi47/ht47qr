# 01 — Durable Outbox, Inbox, and Kafka Conventions

## 1. Overview & Objective

Establishes reliable event-driven communication between modules using the Transactional Outbox Pattern and Inbox Pattern over Apache Kafka. Guarantees at-least-once delivery, idempotent event processing, and durable transaction integrity across restarts.

## 2. Transactional Outbox & Inbox Schemas (Database Migrations)

### SQL Migration: `0001_create_outbox_inbox.up.sql`

```sql
-- Outbox Table for module schema
CREATE TABLE IF NOT EXISTS paymentorder.payment_events_to_forward (
    transaction_id xid8 NOT NULL,
    offset BIGINT NOT NULL,
    uuid VARCHAR(36) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    payload JSONB NOT NULL,
    metadata JSONB NOT NULL,
    PRIMARY KEY (transaction_id, offset)
);

CREATE TABLE IF NOT EXISTS paymentorder.payment_offsets_events_to_forward (
    consumer_group VARCHAR(255) PRIMARY KEY,
    offset BIGINT NOT NULL
);

-- Inbox Table for deduplication
CREATE TABLE IF NOT EXISTS paymentorder.payment_processed_events (
    event_id UUID PRIMARY KEY,
    consumer_group VARCHAR(128) NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## 3. Event Envelope Specification

All Kafka messages publish using a standard JSON envelope:

```json
{
  "event_id": "c3b9b4f4-5b6c-4f7f-8d1e-2a3b4c5d6e7f",
  "event_type": "payment.order.received.v1",
  "occurred_at": "2026-09-22T20:11:05Z",
  "payment_id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
  "correlation_id": "e7b9b4f4-5b6c-4f7f-8d1e-2a3b4c5d6e7f",
  "causation_id": "f8b9b4f4-5b6c-4f7f-8d1e-2a3b4c5d6e7f",
  "version": 1,
  "payload": {}
}
```

## 4. Processing & Resiliency Rules

1. **Partition Key**: Always set Kafka partition key to `payment_id` to preserve per-payment event order.
2. **At-Least-Once Delivery & Inbox Processing**:
   Consumers execute business logic and write `event_id` into `processed_events` in the **same local PostgreSQL transaction**. Duplicate events are ignored cleanly.
3. **Bounded Retries & DLQ**:
   - Max 3 automatic retries with exponential backoff (100ms, 500ms, 2000ms).
   - Unhandled/Poison messages move to Dead Letter Queue topic (`<topic>.dlq`) with full error headers for operational replay.
