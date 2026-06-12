---
layout: default
title: What You Can Build
description: 10 ready-to-deploy SaaS applications built entirely in YAML configuration
---

# What You Can Build with Kora

All ten applications below are built **entirely in YAML configuration files** — zero application code. Each comes with a database, REST API, React admin panel, role-based permissions, and workflow automation.

---

## 1. CRM — Customer Relationship Management

**Track leads, manage deals through a sales pipeline, and close revenue.**

| Feature | How It's Done |
|---------|---------------|
| Contacts & Companies | Link fields connecting people to organizations |
| Deal pipeline | 6-stage workflow: Lead → Qualified → Proposal → Negotiation → Closed Won/Lost |
| Deal value tracking | Computed fields: `line_total = quantity × unit_price`, `total = SUM(items.line_total)` |
| Sales permissions | Sales Reps see only their own deals; Managers see everything |
| Activity log | Calls, meetings, emails linked to contacts and deals |

**Key Kora features**: Link fields, Table fields (line items), workflow state machine, computed fields, owner-scoped permissions.

---

## 2. Help Desk — Ticketing System

**Manage support tickets from open to resolution with SLA tracking.**

| Feature | How It's Done |
|---------|---------------|
| Ticket lifecycle | 6-stage workflow: Open → In Progress → Waiting → Resolved → Closed |
| Agent assignment | Required fields enforced at workflow transitions |
| Priority levels | Select field: Low, Medium, High, Critical |
| Team permissions | Agents see own tickets; Managers see all |

**Key Kora features**: Multi-stage workflow with required fields, role-based permissions.

---

## 3. Project Management

**Plan projects, assign tasks, and track progress — including task hierarchies.**

| Feature | How It's Done |
|---------|---------------|
| Task hierarchy | Self-referencing Link field: Task → Parent Task |
| Progress tracking | Computed: `completion_pct = tasks_done / total_tasks × 100` |
| Task counting | Computed: `COUNT(tasks)` shows total tasks per project |
| Team assignment | Data fields for assignee, priority, and status |

**Key Kora features**: Self-referencing Link, COUNT aggregation, computed percentages.

---

## 4. Inventory Management

**Track products, stock levels, and warehouse movements.**

| Feature | How It's Done |
|---------|---------------|
| Real-time stock | Computed: `stock_qty = SUM(stock_moves.qty_change)` |
| Stock movements | Child table tracking receipts (+), shipments (-), adjustments |
| Reorder alerts | Reorder level field compared against computed stock |
| Product catalog | SKU with unique constraint, categories, pricing |

**Key Kora features**: SUM aggregation over child table, unique constraints, computed inventory levels.

---

## 5. Recruitment — Applicant Tracking

**Post jobs, track candidates through the hiring pipeline.**

| Feature | How It's Done |
|---------|---------------|
| Job lifecycle | 5-stage workflow: Open → Interviewing → Offer → Filled / Cancelled |
| Candidate tracking | Child table with stage tracking per candidate |
| Resume storage | Attach field type for file uploads |
| Applicant count | Computed: `COUNT(candidates)` per job opening |

**Key Kora features**: File attachments, COUNT aggregation, multi-stage workflow.

---

## 6. Invoicing & Billing

**Create invoices, calculate tax, and track payment status.**

| Feature | How It's Done |
|---------|---------------|
| Line items | Child table with `line_total = quantity × unit_price` |
| Tax calculation | Computed: `tax = ROUND(subtotal × tax_rate / 100, 2)` |
| Invoice total | Computed: `total = ROUND(subtotal + tax, 2)` |
| Payment lifecycle | Workflow: Draft → Sent → Paid / Overdue / Cancelled |

**Key Kora features**: Nested computed fields, ROUND precision, currency handling.

---

## 7. Property Management

**Manage properties, tenants, leases, and occupancy.**

| Feature | How It's Done |
|---------|---------------|
| Lease tracking | Child table with start/end dates and rent |
| Days remaining | Computed: `DATEDIFF(today(), end_date)` shows lease expiry |
| Occupancy count | Computed: `COUNT(leases)` per property |
| Tenant management | Lease child table captures tenant details per unit |

**Key Kora features**: DATEDIFF for date calculations, COUNT for occupancy.

---

## 8. Learning Management System

**Create courses with lessons and track student enrollments.**

| Feature | How It's Done |
|---------|---------------|
| Course structure | Child tables for lessons and enrollments |
| Enrollment count | Computed: `COUNT(enrollments)` per course |
| Lesson ordering | Sortable lessons with duration tracking |
| Student progress | Enrollment records track completion status |

**Key Kora features**: Multiple child tables on one doctype, COUNT aggregation.

---

## 9. Event Management

**Organize events, track registrations, and manage capacity.**

| Feature | How It's Done |
|---------|---------------|
| Registration count | Computed: `COUNT(attendees)` — updates automatically |
| Available spots | Computed: `capacity - registration_count` |
| Venue management | Link field to Venue doctype |
| Check-in tracking | Boolean field per attendee |

**Key Kora features**: COUNT with arithmetic, capacity management via computed fields.

---

## 10. Contract Management

**Track contracts, obligations, renewals, and expiry dates.**

| Feature | How It's Done |
|---------|---------------|
| Days remaining | Computed: `DATEDIFF(today(), end_date)` — live countdown |
| Contract lifecycle | Workflow: Draft → Review → Active → Expired / Terminated |
| Obligation tracking | Child table with responsible party and due dates |
| Renewal awareness | Auto-renewal flag + notice period field |

**Key Kora features**: DATEDIFF with today(), lifecycle workflow, obligation tracking.

---

## Platform Capabilities at a Glance

| Capability | What It Means |
|------------|---------------|
| **21 field types** | Text, numbers, dates, select dropdowns, links, file attachments, child tables, rich text, JSON |
| **Computed fields** | `SUM()`, `COUNT()`, `DATEDIFF()`, `ROUND()`, arithmetic — values auto-calculate |
| **Workflow engine** | Define states and transitions with conditions, required fields, and role gating |
| **Permissions** | 10 operations (read/write/create/delete/submit/cancel/amend/export/import/report) per role per doctype |
| **Multi-tenant** | Each site = separate database. Run unlimited sites from one binary |
| **Config versioning** | Every config change is versioned with diff tracking and rollback |
| **REST API** | Full CRUD, filtering, sorting, pagination — automatically from your YAML |
| **React SPA** | Auto-generated forms, lists, search, and workflow buttons from your config |

---

## Get Started

```bash
# 1. Pick a use case
ls config/          # crm, helpdesk, projects, inventory, recruitment, invoicing, propertymgmt, lms, events, contracts

# 2. Set it up
./kora setup --site crm.local --path config/crm/ --db-user root --db-pass yourpass --admin-email you@example.com --admin-password yourpass

# 3. Start the server
./kora serve --port 8000

# 4. Open in browser
open http://localhost:8000/s/crm.local/workspace
```

**That's it.** You now have a running CRM — with database, API, and admin panel — from 6 YAML files.
