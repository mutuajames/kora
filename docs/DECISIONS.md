# Kora — Architecture Decision Records (ADRs)

## ADR-001: Config in Database, Not Filesystem

**Date:** 2026-06-10
**Status:** Accepted

**Context:** The application config (DocTypes, fields, permissions, workflows) could be stored as YAML files on disk or in the database.

**Decision:** Config lives in the database as the source of truth. YAML files are a one-shot import mechanism.

**Rationale:**
- Config can be updated at runtime without filesystem access
- Config is versioned naturally in the database
- Config is per-site in multi-tenant setups
- The API to read/write config is the same API used for all other data
- Mitigation for Git-based workflows: `kora config export` produces YAML files that can be committed

**Trade-offs:**
- Operators lose direct Git-based config management (mitigated by export/import)
- Bootstrapping a fresh site requires database access

---

## ADR-002: Go Over Python/Node.js

**Date:** 2026-06-10
**Status:** Accepted

**Context:** Frappe (the architectural inspiration) is Python. Payload CMS is Node.js. We needed to choose a language.

**Decision:** Go.

**Rationale:**
- Single binary deployment (no interpreter, no virtualenv, no node_modules)
- Strong concurrency (goroutines for job workers)
- Fast startup (milliseconds vs seconds)
- Rich stdlib (`database/sql`, `net/http`, `embed`, `crypto/tls`)
- The Frappe pattern (config-driven engine) is language-agnostic

**Trade-offs:**
- Smaller ecosystem than Python for business applications
- No dynamic module loading (Go hooks must be compiled in)

---

## ADR-003: Generic ORM with map[string]any

**Date:** 2026-06-10
**Status:** Accepted

**Context:** The engine must work with any DocType without code generation.

**Decision:** All documents are `map[string]any` at runtime. Type safety is enforced by the constraint validation layer, not the compiler.

**Rationale:**
- Zero code generation needed
- New DocTypes work immediately after config import
- The engine is truly generic

**Trade-offs:**
- No compile-time type safety for document fields
- Reflection overhead on field access
- `[]byte` from MySQL driver must be converted to `string` for JSON serialization

---

## ADR-004: MySQL Over PostgreSQL (Initially)

**Date:** 2026-06-10
**Status:** Accepted (Phase 1-3), PostgreSQL planned for Phase 4

**Context:** We needed a database for the initial implementation.

**Decision:** MySQL 8.0 for Phase 1-3. PostgreSQL support added in Phase 4.

**Rationale:**
- MySQL DDL is simpler and well-understood
- `INFORMATION_SCHEMA` works for schema introspection
- Go's `database/sql` is database-agnostic (migration layer abstracts differences)
- Frappe uses MariaDB, proving the pattern

---

## ADR-005: HTMX + Alpine.js Over React/Vue

**Date:** 2026-06-11
**Status:** Accepted

**Context:** The admin UI needs to be functional but minimal. It ships in the binary.

**Decision:** HTMX + Alpine.js + Tailwind CSS, loaded via CDN. HTML templates embedded via `go:embed`.

**Rationale:**
- No build step (no webpack, no npm, no node_modules)
- Ships in the binary via Go's `embed` package
- HTMX handles dynamic page loading without a SPA framework
- Alpine.js handles client-side interactivity without a heavy framework
- The admin UI is a thin layer over the REST API, not a separate application

**Trade-offs:**
- Less rich interactivity than a React/Vue SPA
- CDN dependency for JS/CSS (mitigated by optionally bundling these assets)
- Limited offline capability

---

## ADR-006: No External Reverse Proxy

**Date:** 2026-06-11
**Status:** Accepted

**Context:** Frappe uses nginx as a mandatory reverse proxy. We needed to decide whether Kora requires one.

**Decision:** Kora handles everything in-process. No nginx, no Caddy, no external reverse proxy.

**Rationale:**
- Go's `net/http` can serve TLS directly
- `autocert` provides automatic Let's Encrypt certificates
- `x/time/rate` provides in-process rate limiting
- Security headers, CORS, CSRF are all in-process middleware
- Single binary, single process = simpler operations
- Users CAN put a CDN or load balancer in front if they want, but it's optional

**Trade-offs:**
- No static file serving optimization (nginx is faster for static files)
- Rate limiting is per-process, not global (mitigated by Redis-backed limiter in future)
- TLS certificate management is the app's responsibility

**Alternatives considered:**
- **Lura (KrakenD engine):** API gateway framework; overkill since Kora IS the backend, not a proxy
- **Caddy as library:** Possible but Caddy is designed as a server first, library second
- **nginx:** Adds an external dependency and configuration file

---

## ADR-007: Session Cookies Over JWT

**Date:** 2026-06-11
**Status:** Accepted

**Context:** Authentication could use JWT tokens or session cookies.

**Decision:** Session cookies with optional Bearer token fallback.

**Rationale:**
- Session cookies are simpler to secure (HttpOnly, SameSite)
- Server-side session invalidation (logout, password change) works immediately
- No token refresh complexity
- Bearer token fallback supports API clients that can't use cookies
- CSRF protection via double-submit cookie pattern

**Trade-offs:**
- Requires session storage (currently DB, planned Redis)
- Not stateless (each request requires a DB lookup for the session)

---

## ADR-008: expr-lang/expr Over Custom Expression Language

**Date:** 2026-06-11
**Status:** Accepted

**Context:** Constraint conditions and workflow conditions need a safe expression language.

**Decision:** Use `expr-lang/expr` with custom functions (`today`, `now`, `len`, `has_role`).

**Rationale:**
- Safe and sandboxed (no arbitrary code execution)
- Fast (compiles to bytecode)
- Rich operator set already built-in
- Custom functions can be registered for domain-specific needs
- Avoids building and maintaining a custom expression parser

**Trade-offs:**
- Expression syntax is fixed (can't customize operators)
- Compilation errors surface at runtime (expressions are strings in config)

---

## ADR-009: Normalized Child Tables (Not JSON Blobs)

**Date:** 2026-06-11
**Status:** Accepted

**Context:** Table fields (child/parent relationships) could be stored as JSON blobs in the parent row or as normalized separate tables.

**Decision:** Normalized tables with `parent`, `parentfield`, `parenttype` columns.

**Rationale:**
- Child rows are independently queryable
- Referential integrity via foreign keys (future)
- Consistent with the Frappe pattern
- Schema migration applies to child tables too

**Trade-offs:**
- More complex INSERT/UPDATE logic (delete old children, insert new ones)
- More database tables (one child table per Table field)

---

## ADR-010: Additive-Only Schema Migrations (Default)

**Date:** 2026-06-11
**Status:** Accepted

**Context:** Schema migrations could be fully automatic (including destructive changes) or require explicit confirmation.

**Decision:** Additive changes (ADD COLUMN, CREATE INDEX) are auto-applied. Destructive changes (DROP COLUMN, CHANGE TYPE) require `--allow-breaking` flag.

**Rationale:**
- Prevents accidental data loss
- Additive changes are safe and common (adding fields)
- Destructive changes should be intentional and reviewed
- Orphaned columns accumulate until explicitly cleaned (`kora schema clean`)

**Trade-offs:**
- Operators must explicitly approve breaking changes
- Orphaned columns waste storage until cleaned

---

## ADR-011: React SPA Over HTMX+Alpine for Admin UI

**Date:** 2026-06-11
**Status:** Accepted

**Context:** The original admin UI used HTMX + Alpine.js server-rendered templates. As complexity grew (child tables, linked fields, computed expressions, autocomplete), the imperative DOM manipulation became unwieldy.

**Decision:** React 19 + Vite + TanStack (Router/Query/Table/Form) + shadcn/ui + Tailwind CSS v4. Single binary deployment via `go:embed`.

**Rationale:**
- Declarative component model handles complex form interactions naturally
- TanStack suite provides best-in-class table, form, and query management
- shadcn/ui is copy-owned (not a dependency), fully themeable via CSS variables
- TypeScript provides end-to-end type safety from API responses to UI
- No separate deployment — SPA is embedded in the Go binary
- Config-driven: zero code per DocType, everything reads from `/api/system/doctype/:name`

**Trade-offs:**
- Requires Node.js/bun for development builds
- Larger binary (~35MB with embedded SPA) vs pure Go templates
- Build step required before Go compilation

---

## ADR-012: Config-Driven Computed Fields

**Date:** 2026-06-11
**Status:** Accepted

**Context:** Documents often have derived fields (line_total = quantity × unit_price, subtotal = SUM(items.line_total)). Initially hardcoded per doctype in the frontend.

**Decision:** New `computed` field property containing an expression string. The frontend expression evaluator reads it from the doctype schema and evaluates it generically.

**Rationale:**
- Removes all hardcoded business logic from forms
- Same expression syntax for any doctype
- Expressions: arithmetic (`+`, `-`, `*`, `/`), `SUM(table.field)`, `ROUND(expr, N)`
- No backend changes needed — evaluation happens client-side against current form state
- Cascading: changing one field triggers recomputation of all dependent computed fields

**Trade-offs:**
- Client-side evaluation means expressions can't use server-side data
- Expression syntax is limited to what the frontend evaluator supports

---

## ADR-013: Linked Field Auto-Population (`linked_field`)

**Date:** 2026-06-11
**Status:** Accepted

**Context:** When selecting a Product in an Order Item, the `unit_price` should auto-fill from the Product's `selling_price`. This is a general pattern for any Link field.

**Decision:** New `linked_field` property: `"{link_fieldname}.{source_fieldname}"`. When the Link field value changes, the frontend fetches the linked document and populates the target field.

**Rationale:**
- Config-driven — works for any Link field on any doctype
- User can override the auto-populated value
- Composes with `computed`: linked_field triggers → computed cascades
- Single property, simple syntax

---

## ADR-014: Path-Based Multi-Site Access (`/s/:site/`)

**Date:** 2026-06-12
**Status:** Accepted

**Context:** Multi-site routing via Host header requires DNS config or `/etc/hosts` entries. This adds friction for local development and testing.

**Decision:** Add path-based site access as a zero-config fallback. `/s/:site/workspace` routes to the named site. Host-based routing via `Host` header remains the primary mechanism for production. Both methods coexist.

**Rationale:**
- Zero configuration: `localhost:8000/s/airtime/workspace` works immediately
- No `/etc/hosts` entries needed for local development
- Host-based routing still works for production (cleaner URLs, SEO)
- Both methods share the same middleware chain and site context injection
- A `kora_site` cookie persists the site selection across requests for API calls

**Technical implementation:** A `NoRoute` handler intercepts `/s/:site/*` paths, looks up the site by name (with fuzzy matching — `airtime` matches `airtime.local`), injects site context, and serves or rewrites the request. For workspace paths it serves the SPA directly. For API paths it calls `router.HandleContext()` to re-dispatch.

**Trade-offs:**
- Path-based URLs are longer than host-based
- The `kora_site` cookie is needed for API calls to know which site context to use
- `HandleContext` re-runs the full middleware chain for API requests (slight overhead)

---

## ADR-015: Gin `NoRoute` for SPA Serving

**Date:** 2026-06-12
**Status:** Accepted

**Context:** The React SPA needs client-side routing — all paths under `/workspace` should serve `index.html`. Gin's radix tree forbids catch-all routes (`/workspace/*filepath`) alongside other routes at the same prefix level.

**Decision:** Use `router.NoRoute` to serve the SPA for `/workspace` and `/assets` paths, rather than registering explicit routes or middleware.

**Rationale:**
- Avoids Gin's radix tree conflicts between catch-all and exact routes
- Single handler for SPA serving, SPA fallback routing, asset serving, and path-based site access
- Middleware approach doesn't work — Gin matches routes before running middleware, and NoRoute fires after middleware with a 404 status that can't be overridden

**Trade-offs:**
- Only one `NoRoute` handler can be registered — all fallback logic must live in one function
- Any new "catch-all" behavior must be added to this handler

---

## ADR-016: Site-Aware Authentication (Per-Site DB for Sessions)

**Date:** 2026-06-12
**Status:** Accepted

**Context:** The original SessionManager was initialized with `firstDB` (the first site loaded) and used for all auth operations across all sites. This meant sessions created via `/s/fieldwork/api/auth/login` were stored in `airtime_db` (the first site), and session validation always checked `airtime_db`. Cross-site login/logout was broken.

**Decision:** All auth handlers (login, logout, /me, AuthMiddleware) read `site_db` from the Gin context and create a site-specific SessionManager on each request.

**Rationale:**
- Sessions are stored in the correct site's `_kora_session` table
- A session created on the Fieldwork site is only valid for Fieldwork
- No cross-site session leakage
- Consistent with the rest of the API (ORM reads `site_registry`, `site_db` from context)

**Trade-offs:**
- Creates a new SessionManager on each auth request (lightweight — just wraps `*sql.DB`)
- Login at `/s/airtime` with `admin@fieldwork.local` credentials fails (correct behavior — credentials are per-site)
