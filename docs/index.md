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

## Sample Apps

### Todo
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

### Airtime Sales
```
Customer → Order → Product workflow
automated price calculation
role-based approvals
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
