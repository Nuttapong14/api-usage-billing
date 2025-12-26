# Product Requirements Document (PRD)
# API Usage Analytics & Billing Platform

**Document Version:** 1.0  
**Created:** December 2024  
**Author:** Nuttapong  
**Status:** Draft

---

## Executive Summary

แพลตฟอร์มสำหรับติดตามการใช้งาน API แบบ real-time พร้อมระบบคิดค่าบริการอัตโนมัติ รองรับ multi-tenant architecture สำหรับผู้ให้บริการ API ที่ต้องการ monetize APIs ของตน

---

## 1. Problem Statement

### 1.1 ปัญหาที่ต้องการแก้ไข

| ปัญหา | ผลกระทบ |
|-------|---------|
| ไม่มีระบบ track การใช้งาน API แบบ real-time | ไม่สามารถ monitor และ optimize ได้ |
| คิดค่าบริการ manual ทุกเดือน | เสียเวลา, error-prone, delayed billing |
| ไม่มี visibility ให้ลูกค้าเห็น usage | ลูกค้าไม่พอใจ, support ticket เยอะ |
| ไม่สามารถ enforce quota/rate limit ได้แม่นยำ | Revenue leakage, unfair usage |
| ไม่มี analytics สำหรับ business decision | ไม่รู้ว่า API ไหน popular, ลูกค้าไหนมี potential |

### 1.2 Target Users

1. **API Providers** - บริษัทที่ต้องการขาย API (Fintech, Data providers, SaaS)
2. **API Consumers** - ลูกค้าที่ซื้อ API มาใช้งาน
3. **Platform Admins** - ทีมดูแลระบบ
4. **Finance Team** - ทีมการเงินที่ต้องการ billing reports

---

## 2. Product Vision & Goals

### 2.1 Vision Statement

> "เป็น one-stop platform สำหรับ API monetization ที่ทำให้ผู้ให้บริการ API สามารถ track, analyze และ bill ได้อัตโนมัติ 100%"

### 2.2 Success Metrics (KPIs)

| Metric | Target | Measurement |
|--------|--------|-------------|
| Billing Accuracy | 99.99% | Discrepancy rate |
| Real-time Latency | < 100ms | P99 latency for usage update |
| Invoice Generation Time | < 5 min | End-of-month processing |
| Customer Self-Service Rate | > 80% | % issues resolved without support |
| System Uptime | 99.9% | Monthly availability |

---

## 3. Features & Requirements

### 3.1 Core Features

#### F1: Real-time Usage Tracking

**Priority:** P0 (Must Have)

```
┌─────────────────────────────────────────────────────────────────┐
│                     Usage Tracking Flow                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   API Request                                                   │
│       │                                                         │
│       ▼                                                         │
│   ┌───────────┐    ┌──────────────┐    ┌───────────────────┐   │
│   │   Kong    │───▶│ Kong Plugin  │───▶│  Redis Stream     │   │
│   │  Gateway  │    │ (Log/Track)  │    │  (Raw Events)     │   │
│   └───────────┘    └──────────────┘    └─────────┬─────────┘   │
│                                                  │              │
│                                                  ▼              │
│   ┌───────────────────┐    ┌──────────────────────────────┐    │
│   │  Usage Aggregator │◀───│  Stream Consumer (Go Worker) │    │
│   │     Service       │    └──────────────────────────────┘    │
│   └─────────┬─────────┘                                        │
│             │                                                   │
│             ▼                                                   │
│   ┌─────────────────┐    ┌─────────────────┐                   │
│   │  Redis Counter  │    │    Postgres     │                   │
│   │  (Real-time)    │    │  (Persistent)   │                   │
│   └─────────────────┘    └─────────────────┘                   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Requirements:**

| ID | Requirement | Acceptance Criteria |
|----|-------------|---------------------|
| F1.1 | Track ทุก API request ผ่าน Kong | 100% requests logged |
| F1.2 | บันทึก metrics: latency, status code, payload size | ทุก field ถูกบันทึก |
| F1.3 | Real-time counter update | Update ภายใน 1 วินาที |
| F1.4 | Support custom metering (per request, per MB, per compute time) | Config ได้ต่อ API |
| F1.5 | Handle burst traffic 10,000 RPS | No data loss at peak |

---

#### F2: Subscription & Quota Management

**Priority:** P0 (Must Have)

```
┌─────────────────────────────────────────────────────────────────┐
│                    Subscription Tiers                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐          │
│   │    FREE     │   │    PRO      │   │ ENTERPRISE  │          │
│   ├─────────────┤   ├─────────────┤   ├─────────────┤          │
│   │ 1,000 req/  │   │ 100,000 req/│   │ Unlimited   │          │
│   │ month       │   │ month       │   │             │          │
│   │             │   │             │   │             │          │
│   │ 10 req/sec  │   │ 100 req/sec │   │ 1000 req/sec│          │
│   │             │   │             │   │             │          │
│   │ No SLA      │   │ 99.5% SLA   │   │ 99.9% SLA   │          │
│   │             │   │             │   │             │          │
│   │ ฿0/month    │   │ ฿2,999/month│   │ Custom      │          │
│   └─────────────┘   └─────────────┘   └─────────────┘          │
│                                                                 │
│   Overage: ฿0.01 per request after quota exhausted              │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Requirements:**

| ID | Requirement | Acceptance Criteria |
|----|-------------|---------------------|
| F2.1 | สร้าง subscription tier ได้ไม่จำกัด | Admin UI + API |
| F2.2 | กำหนด quota per tier (requests, bandwidth, compute) | Flexible config |
| F2.3 | Rate limiting ตาม tier | Kong enforced |
| F2.4 | Soft limit warning (80%, 90%, 100%) | Email + Webhook |
| F2.5 | Hard limit action (block, throttle, overage billing) | Configurable |
| F2.6 | Upgrade/Downgrade tier mid-cycle | Pro-rated billing |

---

#### F3: Billing & Invoice Automation

**Priority:** P0 (Must Have)

```
┌─────────────────────────────────────────────────────────────────┐
│                    Billing Cycle Flow                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Day 1-30: Usage Accumulation                                   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  Real-time Counters ──▶ Daily Aggregation ──▶ Monthly   │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  Day 1 (Next Month): Invoice Generation                        │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                                                         │   │
│  │  ┌──────────┐    ┌──────────┐    ┌──────────┐          │   │
│  │  │ Collect  │───▶│ Calculate│───▶│ Generate │          │   │
│  │  │ Usage    │    │ Charges  │    │ Invoice  │          │   │
│  │  └──────────┘    └──────────┘    └────┬─────┘          │   │
│  │                                       │                 │   │
│  │                    ┌──────────────────┼────────────┐    │   │
│  │                    ▼                  ▼            ▼    │   │
│  │              ┌──────────┐      ┌──────────┐  ┌───────┐  │   │
│  │              │   PDF    │      │  Email   │  │Webhook│  │   │
│  │              │ Storage  │      │  Send    │  │ Notify│  │   │
│  │              └──────────┘      └──────────┘  └───────┘  │   │
│  │                                                         │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  Day 7: Payment Reminder                                        │
│  Day 15: Payment Due                                            │
│  Day 30: Service Suspension (if unpaid)                         │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Requirements:**

| ID | Requirement | Acceptance Criteria |
|----|-------------|---------------------|
| F3.1 | Auto-generate invoice ทุกสิ้นเดือน | K8S CronJob |
| F3.2 | Support billing models: Fixed, Usage-based, Hybrid | Config per tier |
| F3.3 | Generate PDF invoice | Thai + English |
| F3.4 | Tax calculation (VAT 7%) | ถูกต้องตามกฎหมาย |
| F3.5 | Credit/Debit notes สำหรับ adjustment | Audit trail |
| F3.6 | Multi-currency support | THB, USD, SGD |
| F3.7 | Payment gateway integration | Omise, Stripe, Bank Transfer |

---

#### F4: Analytics Dashboard

**Priority:** P1 (Should Have)

```
┌─────────────────────────────────────────────────────────────────┐
│                    Dashboard Views                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                 PROVIDER DASHBOARD                       │   │
│  ├─────────────────────────────────────────────────────────┤   │
│  │  📊 Total Revenue (MTD)        📈 API Calls (Today)     │   │
│  │     ฿1,234,567                    12.5M                 │   │
│  │                                                         │   │
│  │  📉 Usage Trend (30 days)                               │   │
│  │  ▁▂▃▅▆▇█▇▆▅▄▃▂▁▂▃▄▅▆▇█▇▆▅▄▃▂▁▂▃                        │   │
│  │                                                         │   │
│  │  🏆 Top APIs              👥 Top Customers              │   │
│  │  1. /v1/payments  45%     1. TechCorp    ฿450K         │   │
│  │  2. /v1/users     30%     2. FinanceApp  ฿320K         │   │
│  │  3. /v1/reports   25%     3. StartupXYZ  ฿180K         │   │
│  │                                                         │   │
│  │  ⚠️ Alerts                                              │   │
│  │  • 3 customers approaching quota                        │   │
│  │  • 2 invoices overdue                                   │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                 CUSTOMER DASHBOARD                       │   │
│  ├─────────────────────────────────────────────────────────┤   │
│  │  📊 Current Usage          📈 Quota Remaining           │   │
│  │     45,230 / 100,000          54.77%                    │   │
│  │     ████████████░░░░░░░░                                │   │
│  │                                                         │   │
│  │  💰 Current Bill (MTD)     📅 Next Invoice              │   │
│  │     ฿2,999 + ฿0 overage       Jan 1, 2025              │   │
│  │                                                         │   │
│  │  📉 My Usage (7 days)                                   │   │
│  │  Mon  Tue  Wed  Thu  Fri  Sat  Sun                      │   │
│  │   █    █    █    █    █    ▄    ▂                       │   │
│  │  8.2K 7.8K 9.1K 8.5K 8.0K 3.2K 1.1K                    │   │
│  │                                                         │   │
│  │  🔑 API Keys               📄 Invoices                  │   │
│  │  • Production: ****a1b2    • Dec 2024: ฿2,999 ✓ Paid   │   │
│  │  • Staging: ****c3d4       • Nov 2024: ฿2,999 ✓ Paid   │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Requirements:**

| ID | Requirement | Acceptance Criteria |
|----|-------------|---------------------|
| F4.1 | Real-time usage dashboard | < 5 sec refresh |
| F4.2 | Historical analytics (daily/weekly/monthly) | 12 months retention |
| F4.3 | API-level breakdown | Per endpoint stats |
| F4.4 | Error rate monitoring | By status code |
| F4.5 | Latency percentiles (P50, P95, P99) | Per API |
| F4.6 | Export reports (CSV, PDF) | Scheduled + On-demand |
| F4.7 | Custom date range filter | Any range |

---

#### F5: Alerting & Notifications

**Priority:** P1 (Should Have)

**Alert Types:**

| Alert | Trigger | Channels | Action |
|-------|---------|----------|--------|
| Quota Warning | 80% usage | Email, Webhook | Notify |
| Quota Critical | 95% usage | Email, SMS, Webhook | Notify + Throttle |
| Quota Exceeded | 100% usage | Email, SMS, Webhook | Block or Overage |
| Payment Due | 7 days before | Email | Reminder |
| Payment Overdue | Due date passed | Email, SMS | Warning |
| Service Suspension | 30 days overdue | Email, SMS | Suspend access |
| Anomaly Detection | Unusual spike/drop | Email, Webhook | Alert for review |
| Error Rate High | > 5% errors | Email, Webhook | Alert ops team |

---

#### F6: API Key Management

**Priority:** P0 (Must Have)

**Requirements:**

| ID | Requirement | Acceptance Criteria |
|----|-------------|---------------------|
| F6.1 | Generate API keys (multiple per customer) | Unlimited keys |
| F6.2 | Key rotation without downtime | Grace period config |
| F6.3 | Key-level permissions (read/write/admin) | RBAC |
| F6.4 | Key expiration policy | Auto-expire option |
| F6.5 | IP whitelist per key | CIDR support |
| F6.6 | Usage tracking per key | Separate counters |
| F6.7 | Revoke key instantly | Real-time enforcement |

---

### 3.2 Non-Functional Requirements

#### NFR1: Performance

| Metric | Requirement |
|--------|-------------|
| API Gateway Latency | < 10ms added latency |
| Usage Query Response | < 200ms for dashboard |
| Concurrent Users | 10,000 dashboard users |
| Event Processing | 50,000 events/second |
| Invoice Generation | < 5 minutes for 10,000 customers |

#### NFR2: Scalability

| Component | Scaling Strategy |
|-----------|------------------|
| Kong Gateway | Horizontal (K8S HPA) |
| Usage Workers | Horizontal (based on queue depth) |
| Redis | Cluster mode (6 nodes min) |
| Postgres | Read replicas + Connection pooling |
| Dashboard | CDN + Static assets |

#### NFR3: Reliability

| Metric | Target |
|--------|--------|
| Uptime | 99.9% (8.76 hours downtime/year) |
| Data Durability | 99.999999999% (11 nines) |
| RTO (Recovery Time Objective) | < 1 hour |
| RPO (Recovery Point Objective) | < 5 minutes |

#### NFR4: Security

| Requirement | Implementation |
|-------------|----------------|
| Authentication | Keycloak (OIDC) |
| API Key Encryption | AES-256 at rest |
| Data Encryption | TLS 1.3 in transit |
| Audit Logging | All admin actions |
| GDPR Compliance | Data export/deletion |
| PCI-DSS | If handling payments |

---

## 4. Security Architecture

### 4.1 API Key Security Model

ความปลอดภัยของ API Key เป็นหัวใจของระบบ เราใช้หลักการ **"Never Store Plain Key"** เพื่อสร้างความเชื่อมั่นให้ลูกค้า

#### 4.1.1 Key Generation & Storage

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    API Key Security Model                                │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│   เมื่อสร้าง Key ใหม่:                                                   │
│                                                                         │
│   ┌──────────────────┐      ┌──────────────────────────────────────────┐│
│   │  Generate Key    │      │   apik_example_a1b2c3d4e5f6g7h8i9j0   ││
│   │  (Random 32B)    │─────▶│   ─────────────────────────────────────  ││
│   └──────────────────┘      │   แสดงให้ลูกค้า ครั้งเดียว เท่านั้น!      ││
│                             └──────────────────────────────────────────┘│
│                                              │                          │
│                                              ▼                          │
│   ┌──────────────────────────────────────────────────────────────────┐ │
│   │                      สิ่งที่เราเก็บใน Database                    │ │
│   ├──────────────────────────────────────────────────────────────────┤ │
│   │                                                                  │ │
│   │   key_prefix: "apik_live_a1"     ← สำหรับแสดง UI (8 ตัวแรก)      │ │
│   │                                                                  │ │
│   │   key_hash: "8f14e45f..."        ← SHA-256 hash (ไม่ reverse ได้)│ │
│   │                                                                  │ │
│   │   customer_id: "uuid..."         ← เจ้าของ key                   │ │
│   │                                                                  │ │
│   └──────────────────────────────────────────────────────────────────┘ │
│                                                                         │
│   ⚠️  ถ้า Database ถูก hack → Hacker ได้แค่ hash → ใช้งานไม่ได้!        │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

**Key Format Specification:**

| Component | Format | Example |
|-----------|--------|---------|
| Prefix | `apik_` + environment | `apik_live_`, `apik_test_` |
| Random Part | 32 bytes, base62 encoded | `a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6` |
| Full Key | prefix + random | `apik_example_a1b2c3d4e5f6g7h8i9j0o5p6` |
| Stored Hash | SHA-256 | `8f14e45fceea167a5a36dedd4bea2543...` |
| Display Prefix | First 12 chars | `apik_live_a1****` |

#### 4.1.2 Key Authentication Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    Request Authentication Flow                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│   Customer Request                                                      │
│   ┌─────────────────────────────────────────────────────────────────┐  │
│   │ GET /api/v1/payments                                            │  │
│   │ Authorization: Bearer apik_example_a1b2c3d4e5f6g7h8i9j0      │  │
│   └───────────────────────────────┬─────────────────────────────────┘  │
│                                   │                                     │
│                                   ▼                                     │
│   ┌─────────────────────────────────────────────────────────────────┐  │
│   │                         KONG GATEWAY                             │  │
│   │                                                                  │  │
│   │   Step 1: Extract Key จาก Header                                │  │
│   │                     │                                            │  │
│   │                     ▼                                            │  │
│   │   Step 2: Hash Key ด้วย SHA-256                                  │  │
│   │           SHA256("apik_live_a1...") = "8f14e45f..."              │  │
│   │                     │                                            │  │
│   │                     ▼                                            │  │
│   │   Step 3: Lookup hash ใน Redis Cache (fast path)                 │  │
│   │           ┌─────────────────────────────────────────────┐       │  │
│   │           │ Redis: key_cache:8f14e45f → {customer_id,   │       │  │
│   │           │        tier, rate_limit, permissions, ip}   │       │  │
│   │           └─────────────────────────────────────────────┘       │  │
│   │                     │                                            │  │
│   │                     ▼                                            │  │
│   │   Step 4: Cache miss? → Query Postgres → Cache result            │  │
│   │                     │                                            │  │
│   │                     ▼                                            │  │
│   │   Step 5: Validate                                               │  │
│   │           • is_active = true?                                    │  │
│   │           • expires_at > now()?                                  │  │
│   │           • request_ip IN ip_whitelist?                          │  │
│   │           • has required permissions?                            │  │
│   │                     │                                            │  │
│   │           ┌─────────┴─────────┐                                  │  │
│   │           ▼                   ▼                                  │  │
│   │       ✅ PASS              ❌ REJECT                             │  │
│   │     Forward to            Return 401/403                         │  │
│   │     upstream              with error code                        │  │
│   │                                                                  │  │
│   └─────────────────────────────────────────────────────────────────┘  │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

#### 4.1.3 API Key Lifecycle Management

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         API Key Lifecycle                                │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│   1️⃣ CREATE                                                             │
│   ┌─────────────────────────────────────────────────────────────────┐  │
│   │   Customer Portal: "Create Production Key"                      │  │
│   │                      │                                          │  │
│   │                      ▼                                          │  │
│   │   System:  • Generate cryptographically random 32 bytes         │  │
│   │            • Encode as base62: apik_live_a1c3d4...              │  │
│   │            • Store in DB: prefix + SHA256(full_key)             │  │
│   │            • Set default permissions & no expiration            │  │
│   │                      │                                          │  │
│   │                      ▼                                          │  │
│   │   Response: Show key ONCE with warning:                         │  │
│   │            "⚠️ Copy now! This key won't be shown again."        │  │
│   │                                                                 │  │
│   │   Audit Log: KEY_CREATED by user X at timestamp                 │  │
│   └─────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│   2️⃣ USE                                                                │
│   ┌─────────────────────────────────────────────────────────────────┐  │
│   │   Every API Request:                                            │  │
│   │   • Hash key → lookup → verify                                  │  │
│   │   • Check: active, not expired, IP allowed, has permission      │  │
│   │   • Enforce rate limit based on tier                            │  │
│   │   • Enforce quota based on subscription                         │  │
│   │   • Log usage event with api_key_id                             │  │
│   │   • Update last_used_at timestamp                               │  │
│   └─────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│   3️⃣ ROTATE (Zero-Downtime)                                            │
│   ┌─────────────────────────────────────────────────────────────────┐  │
│   │   Customer: "Rotate this key"                                   │  │
│   │                      │                                          │  │
│   │                      ▼                                          │  │
│   │   System:  • Generate NEW key (shown once)                      │  │
│   │            • OLD key marked: expires_at = now + 24h             │  │
│   │            • Both keys work during grace period                 │  │
│   │            • Customer deploys new key to their systems          │  │
│   │            • Old key auto-expires after grace period            │  │
│   │                                                                 │  │
│   │   Timeline:                                                     │  │
│   │   ├─────────────────┼─────────────────┼─────────────────────┤  │  │
│   │   T+0              T+12h              T+24h                     │  │
│   │   Rotate           Warning            Old key                   │  │
│   │   requested        email sent         disabled                  │  │
│   │                                                                 │  │
│   │   ✅ Zero downtime! Customer has 24h to update their apps      │  │
│   │                                                                 │  │
│   │   Audit Log: KEY_ROTATED old_key_id → new_key_id               │  │
│   └─────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│   4️⃣ REVOKE (Instant)                                                  │
│   ┌─────────────────────────────────────────────────────────────────┐  │
│   │   Trigger: Customer clicks "Delete" OR "Key compromised!"      │  │
│   │                      │                                          │  │
│   │                      ▼                                          │  │
│   │   System:  • Set is_active = false in Postgres                  │  │
│   │            • DELETE from Redis cache immediately                │  │
│   │            • Publish invalidation event to all Kong nodes       │  │
│   │                      │                                          │  │
│   │                      ▼                                          │  │
│   │   Effect:  Next request with this key → 401 Unauthorized        │  │
│   │            Latency: < 1 second globally                         │  │
│   │                                                                 │  │
│   │   ⚡ Instant effect across all nodes!                           │  │
│   │                                                                 │  │
│   │   Audit Log: KEY_REVOKED reason: user_initiated | compromised  │  │
│   └─────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│   5️⃣ EXPIRE (Automatic)                                                │
│   ┌─────────────────────────────────────────────────────────────────┐  │
│   │   For keys with expiration date set:                            │  │
│   │                                                                 │  │
│   │   • 7 days before: Email reminder to rotate                     │  │
│   │   • 1 day before: Final warning email                           │  │
│   │   • Expiry time: Key stops working automatically                │  │
│   │   • Webhook sent: key.expired event                             │  │
│   │                                                                 │  │
│   │   Audit Log: KEY_EXPIRED auto-expiry                            │  │
│   └─────────────────────────────────────────────────────────────────┘  │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

#### 4.1.4 Security Features Matrix

| Feature | Implementation | Purpose |
|---------|----------------|---------|
| **Never Store Plain Key** | Store only SHA-256 hash | DB breach doesn't expose keys |
| **Show Once** | Key displayed only at creation | Minimize exposure window |
| **Key Prefix** | `apik_live_` / `apik_test_` | Identify environment, prevent accidents |
| **Rotation with Grace Period** | 24h overlap for both keys | Zero-downtime key changes |
| **Instant Revoke** | Redis pub/sub invalidation | Compromised keys blocked in <1s |
| **IP Whitelist** | CIDR notation per key | Stolen key useless from other IPs |
| **Expiration Policy** | Optional TTL per key | Force periodic rotation |
| **Scoped Permissions** | `read`, `write`, `admin` | Least privilege principle |
| **Rate Limiting** | Per key, per tier | Prevent abuse |
| **Usage Tracking** | Every request logged | Audit trail |
| **Anomaly Detection** | ML-based pattern analysis | Detect compromised keys |

#### 4.1.5 Why API Key is Required for Tracking

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    Why API Key = Customer Identity                       │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│   ❌ WITHOUT API Key:                                                   │
│   ┌─────────────────────────────────────────────────────────────────┐  │
│   │                                                                 │  │
│   │   Incoming Request from IP: 203.150.xxx.xxx                    │  │
│   │                                                                 │  │
│   │   Problems:                                                     │  │
│   │   ✗ IP could be shared (NAT, corporate proxy, cloud provider)  │  │
│   │   ✗ Cannot identify which customer                             │  │
│   │   ✗ Cannot determine subscription tier                         │  │
│   │   ✗ Cannot enforce correct quota                               │  │
│   │   ✗ Cannot bill the right customer                             │  │
│   │   ✗ No accountability for API abuse                            │  │
│   │                                                                 │  │
│   └─────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│   ✅ WITH API Key:                                                      │
│   ┌─────────────────────────────────────────────────────────────────┐  │
│   │                                                                 │  │
│   │   API Key: apik_live_a1... → Hash → Lookup                     │  │
│   │                                                                 │  │
│   │   Instant Knowledge:                                            │  │
│   │   ✓ Customer: TechCorp Ltd. (customer_id: uuid-123)            │  │
│   │   ✓ Subscription: Pro Tier ($299/month)                        │  │
│   │   ✓ Monthly Quota: 100,000 requests                            │  │
│   │   ✓ Rate Limit: 100 requests/second                            │  │
│   │   ✓ Current Usage: 45,230 requests this month                  │  │
│   │   ✓ Remaining: 54,770 requests                                 │  │
│   │   ✓ Key Used: "Production API" (api_key_id: uuid-456)          │  │
│   │   ✓ Permissions: read, write                                   │  │
│   │   ✓ IP Whitelist: 203.150.xxx.0/24 ✓ Matched                   │  │
│   │                                                                 │  │
│   │   Actions Enabled:                                              │  │
│   │   → Enforce rate limit for this tier                           │  │
│   │   → Increment usage counter                                    │  │
│   │   → Check quota and alert if near limit                        │  │
│   │   → Log request for billing                                    │  │
│   │   → Apply tier-specific features                               │  │
│   │                                                                 │  │
│   └─────────────────────────────────────────────────────────────────┘  │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 4.2 Usage Tracking Pipeline (with Security Context)

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                         Secure Usage Tracking Pipeline                        │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  ┌─────────────┐                                                             │
│  │   Client    │                                                             │
│  │   Request   │                                                             │
│  └──────┬──────┘                                                             │
│         │ Authorization: Bearer apik_live_a1...                              │
│         ▼                                                                    │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │                           KONG GATEWAY                                 │  │
│  │                                                                        │  │
│  │  ┌─────────────┐   ┌─────────────┐   ┌──────────────────────────────┐ │  │
│  │  │    Auth     │──▶│    Rate     │──▶│   Usage Tracking Plugin      │ │  │
│  │  │   Plugin    │   │   Limiting  │   │                              │ │  │
│  │  │             │   │   Plugin    │   │  Captures (non-sensitive):   │ │  │
│  │  │ • Verify    │   │             │   │  • customer_id (from hash)   │ │  │
│  │  │   key hash  │   │ • Check     │   │  • api_key_id (uuid only)    │ │  │
│  │  │ • Extract   │   │   quota     │   │  • endpoint path             │ │  │
│  │  │   customer  │   │ • Enforce   │   │  • HTTP method               │ │  │
│  │  │   context   │   │   rate/sec  │   │  • status code               │ │  │
│  │  │ • Validate  │   │             │   │  • request/response size     │ │  │
│  │  │   IP/perms  │   │             │   │  • latency (ms)              │ │  │
│  │  │             │   │             │   │  • timestamp                 │ │  │
│  │  └──────┬──────┘   └─────────────┘   │                              │ │  │
│  │         │                            │  ⚠️ NEVER captures:          │ │  │
│  │         ▼                            │  • Request body              │ │  │
│  │  Inject internal headers:            │  • Response body             │ │  │
│  │  X-Customer-ID: uuid-123             │  • Authorization header      │ │  │
│  │  X-API-Key-ID: uuid-456              │  • Sensitive headers         │ │  │
│  │  X-Request-ID: uuid-789              │  • Query parameters          │ │  │
│  │  X-Tier: pro                         └──────────────┬───────────────┘ │  │
│  │                                                     │                  │  │
│  └─────────────────────────────────────────────────────┼──────────────────┘  │
│                         │                              │                      │
│                         ▼                              ▼ (async, non-blocking)│
│              ┌─────────────────────┐      ┌────────────────────────────────┐ │
│              │   Upstream API      │      │        Redis Streams           │ │
│              │   (Your Service)    │      │                                │ │
│              │                     │      │   XADD usage_events *          │ │
│              │   Has access to:    │      │     customer_id uuid-123       │ │
│              │   X-Customer-ID     │      │     api_key_id uuid-456        │ │
│              │   X-Request-ID      │      │     endpoint /v1/payments      │ │
│              │   for correlation   │      │     method POST                │ │
│              │                     │      │     status 200                 │ │
│              └─────────────────────┘      │     latency_ms 45              │ │
│                                           │     ...                        │ │
│                                           └───────────────┬────────────────┘ │
│                                                           │                   │
│         ┌─────────────────────────────────────────────────┘                   │
│         ▼                                                                     │
│  ┌───────────────────────────────────────────────────────────────────────┐   │
│  │                      USAGE AGGREGATOR WORKERS                          │   │
│  │                                                                        │   │
│  │   Consumer Group (Horizontally Scalable)                               │   │
│  │   ┌─────────────────────────────────────────────────────────────────┐ │   │
│  │   │  Worker 1 ─┬─▶ Process events in parallel                       │ │   │
│  │   │  Worker 2 ─┤   • Exactly-once processing (Redis Streams ACK)    │ │   │
│  │   │  Worker 3 ─┤   • Batch writes to Postgres                       │ │   │
│  │   │  Worker N ─┘   • Auto-scale based on queue depth                │ │   │
│  │   └─────────────────────────────────────────────────────────────────┘ │   │
│  │                              │                                         │   │
│  │         ┌────────────────────┼────────────────────┐                   │   │
│  │         ▼                    ▼                    ▼                   │   │
│  │  ┌─────────────┐    ┌─────────────────┐   ┌─────────────────────┐    │   │
│  │  │   Redis     │    │    Postgres     │   │   Quota & Alert     │    │   │
│  │  │  Counters   │    │   (Durable)     │   │     Engine          │    │   │
│  │  │             │    │                 │   │                     │    │   │
│  │  │ Real-time   │    │ usage_records   │   │ if usage >= 80%:    │    │   │
│  │  │ counters:   │    │ (partitioned)   │   │   → quota_warning   │    │   │
│  │  │             │    │                 │   │                     │    │   │
│  │  │ customer:   │    │ usage_daily     │   │ if usage >= 95%:    │    │   │
│  │  │ {id}:{month}│    │ (aggregated)    │   │   → quota_critical  │    │   │
│  │  │ = 45231     │    │                 │   │                     │    │   │
│  │  │             │    │ Retention:      │   │ if usage >= 100%:   │    │   │
│  │  │ TTL: 35 days│    │ 12 months       │   │   → check overage   │    │   │
│  │  │             │    │ then archive    │   │     or block        │    │   │
│  │  └─────────────┘    └─────────────────┘   └─────────────────────┘    │   │
│  │                                                                        │   │
│  └────────────────────────────────────────────────────────────────────────┘   │
│                                                                               │
└───────────────────────────────────────────────────────────────────────────────┘
```

### 4.3 Security Guarantees for Customers

| Customer Concern | Our Answer | Technical Implementation |
|------------------|------------|--------------------------|
| "What if your database is breached?" | Attackers get only hashes, keys are useless | SHA-256 one-way hash, no plain text storage |
| "Can your employees see my API key?" | No, we never store the plain key | Key shown once at creation, only hash stored |
| "What if my key is leaked?" | Revoke instantly, effective in <1 second | Redis pub/sub for real-time invalidation |
| "How do I safely rotate keys?" | Zero-downtime rotation with 24h grace period | Both old and new keys valid during grace |
| "Can someone use my key from another location?" | Set IP whitelist to restrict usage | CIDR-based IP validation per key |
| "How long can a key be valid?" | Set expiration policy (optional) | Configurable TTL with auto-expiry |
| "What data do you log?" | Only metadata, never request/response content | Strict logging policy, no PII in tracking |
| "Who used my API and when?" | Full audit trail available | Detailed logs with api_key_id, timestamp |
| "Is my key encrypted in transit?" | Yes, TLS 1.3 enforced | Certificate pinning optional |
| "What about at rest?" | Yes, encrypted storage | AES-256 for database encryption |

### 4.4 Security Compliance Checklist

#### 4.4.1 Data Protection

| Requirement | Status | Implementation |
|-------------|--------|----------------|
| Encryption at Rest | ✅ | PostgreSQL TDE, MinIO encryption |
| Encryption in Transit | ✅ | TLS 1.3, HSTS enabled |
| Key Management | ✅ | HashiCorp Vault / K8S Secrets |
| Data Retention Policy | ✅ | 12 months active, then archive |
| Data Deletion | ✅ | GDPR-compliant deletion on request |
| Backup Encryption | ✅ | Encrypted backups to separate location |

#### 4.4.2 Access Control

| Requirement | Status | Implementation |
|-------------|--------|----------------|
| Authentication | ✅ | Keycloak OIDC/OAuth2 |
| Authorization | ✅ | RBAC with fine-grained permissions |
| MFA for Admin | ✅ | TOTP/WebAuthn via Keycloak |
| Session Management | ✅ | Short-lived tokens, refresh rotation |
| API Key Scopes | ✅ | read/write/admin permissions |
| IP Restrictions | ✅ | Per-key CIDR whitelist |

#### 4.4.3 Audit & Monitoring

| Requirement | Status | Implementation |
|-------------|--------|----------------|
| Audit Logging | ✅ | All admin actions logged |
| Access Logs | ✅ | Kong access logs, immutable storage |
| Anomaly Detection | ✅ | ML-based pattern analysis |
| Real-time Alerting | ✅ | PagerDuty/Slack integration |
| Log Retention | ✅ | 90 days hot, 1 year cold storage |
| Tamper-proof Logs | ✅ | Append-only, signed entries |

#### 4.4.4 Infrastructure Security

| Requirement | Status | Implementation |
|-------------|--------|----------------|
| Network Segmentation | ✅ | K8S network policies |
| Pod Security | ✅ | Non-root, read-only filesystem |
| Secret Management | ✅ | Sealed Secrets / External Secrets |
| Vulnerability Scanning | ✅ | Trivy in CI/CD pipeline |
| Dependency Scanning | ✅ | Dependabot / Snyk |
| Penetration Testing | ✅ | Annual third-party pentest |

---

## 5. Technical Architecture

### 5.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           KUBERNETES CLUSTER                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                         INGRESS LAYER                                │   │
│  │  ┌─────────────┐                           ┌─────────────────────┐  │   │
│  │  │   Nginx     │                           │    Cert Manager     │  │   │
│  │  │  Ingress    │                           │    (Let's Encrypt)  │  │   │
│  │  └──────┬──────┘                           └─────────────────────┘  │   │
│  └─────────┼───────────────────────────────────────────────────────────┘   │
│            │                                                                │
│  ┌─────────▼───────────────────────────────────────────────────────────┐   │
│  │                         API GATEWAY LAYER                            │   │
│  │                                                                      │   │
│  │  ┌─────────────────────────────────────────────────────────────┐    │   │
│  │  │                        KONG GATEWAY                          │    │   │
│  │  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────────┐│    │   │
│  │  │  │   Auth   │ │  Rate    │ │  Usage   │ │   Request        ││    │   │
│  │  │  │  Plugin  │ │  Limit   │ │ Tracking │ │   Transform      ││    │   │
│  │  │  └──────────┘ └──────────┘ └──────────┘ └──────────────────┘│    │   │
│  │  └─────────────────────────────────────────────────────────────┘    │   │
│  │                                                                      │   │
│  └──────────────────────────────────┬──────────────────────────────────┘   │
│                                     │                                       │
│  ┌──────────────────────────────────┼──────────────────────────────────┐   │
│  │                         SERVICE LAYER                                │   │
│  │                                  │                                   │   │
│  │    ┌─────────────┐    ┌─────────▼─────────┐    ┌─────────────┐      │   │
│  │    │   Auth      │    │     API Usage     │    │  Billing    │      │   │
│  │    │   Service   │    │     Service       │    │  Service    │      │   │
│  │    │   (Go)      │    │     (Go)          │    │  (Go)       │      │   │
│  │    └──────┬──────┘    └─────────┬─────────┘    └──────┬──────┘      │   │
│  │           │                     │                      │             │   │
│  │    ┌──────┴──────┐    ┌─────────┴─────────┐    ┌──────┴──────┐      │   │
│  │    │ Subscription│    │    Analytics      │    │  Invoice    │      │   │
│  │    │   Service   │    │    Service        │    │  Service    │      │   │
│  │    │   (Go)      │    │    (Go)           │    │  (Go)       │      │   │
│  │    └─────────────┘    └───────────────────┘    └─────────────┘      │   │
│  │                                                                      │   │
│  │    ┌─────────────┐    ┌───────────────────┐    ┌─────────────┐      │   │
│  │    │ Notification│    │    Report         │    │  Webhook    │      │   │
│  │    │   Service   │    │    Generator      │    │  Service    │      │   │
│  │    │   (Go)      │    │    (Go)           │    │  (Go)       │      │   │
│  │    └─────────────┘    └───────────────────┘    └─────────────┘      │   │
│  │                                                                      │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                         WORKER LAYER                                  │   │
│  │                                                                       │   │
│  │  ┌───────────────┐  ┌───────────────┐  ┌───────────────────────────┐ │   │
│  │  │ Usage         │  │   Billing     │  │      Report               │ │   │
│  │  │ Aggregator    │  │   Processor   │  │      Generator            │ │   │
│  │  │ (Stream       │  │   (CronJob)   │  │      (CronJob)            │ │   │
│  │  │  Consumer)    │  │               │  │                           │ │   │
│  │  └───────────────┘  └───────────────┘  └───────────────────────────┘ │   │
│  │                                                                       │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                         FRONTEND LAYER                                │   │
│  │                                                                       │   │
│  │  ┌───────────────────────────┐  ┌───────────────────────────────────┐│   │
│  │  │     Admin Dashboard       │  │       Customer Portal             ││   │
│  │  │     (Next.js)             │  │       (Next.js)                   ││   │
│  │  └───────────────────────────┘  └───────────────────────────────────┘│   │
│  │                                                                       │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                         DATA LAYER                                    │   │
│  │                                                                       │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │   │
│  │  │  Keycloak   │  │   Redis     │  │  Postgres   │  │    MinIO    │  │   │
│  │  │  (Auth)     │  │  Cluster    │  │  (Primary   │  │  (S3-compat │  │   │
│  │  │             │  │             │  │   + Read    │  │   Storage)  │  │   │
│  │  │             │  │             │  │   Replicas) │  │             │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘  │   │
│  │                                                                       │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                         OBSERVABILITY                                 │   │
│  │                                                                       │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │   │
│  │  │ Prometheus  │  │   Grafana   │  │    Loki     │  │   Jaeger    │  │   │
│  │  │ (Metrics)   │  │ (Dashboard) │  │   (Logs)    │  │  (Tracing)  │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘  │   │
│  │                                                                       │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 5.2 Technology Stack

#### 5.2.1 Core Infrastructure

| Component | Technology | Justification |
|-----------|------------|---------------|
| Container Orchestration | **Kubernetes (K8S)** | Production-grade orchestration, auto-scaling, self-healing |
| API Gateway | **Kong Gateway (OSS)** | Plugin ecosystem, high performance, declarative config |
| Identity & Access | **Keycloak** | OIDC/OAuth2, multi-tenant, RBAC, social login |
| Cache & Queue | **Redis Cluster** | Streams for events, counters for real-time, pub/sub |
| Primary Database | **PostgreSQL 16** | ACID, JSONB, partitioning, mature ecosystem |
| Object Storage | **MinIO** | S3-compatible, self-hosted, cost-effective |
| Ingress Controller | **Nginx Ingress** | Battle-tested, SSL termination |
| Cert Management | **Cert Manager** | Auto-renewal, Let's Encrypt |

#### 5.2.2 Application Layer

| Component | Technology | Justification |
|-----------|------------|---------------|
| Backend Services | **Go 1.22+** | Performance, concurrency, small footprint |
| Web Framework | **Fiber v2** | Fast, Express-like API, low memory |
| ORM | **GORM** | Feature-rich, migrations, hooks |
| Frontend | **Next.js 14** | SSR/SSG, React ecosystem, API routes |
| UI Library | **shadcn/ui + Tailwind** | Modern, accessible, customizable |
| Charts | **Recharts / Apache ECharts** | Rich visualization |

#### 5.2.3 Messaging & Events

| Component | Technology | Justification |
|-----------|------------|---------------|
| Event Streaming | **Redis Streams** | Lightweight, built-in with Redis |
| Background Jobs | **Asynq** | Go-native, Redis-backed, reliable |
| Webhook Delivery | **Custom (with retry)** | Control over retry logic |

#### 5.2.4 Observability Stack

| Component | Technology | Justification |
|-----------|------------|---------------|
| Metrics | **Prometheus** | De facto standard, PromQL |
| Visualization | **Grafana** | Rich dashboards, alerting |
| Logging | **Loki + Promtail** | Log aggregation, Grafana-native |
| Tracing | **Jaeger** | Distributed tracing, OpenTelemetry |
| APM | **OpenTelemetry** | Vendor-neutral instrumentation |

#### 5.2.5 DevOps & CI/CD

| Component | Technology | Justification |
|-----------|------------|---------------|
| CI/CD | **GitHub Actions** | Native integration, generous free tier |
| Container Registry | **GitHub Container Registry** | Free for public, close to code |
| IaC | **Terraform** | Cloud-agnostic, state management |
| K8S Manifests | **Kustomize / Helm** | Environment overlays |
| Secret Management | **Sealed Secrets / SOPS** | GitOps-friendly encryption |

#### 5.2.6 External Integrations

| Integration | Options | Purpose |
|-------------|---------|---------|
| Payment Gateway | Omise, Stripe, SCB API | Payment processing |
| Email | AWS SES, SendGrid, Mailgun | Transactional emails |
| SMS | ThaiBulkSMS, Twilio | Critical alerts |
| Tax | e-Tax API (กรมสรรพากร) | Tax invoice submission |

---

### 5.3 Database Schema (Core Tables)

```sql
-- Organizations (API Providers)
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Customers (API Consumers)
CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organizations(id),
    keycloak_user_id UUID UNIQUE NOT NULL,
    company_name VARCHAR(255),
    tax_id VARCHAR(20),
    email VARCHAR(255) NOT NULL,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Subscription Tiers
CREATE TABLE subscription_tiers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organizations(id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price_monthly DECIMAL(12,2) NOT NULL,
    price_yearly DECIMAL(12,2),
    quota_requests BIGINT,          -- NULL = unlimited
    quota_bandwidth_mb BIGINT,       -- NULL = unlimited
    rate_limit_per_second INT,
    overage_rate DECIMAL(10,4),      -- per request after quota
    features JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Customer Subscriptions
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID REFERENCES customers(id),
    tier_id UUID REFERENCES subscription_tiers(id),
    status VARCHAR(20) DEFAULT 'active',  -- active, paused, cancelled
    billing_cycle VARCHAR(20) DEFAULT 'monthly',
    current_period_start DATE NOT NULL,
    current_period_end DATE NOT NULL,
    cancelled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- API Keys
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID REFERENCES customers(id),
    key_hash VARCHAR(64) NOT NULL,  -- SHA-256 hash
    key_prefix VARCHAR(8) NOT NULL,  -- For display: "apik_live_a1..."
    name VARCHAR(100),
    permissions JSONB DEFAULT '["read"]',
    ip_whitelist INET[],
    expires_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);

-- Usage Records (Partitioned by month)
CREATE TABLE usage_records (
    id BIGSERIAL,
    customer_id UUID NOT NULL,
    api_key_id UUID,
    endpoint VARCHAR(500) NOT NULL,
    method VARCHAR(10) NOT NULL,
    status_code SMALLINT NOT NULL,
    request_size_bytes INT,
    response_size_bytes INT,
    latency_ms INT,
    recorded_at TIMESTAMPTZ DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
) PARTITION BY RANGE (recorded_at);

-- Create partitions (example for 2025)
CREATE TABLE usage_records_2025_01 PARTITION OF usage_records
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

-- Daily Usage Aggregates (for faster queries)
CREATE TABLE usage_daily (
    id BIGSERIAL PRIMARY KEY,
    customer_id UUID REFERENCES customers(id),
    api_key_id UUID REFERENCES api_keys(id),
    date DATE NOT NULL,
    total_requests BIGINT DEFAULT 0,
    successful_requests BIGINT DEFAULT 0,
    failed_requests BIGINT DEFAULT 0,
    total_request_bytes BIGINT DEFAULT 0,
    total_response_bytes BIGINT DEFAULT 0,
    avg_latency_ms DECIMAL(10,2),
    p95_latency_ms DECIMAL(10,2),
    p99_latency_ms DECIMAL(10,2),
    endpoints_breakdown JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(customer_id, api_key_id, date)
);

-- Invoices
CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_number VARCHAR(50) UNIQUE NOT NULL,
    customer_id UUID REFERENCES customers(id),
    subscription_id UUID REFERENCES subscriptions(id),
    billing_period_start DATE NOT NULL,
    billing_period_end DATE NOT NULL,
    subtotal DECIMAL(12,2) NOT NULL,
    tax_rate DECIMAL(5,4) DEFAULT 0.07,
    tax_amount DECIMAL(12,2) NOT NULL,
    total DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'THB',
    status VARCHAR(20) DEFAULT 'draft',  -- draft, pending, paid, overdue, cancelled
    due_date DATE NOT NULL,
    paid_at TIMESTAMPTZ,
    pdf_url VARCHAR(500),
    line_items JSONB NOT NULL,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Payments
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID REFERENCES invoices(id),
    amount DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'THB',
    payment_method VARCHAR(50),  -- credit_card, bank_transfer, promptpay
    payment_gateway VARCHAR(50),  -- omise, stripe, manual
    gateway_transaction_id VARCHAR(255),
    status VARCHAR(20) DEFAULT 'pending',
    paid_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Webhook Endpoints (Customer-defined)
CREATE TABLE webhook_endpoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID REFERENCES customers(id),
    url VARCHAR(500) NOT NULL,
    secret VARCHAR(100) NOT NULL,  -- For signature verification
    events TEXT[] NOT NULL,  -- ['usage.quota_warning', 'invoice.created']
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Webhook Deliveries (Audit log)
CREATE TABLE webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    endpoint_id UUID REFERENCES webhook_endpoints(id),
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    response_status SMALLINT,
    response_body TEXT,
    attempts INT DEFAULT 1,
    next_retry_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

---

### 5.4 API Specification (Core Endpoints)

#### Public APIs (Customer-facing)

```yaml
# Usage APIs
GET  /v1/usage/current          # Current billing period usage
GET  /v1/usage/history          # Historical usage with date range
GET  /v1/usage/realtime         # WebSocket for live updates

# Subscription APIs
GET  /v1/subscription           # Current subscription details
GET  /v1/subscription/tiers     # Available tiers
POST /v1/subscription/upgrade   # Upgrade tier
POST /v1/subscription/downgrade # Downgrade tier

# API Key Management
GET    /v1/api-keys             # List all keys
POST   /v1/api-keys             # Create new key
DELETE /v1/api-keys/:id         # Revoke key
POST   /v1/api-keys/:id/rotate  # Rotate key

# Billing APIs
GET  /v1/invoices               # List invoices
GET  /v1/invoices/:id           # Invoice details
GET  /v1/invoices/:id/pdf       # Download PDF

# Webhook APIs
GET    /v1/webhooks             # List endpoints
POST   /v1/webhooks             # Create endpoint
DELETE /v1/webhooks/:id         # Delete endpoint
POST   /v1/webhooks/:id/test    # Send test event
```

#### Admin APIs (Provider-facing)

```yaml
# Customer Management
GET    /admin/customers
GET    /admin/customers/:id
POST   /admin/customers/:id/suspend
POST   /admin/customers/:id/unsuspend

# Tier Management
GET    /admin/tiers
POST   /admin/tiers
PUT    /admin/tiers/:id
DELETE /admin/tiers/:id

# Analytics
GET    /admin/analytics/revenue
GET    /admin/analytics/usage
GET    /admin/analytics/top-customers
GET    /admin/analytics/top-apis

# Billing Operations
POST   /admin/invoices/generate    # Manual generation
POST   /admin/invoices/:id/resend  # Resend invoice email
POST   /admin/credit-notes         # Issue credit note
```

---

## 6. User Stories

### Epic 1: Usage Tracking

| ID | As a | I want to | So that | Priority |
|----|------|-----------|---------|----------|
| US-1.1 | Customer | See my current API usage in real-time | I can monitor my consumption | P0 |
| US-1.2 | Customer | View usage breakdown by API endpoint | I can identify heavy usage patterns | P1 |
| US-1.3 | Customer | See usage per API key | I can track usage by application | P1 |
| US-1.4 | Provider | Track all customer usage | I can bill accurately | P0 |
| US-1.5 | Provider | Set custom metering rules | I can charge differently per API | P1 |

### Epic 2: Subscription Management

| ID | As a | I want to | So that | Priority |
|----|------|-----------|---------|----------|
| US-2.1 | Customer | See available subscription tiers | I can choose the right plan | P0 |
| US-2.2 | Customer | Upgrade my subscription | I can get more quota/features | P0 |
| US-2.3 | Customer | Downgrade my subscription | I can reduce costs | P1 |
| US-2.4 | Provider | Create custom subscription tiers | I can offer flexible pricing | P0 |
| US-2.5 | Provider | Configure quota limits per tier | I can differentiate offerings | P0 |

### Epic 3: Billing & Invoicing

| ID | As a | I want to | So that | Priority |
|----|------|-----------|---------|----------|
| US-3.1 | Customer | Receive automatic monthly invoices | I don't have to request them | P0 |
| US-3.2 | Customer | Download PDF invoices | I can use them for accounting | P0 |
| US-3.3 | Customer | Pay invoices online | I can pay conveniently | P0 |
| US-3.4 | Provider | Generate invoices automatically | I don't have to do it manually | P0 |
| US-3.5 | Finance | See all pending payments | I can follow up on overdue | P0 |
| US-3.6 | Finance | Issue credit notes | I can handle refunds/adjustments | P1 |

### Epic 4: Alerting

| ID | As a | I want to | So that | Priority |
|----|------|-----------|---------|----------|
| US-4.1 | Customer | Get notified when approaching quota | I can plan upgrades | P0 |
| US-4.2 | Customer | Receive invoice reminders | I don't miss payments | P1 |
| US-4.3 | Customer | Configure webhook endpoints | I can integrate with my systems | P1 |
| US-4.4 | Provider | Alert on suspicious usage | I can detect abuse | P2 |

### Epic 5: API Key Management

| ID | As a | I want to | So that | Priority |
|----|------|-----------|---------|----------|
| US-5.1 | Customer | Create multiple API keys | I can use different keys per app | P0 |
| US-5.2 | Customer | Rotate keys without downtime | I can maintain security | P1 |
| US-5.3 | Customer | Set IP restrictions on keys | I can enhance security | P2 |
| US-5.4 | Customer | See which key made each request | I can debug issues | P1 |

---

## 7. Implementation Phases

### Phase 1: Foundation (8 weeks)

**Goal:** Basic usage tracking and subscription management

| Week | Deliverables |
|------|--------------|
| 1-2 | K8S cluster setup, Kong deployment, Keycloak config |
| 3-4 | PostgreSQL setup, Redis cluster, MinIO deployment |
| 5-6 | Usage tracking service, Kong plugin for logging |
| 7-8 | Subscription service, basic customer portal |

**Milestone:** Can track API usage and manage subscriptions

### Phase 2: Billing (6 weeks)

**Goal:** Automated billing and invoicing

| Week | Deliverables |
|------|--------------|
| 9-10 | Invoice generation service, PDF generation |
| 11-12 | Payment gateway integration (Omise) |
| 13-14 | Billing CronJobs, email notifications |

**Milestone:** Automatic monthly billing cycle working

### Phase 3: Analytics & Alerts (4 weeks)

**Goal:** Rich dashboards and proactive alerts

| Week | Deliverables |
|------|--------------|
| 15-16 | Analytics service, Grafana dashboards |
| 17-18 | Alerting system, webhook delivery |

**Milestone:** Full visibility into usage and revenue

### Phase 4: Polish & Scale (4 weeks)

**Goal:** Production-ready system

| Week | Deliverables |
|------|--------------|
| 19-20 | Performance optimization, load testing |
| 21-22 | Documentation, monitoring, security audit |

**Milestone:** Ready for production launch

---

## 8. Risks & Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Data loss during high traffic | High | Medium | Redis Streams with persistence, async write to Postgres |
| Invoice calculation errors | High | Low | Extensive testing, reconciliation jobs |
| Payment gateway downtime | Medium | Low | Support multiple gateways, manual fallback |
| Kong plugin performance | Medium | Medium | Benchmark testing, async logging |
| Keycloak single point of failure | High | Low | HA deployment, token caching |

---

## 9. Success Criteria

### MVP Success (Phase 1-2 Complete)

- [ ] Track 100% of API requests
- [ ] Generate invoices within 5 minutes of month end
- [ ] 99.9% uptime for tracking system
- [ ] Customer can see usage in real-time (< 5 sec delay)

### Full Product Success (All Phases Complete)

- [ ] Support 10,000+ customers
- [ ] Handle 50,000 requests/second tracking
- [ ] Invoice accuracy 99.99%
- [ ] Customer satisfaction score > 4.5/5
- [ ] Zero manual intervention for standard billing

---

## 10. Appendices

### A. Glossary

| Term | Definition |
|------|------------|
| API Provider | Organization that offers APIs and uses this platform to monetize them |
| Customer | End-user who consumes APIs and pays for usage |
| Tier | Subscription level with specific quota and pricing |
| Overage | Usage beyond the included quota, billed separately |
| Metering | Method of counting usage (per request, per MB, etc.) |
| Key Hash | SHA-256 hash of API key, stored instead of plain key |
| Key Rotation | Process of replacing an API key with a new one |
| Grace Period | Overlap time when both old and new keys are valid |

### B. Security Threat Model

| Threat | Impact | Likelihood | Mitigation |
|--------|--------|------------|------------|
| Database breach exposing API keys | High | Medium | Store only SHA-256 hashes |
| Man-in-the-middle attack | High | Low | TLS 1.3, certificate pinning |
| Stolen API key | High | Medium | IP whitelist, instant revoke, anomaly detection |
| Brute force key guessing | Medium | Low | 256-bit entropy, rate limiting on auth |
| Insider threat | High | Low | Audit logging, least privilege, no plain key access |
| DDoS on tracking pipeline | Medium | Medium | Redis Streams backpressure, auto-scaling |
| Invoice tampering | High | Low | Immutable audit log, signed invoices |

### C. API Key Security Specifications

```
Key Format:
┌────────────────────────────────────────────────────────────┐
│  apik_example_a1b2c3d4e5f6g7h8i9j0o5p6q7r8s9t0u1v2    │
│  ├──────┼────┼──────────────────────────────────────────┤   │
│  │prefix│env │           random (32 bytes)              │   │
│  │ sk_  │live│           base62 encoded                 │   │
│  └──────┴────┴──────────────────────────────────────────┘   │
└────────────────────────────────────────────────────────────┘

Entropy Calculation:
- Random part: 32 bytes = 256 bits
- Possible combinations: 2^256 ≈ 1.16 × 10^77
- Brute force at 1 trillion guesses/sec: 3.7 × 10^57 years

Storage Format (PostgreSQL):
- key_prefix: VARCHAR(12) - "apik_live_a1" (for UI display)
- key_hash: CHAR(64) - SHA-256 hex digest
- key_salt: CHAR(32) - Optional additional salt (future)
```

### D. References

- [Kong Gateway Documentation](https://docs.konghq.com)
- [Keycloak Admin Guide](https://www.keycloak.org/docs/latest/server_admin/)
- [Stripe Billing Model](https://stripe.com/docs/billing) (for inspiration)
- [AWS API Gateway Usage Plans](https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-api-usage-plans.html)

### E. Competitive Analysis

| Feature | Our Platform | Stripe Billing | AWS API Gateway | Moesif |
|---------|--------------|----------------|-----------------|--------|
| Usage Tracking | ✅ | ✅ | ✅ | ✅ |
| Custom Metering | ✅ | ✅ | ❌ | ✅ |
| Self-hosted | ✅ | ❌ | ❌ | ❌ |
| Thai Payment Gateway | ✅ | ⚠️ | ❌ | ❌ |
| Thai Tax Invoice | ✅ | ❌ | ❌ | ❌ |
| Pricing | Self-hosted cost | % of revenue | Per request | Subscription |

---

*Document End*
