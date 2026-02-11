# ORC Database Schema

**Status**: Living document
**Last Updated**: 2026-02-08

This document contains the core entity-relationship diagram for ORC's database schema.

For the complete schema including messaging and auxiliary tables, see `internal/db/schema.sql`.

---

## Core Entity Relationships

```mermaid
erDiagram
    FACTORY ||--o{ WORKSHOP : contains
    FACTORY ||--o{ COMMISSION : owns
    WORKSHOP ||--o{ WORKBENCH : contains
    WORKSHOP ||--|| GATEHOUSE : "has one"
    COMMISSION ||--o{ SHIPMENT : contains
    COMMISSION ||--o{ TOME : contains
    SHIPMENT ||--o{ TASK : contains
    SHIPMENT ||--o{ NOTE : contains
    TOME ||--o{ NOTE : contains
    TASK ||--o{ PLAN : "planned by"
    TASK ||--o| RECEIPT : "completed with"
    PLAN ||--o| APPROVAL : "approved by"

    FACTORY {
        string id PK
        string name
        string status
    }
    WORKSHOP {
        string id PK
        string factory_id FK
        string name
        string status
        string active_commission_id FK
    }
    WORKBENCH {
        string id PK
        string workshop_id FK
        string name
        string repo_id FK
        string status
        string focused_id
    }
    GATEHOUSE {
        string id PK
        string workshop_id FK
        string status
        string focused_id
    }
    COMMISSION {
        string id PK
        string factory_id FK
        string title
        string status
        boolean pinned
    }
    SHIPMENT {
        string id PK
        string commission_id FK
        string title
        string status
        string branch
        boolean pinned
    }
    TASK {
        string id PK
        string shipment_id FK
        string commission_id FK
        string title
        string status
        string type
        string priority
    }
    TOME {
        string id PK
        string commission_id FK
        string title
        string status
        boolean pinned
    }
    NOTE {
        string id PK
        string commission_id FK
        string shipment_id FK
        string tome_id FK
        string title
        string type
        string status
    }
    PLAN {
        string id PK
        string task_id FK
        string commission_id FK
        string title
        string status
        text content
    }
    APPROVAL {
        string id PK
        string plan_id FK
        string task_id FK
        string mechanism
        string outcome
    }
    RECEIPT {
        string id PK
        string task_id FK
        string delivered_outcome
        string status
    }
```

---

## Table Descriptions

| Table | Purpose | Key Fields |
|-------|---------|------------|
| **factories** | TMux sessions / runtime environments | name, status |
| **workshops** | TMux sessions within a factory | factory_id, name, active_commission_id |
| **workbenches** | Git worktrees within a workshop | workshop_id, repo_id, focused_id |
| **gatehouses** | Goblin seats (1:1 with workshop) | workshop_id, focused_id |
| **commissions** | Top-level coordination scopes | factory_id, title, status |
| **shipments** | Work containers with lifecycle | commission_id, title, status, branch |
| **tasks** | Atomic units of work | shipment_id, title, status, type, priority |
| **tomes** | Knowledge containers | commission_id, title, status |
| **notes** | Observations, learnings, decisions | shipment_id, tome_id, title, type |
| **plans** | Implementation plans (1:many with task) | task_id, title, content, status |
| **approvals** | Plan approvals (1:1 with plan) | plan_id, mechanism, outcome |
| **receipts** | Task completion records (1:1 with task) | task_id, delivered_outcome, status |

---

## Hierarchy Summary

**Infrastructure:**
```
Factory → Workshop → Workbench
                  → Gatehouse (1:1)
```

**Work Tracking:**
```
Commission → Shipment → Task → Plan → Approval
                     → Note     → Receipt
          → Tome → Note
```

---

## See Also

- `internal/db/schema.sql` - Complete schema
- `docs/architecture.md` - System architecture overview
- `docs/shipment-lifecycle.md` - Shipment state machine
