# 15 — Capacity, Resilience, and Production Readiness

## 1. Overview & Objective

Defines the infrastructure, scalability, and operational benchmarks required to process **2,500 – 4,000 TPS peak capacity** (25M – 35M daily payments) while ensuring 99.99% system availability.

## 2. Database Scalability & Partitioning Policy

- **Time Partitioning**: High-volume append-only tables (`paymentorder.payment_audit_logs`, `paymentorder.watermill_events_to_forward`, `paymentstatus.status_reports`) use monthly PostgreSQL range partitioning.
- **Sharding Strategy**: Horizontal sharding by `debtor_account` (Account_ID) when primary cluster write limits are reached.

### SQL Migration: `0002_partition_audit_logs.up.sql`

```sql
-- Partitioning audit log by month
CREATE TABLE paymentorder.payment_audit_logs_partitioned (
    id UUID NOT NULL,
    payment_id UUID NOT NULL,
    from_state VARCHAR(32),
    to_state VARCHAR(32) NOT NULL,
    reason TEXT,
    actor_id VARCHAR(64),
    causation_id UUID,
    correlation_id UUID,
    created_at TIMESTAMPTZ NOT NULL
) PARTITION BY RANGE (created_at);

CREATE TABLE paymentorder.payment_audit_logs_y2026m09 PARTITION OF paymentorder.payment_audit_logs_partitioned
    FOR VALUES FROM ('2026-09-01 00:00:00+00') TO ('2026-10-01 00:00:00+00');
```

## 3. Kafka Queue Segregation & Worker Allocation

- **Retail Real-time Stream**: Topic `payment.retail.instant.v1` (Allocated 70% Core Banking Connection Capacity).
- **Corporate Bulk Stream**: Topic `payment.corporate.bulk.v1` (Allocated 30% Capacity with Token-bucket throttling).

## 4. Operational Runbooks & Observability

- **Prometheus Metrics**: `payment_orders_total{state}`, `kafka_consumer_lag`, `core_banking_latency_seconds`.
- **Grafana Dashboard**: System throughput TPS, P99 latency, Outbox age, DLQ queue depth.
- **Disaster Recovery (DR)**: RPO = 0 (No lost committed payments), RTO < 15 minutes via Automated HA PostgreSQL Failover and Kafka Replay.
