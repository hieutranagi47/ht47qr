# 14 — Notifications & Webhook Callbacks

## 1. Overview & Objective
`notifications` dispatches Push Notifications, SMS alerts to retail customers, and signed Webhook Callbacks to corporate clients upon terminal payment state transitions. It operates asynchronously and guarantees that notification delivery failures NEVER halt or reverse payment completion.

## 2. Domain Model (DDD & CQRS Architecture)

### 2.1 Aggregates & Value Objects
- **Aggregate**: `NotificationLog`
- **Value Objects**:
  - `Channel` (`PUSH`, `SMS`, `WEBHOOK`)
  - `DeliveryStatus` (`PENDING`, `DELIVERED`, `FAILED`, `DLQ`)
  - `Signature` (HMAC-SHA256)

### 2.2 Commands & Events
- **Command**: `DispatchNotificationCommand`
- **Domain Events**:
  - `NotificationDeliveredEvent`
  - `NotificationFailedEvent`

## 3. Database Schema (`adapters/db/migration/0001_create_notification_logs.up.sql`)
```sql
CREATE SCHEMA IF NOT EXISTS notifications;

CREATE TABLE notifications.notification_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID NOT NULL,
    channel VARCHAR(16) NOT NULL,
    recipient VARCHAR(128) NOT NULL,
    payload_version VARCHAR(16) NOT NULL,
    status VARCHAR(16) NOT NULL,
    attempt_count INT DEFAULT 0,
    signature VARCHAR(128),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## 4. Webhook Security Specification
All corporate webhooks include HTTP headers:
- `X-Payment-Signature`: `t=timestamp,v1=HMAC_SHA256(secret, payload)`
