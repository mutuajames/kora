# Kora — Config-Driven Application Engine

Define your application — data model, permissions, workflows — in YAML. Kora gives you a database, REST API, React admin UI, and background jobs. No code generation.

## Quick Start

```bash
docker compose up -d mysql                           # Start MySQL
cd ui && bun install && bun run build && cp -r dist ../workspace/dist && cd ..  # Build UI
go run . setup --site airtime.local --path config/airtime/ --db-user root --db-pass kora123 --admin-email admin@airtime.local --admin-password admin123
go run . serve --port 8000                           # Start server
```

Open **http://localhost:8000/workspace** — login with `admin@airtime.local` / `admin123`.

## Multi-Site

```
http://localhost:8000/s/airtime/workspace     → Airtime app
http://localhost:8000/s/fieldwork/workspace   → Fieldwork app
http://localhost:8000/console/login           → System console
```

No DNS or `/etc/hosts` needed. Path-based routing just works. For production, add real domains — Host-based routing takes over automatically.

## Documentation

| Document | What it covers |
|---|---|
| [SETUP.md](docs/SETUP.md) | Prerequisites, quick start, multi-site setup, config management, production deployment, troubleshooting |
| [ARCHITECTURE.md](docs/ARCHITECTURE.md) | System design, request flow, middleware, multi-tenancy, expression engine, schema migration, computed fields |
| [CONFIG.md](docs/CONFIG.md) | DocType/field reference, constraints, workflows, permissions, link fields, computed expressions, back-references |
| [API.md](docs/API.md) | REST API reference, auth, CRUD, workflow actions, system endpoints, error formats |
| [DECISIONS.md](docs/DECISIONS.md) | Architecture Decision Records — why React SPA, config-driven computed fields, path-based multi-site, Gin NoRoute, site-aware auth |
| [NETWORKING.md](docs/NETWORKING.md) | TLS, autocert, HTTP→HTTPS, rate limiting, security headers, CORS |

## Project Structure

```
kora/
├── cli/            # Cobra CLI: serve, setup, migrate, config
├── api/            # REST handlers, CRUD, system endpoints
├── auth/           # Session auth, CSRF, SystemGuard, SiteGuard
├── net/            # SiteRouter, TLS, security headers, rate limiting
├── doctype/        # DocType, Field, Registry, permissions, workflow, expressions
├── orm/            # Generic CRUD on map[string]any documents
├── schema/         # INFORMATION_SCHEMA diff → DDL migration
├── configstore/    # Config persistence (_kora_* tables)
├── workspace/      # React SPA serving (go:embed)
├── console/        # System admin console (server-rendered)
├── scheduler/      # Cron-style background jobs
├── site/           # Site config loading, DB connection
├── email/          # Email sending (mock for dev)
├── config/         # Sample app YAML configs (airtime, fieldwork)
├── ui/             # React 19 SPA (Vite + TanStack + shadcn/ui)
├── docs/           # Documentation
└── sites/          # Per-site config and files
```

## Tech Stack

**Backend:** Go 1.25, Gin, MySQL 8.0, expr-lang/expr, Cobra CLI, autocert (Let's Encrypt)
**Frontend:** React 19, TanStack Router/Query/Table/Form, shadcn/ui, Tailwind CSS v4, Zustand
**Single binary:** Everything embedded via `go:embed` — no separate deployment
