---
layout: default
title: Kora
description: Config-Driven Application Engine
---

# Kora

**Describe your application in YAML. Get a database, API, and admin panel — automatically.**

---

## How It Works

```
You write this:                    Kora builds:

config/todo/                       ✅ MySQL database
  doctypes/todo.yaml               ✅ REST API (CRUD + auth)
  roles.yaml                       ✅ React admin panel
  permissions.yaml                 ✅ Forms, lists, filters
                                   ✅ Mobile responsive
```

Three YAML files. One binary. **Zero application code.**

---

## What You Get

| Feature | How |
|---|---|
| **Database** | Tables created and migrated automatically |
| **REST API** | CRUD, authentication, permissions, CSRF |
| **Admin UI** | Config-driven forms, lists, searchable dropdowns, computed fields |
| **Workflows** | State machines — Draft → Submitted → Approved |
| **Multi-site** | One server, many apps, separate databases |
| **Single binary** | Go + React SPA embedded via go:embed |

---

## What You Can Build

**[10 complete SaaS applications](USECASES.md) — all in YAML config, zero application code:**

| # | Application | Highlights |
|---|-------------|------------|
| 1 | **CRM** | Deal pipeline, computed totals, sales permissions |
| 2 | **Help Desk** | Ticket lifecycle, agent assignment, SLA tracking |
| 3 | **Project Management** | Task hierarchy, progress tracking, COUNT aggregation |
| 4 | **Inventory** | Real-time stock levels, warehouse movements |
| 5 | **Recruitment** | Job pipeline, candidate tracking, resume uploads |
| 6 | **Invoicing** | Line items, tax calculation, payment lifecycle |
| 7 | **Property Management** | Lease tracking, occupancy, DATEDIFF expiry |
| 8 | **LMS** | Course builder, enrollments, student tracking |
| 9 | **Event Management** | Registrations, capacity management, check-in |
| 10 | **Contract Management** | Obligations, renewals, DATEDIFF countdown |

Each comes with database, REST API, React admin panel, role-based permissions, and workflow automation — from 3-6 YAML files.

### Quick Example: Todo

```yaml
name: Todo
module: Tasks
title_field: title

fields:
  - fieldname: title
    fieldtype: Data
    reqd: true
  - fieldname: status
    fieldtype: Select
    options: |
      Pending
      In Progress
      Done
  - fieldname: due_date
    fieldtype: Date
```

---

## Quick Start

```bash
git clone https://github.com/asenawritescode/kora.git
cd kora
docker compose up -d mysql
make dev
```

Open **http://localhost:8000/s/todo/workspace**

---

## Documentation

- [**What You Can Build**](USECASES.md) — 10 SaaS applications with config examples
- [Setup Guide](SETUP.md) — Prerequisites, installation, multi-site, production
- [Configuration Reference](CONFIG.md) — DocTypes, fields, constraints, workflows
- [API Reference](API.md) — REST endpoints, auth, envelope format
- [Architecture](ARCHITECTURE.md) — Request flow, middleware, multi-tenancy
- [Decisions](DECISIONS.md) — Why React SPA, computed fields, path-based routing

---

## License

AGPL-3.0 — Free software. Network use is distribution.

<p style="margin-top: 3rem; font-size: 0.85rem; color: #666">
  Built with Go, React, and conviction.
</p>
