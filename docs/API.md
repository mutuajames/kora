# Kora — REST API Reference

## Authentication

All `/api/resource/*` and `/api/system/*` endpoints require authentication.

### Login

```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "admin@fieldwork.local",
  "password": "admin123"
}
```

Response:
```json
{
  "data": {
    "name": "Administrator",
    "email": "admin@fieldwork.local",
    "full_name": "Administrator",
    "roles": ["Administrator"]
  },
  "sid": "abc123..."
}
```

The session ID is set as a cookie (`kora_sid`) and also returned in the response body. For browser clients, the cookie is used automatically. For API clients, pass it as:

```http
Cookie: kora_sid=abc123...
```
or
```http
Authorization: Bearer abc123...
```

### CSRF Protection

State-changing requests (POST/PUT/DELETE) require a CSRF token. A token is automatically set as a cookie (`kora_csrf`) on your first GET request. Include it as a header:

```http
X-Kora-CSRF-Token: <token-value>
```

The header value must match the `kora_csrf` cookie value.

### Logout

```http
POST /api/auth/logout
Cookie: kora_sid=abc123...
```

### Get Current User

```http
GET /api/auth/me
Cookie: kora_sid=abc123...
```

---

## CRUD Endpoints

All endpoints follow the pattern `/api/resource/{DocType}`. The `{DocType}` name is case-sensitive and space-preserved (e.g., `/api/resource/Work%20Order`).

### List Documents

```http
GET /api/resource/Customer
```

**Query Parameters:**

| Parameter | Type | Default | Description |
|---|---|---|---|
| `limit` | int | 50 | Max results (max 500) |
| `offset` | int | 0 | Pagination offset |
| `order_by` | string | `modified DESC` | Sort column + direction |
| `fields` | JSON array | all | Fields to return: `["name","company_name"]` |
| `filters` | JSON array | none | Filter conditions |

**Filter format:** `[["field", "operator", value], ...]`

Supported operators: `=`, `!=`, `>`, `>=`, `<`, `<=`, `like`, `not like`, `in`, `not in`, `between`, `is`, `is not`

Example:
```http
GET /api/resource/Work%20Order?filters=[["status","in",["Approved","In Progress"]],["priority","=","High"]]&limit=25&offset=0
```

**Response:**
```json
{
  "data": [
    {
      "name": "CUST-0001",
      "company_name": "Acme Corp",
      "email": "info@acme.com",
      "doc_status": 0
    }
  ],
  "meta": {
    "doctype": "Customer",
    "total": 42
  }
}
```

### Get Document

```http
GET /api/resource/Customer/CUST-0001
```

**Response:**
```json
{
  "data": {
    "name": "CUST-0001",
    "company_name": "Acme Corp",
    "email": "info@acme.com",
    ...
  },
  "meta": {
    "doctype": "Customer"
  }
}
```

### Create Document

```http
POST /api/resource/Customer
Content-Type: application/json

{
  "company_name": "Acme Corp",
  "email": "info@acme.com",
  "phone": "555-0100",
  "city": "New York"
}
```

**Response:** `201 Created`
```json
{
  "data": {
    "name": "CUST-0002",
    "company_name": "Acme Corp",
    ...
  },
  "meta": {
    "doctype": "Customer"
  }
}
```

**With child table:**
```json
{
  "title": "Fix HVAC",
  "customer": "CUST-0001",
  "scheduled_date": "2026-07-15",
  "items": [
    {
      "equipment": "EQUI-0001",
      "description": "Annual maintenance",
      "estimated_hours": 2.0
    }
  ]
}
```

### Update Document

```http
PUT /api/resource/Customer/CUST-0001
Content-Type: application/json

{
  "phone": "555-0200",
  "city": "Boston"
}
```

Only the fields you send are updated. Read-only fields are silently ignored. Child tables are fully replaced (old rows deleted, new rows inserted).

**Response:** `200 OK`

### Delete Document

```http
DELETE /api/resource/Customer/CUST-0001
```

**Response:** `200 OK`
```json
{
  "data": {
    "message": "deleted"
  },
  "meta": {
    "doctype": "Customer"
  }
}
```

---

## Workflow Actions

```http
POST /api/resource/Work%20Order/WO-0001/workflow_action
Content-Type: application/json

{
  "action": "Submit for Approval"
}
```

**Response:** `200 OK` — document with updated status.
**Errors:** `400` — transition not available (wrong role, condition not met, missing required fields).

---

## System Schema API

### Get Doctype Schema

Returns the full DocType definition, workflow, permissions, and inbound references. The frontend form and list engines derive their structure from this response.

```http
GET /api/system/doctype/Order
```

**Response:**
```json
{
  "data": {
    "doctype": {
      "name": "Order",
      "module": "Airtime",
      "title_field": "title",
      "fields": [
        {"fieldname": "title", "fieldtype": "Data", "label": "Order Title", "reqd": true, ...},
        {"fieldname": "customer", "fieldtype": "Link", "label": "Customer", "options": "Customer", ...},
        {"fieldname": "items", "fieldtype": "Table", "label": "Items", "options": "Order Item", ...},
        {"fieldname": "subtotal", "fieldtype": "Currency", "computed": "SUM(items.line_total)", "read_only": true, ...},
        {"fieldname": "unit_price", "fieldtype": "Currency", "linked_field": "product.selling_price", ...}
      ]
    },
    "workflow": {
      "states": [{"state": "Draft", "doc_status": 0, "allow_edit": "Sales Agent", "style": "default"}],
      "transitions": [{"action": "Confirm Order", "from": "Draft", "to": "Confirmed", ...}],
      "state_field": "status"
    },
    "permissions": {"read": true, "write": true, "create": true, "delete": false, ...},
    "transitions": [{"action": "Confirm Order", "from": "Draft", ...}],
    "referenced_by": [
      {"doctype": "Service Report", "fieldname": "order", "label": "Order"}
    ]
  }
}
```

**Query Parameters:**

| Param | Purpose |
|---|---|
| `?state=Draft` | Return available transitions from this state for current user |

**`referenced_by` field:** Lists all doctypes that have Link fields pointing to this doctype. Used by the frontend to show "Related Documents" panels. E.g., viewing a Customer shows related Orders because `Order.customer → Customer`.

**`computed` and `linked_field`:** Fields may include `computed` (expression string like `"quantity * unit_price"`) or `linked_field` (like `"product.selling_price"`). The frontend uses these to auto-populate and auto-calculate field values. See CONFIG.md for details.

### Get Navigation

Returns sidebar structure and current user info.

```http
GET /api/system/navigation
```

**Response:**
```json
{
  "data": {
    "modules": [
      {
        "module": "Airtime",
        "label": "Airtime",
        "doctypes": [
          {"name": "Customer", "label": "Customer", "is_child": false},
          {"name": "Order", "label": "Order", "is_child": false}
        ]
      }
    ],
    "branding": {"app_name": "Kora", "primary_color": "#2563eb"},
    "user": {"name": "admin", "full_name": "Administrator", "email": "admin@...", "roles": ["Administrator"]}
  }
}
```

### Get Auth Providers

Public endpoint — no auth required. Returns enabled authentication methods.

```http
GET /api/auth/providers
```

**Response:**
```json
{
  "data": {
    "providers": [{"name": "password", "label": "Email & Password"}]
  }
}
```

---

## Validation Errors

### Field-level errors (single)

```json
{
  "error": {
    "type": "ValidationError",
    "message": "Full Name is required.",
    "field": "full_name",
    "doctype": "Customer"
  }
}
```

### Unique constraint errors

When a `unique: true` field has a duplicate value:

```json
{
  "error": {
    "type": "UniqueConstraint",
    "message": "ID Number must be unique. Value \"33333390\" already exists in CUST-0001.",
    "field": "id_number",
    "doctype": "Customer"
  }
}
```

The frontend displays this as an inline error on the specific field, with a red border and error text.

### Multiple validation errors
```json
{
  "error": {
    "errors": [
      {"type": "ValidationError", "message": "...", "field": "title"},
      {"type": "UniqueConstraint", "message": "...", "field": "email"}
    ]
  }
}
```

---

## System Config API

### List Config Versions

```http
GET /api/system/config/versions
```

### Get Config Version

```http
GET /api/system/config/versions/cv-fieldwork.local-1
```

### Diff Config Versions

```http
GET /api/system/config/diff?from=cv-fieldwork.local-1&to=cv-fieldwork.local-2
```

---

## Response Envelope

All success responses follow:

```json
{
  "data": { ... },
  "meta": {
    "config_version": 14,
    "doctype": "Customer",
    "total": 42
  }
}
```

All error responses follow:

```json
{
  "error": {
    "type": "ValidationError",
    "message": "Estimated hours must be at least 0.5.",
    "field": "estimated_hours",
    "doctype": "Work Order"
  }
}
```

Multiple validation errors:

```json
{
  "error": {
    "errors": [
      {"type": "ValidationError", "message": "...", "field": "title"},
      {"type": "ValidationError", "message": "...", "field": "customer"}
    ]
  }
}
```

---

## HTTP Status Codes

| Code | Meaning |
|---|---|
| 200 | OK (GET, PUT, DELETE) |
| 201 | Created (POST) |
| 204 | No Content (OPTIONS preflight) |
| 400 | Bad Request (invalid JSON, validation errors, workflow errors) |
| 401 | Unauthorized (missing or expired session) |
| 403 | Forbidden (permission denied, CSRF mismatch) |
| 404 | Not Found (missing DocType or document) |
| 429 | Too Many Requests (rate limit exceeded) |
| 500 | Internal Server Error (DB errors) |

---

## Security Headers (Every Response)

```
Content-Security-Policy: default-src 'self'; script-src ...; style-src ...
Referrer-Policy: same-origin
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-Request-Id: <uuid>
X-Xss-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000 (if TLS enabled)
```

## CORS Headers

```
Access-Control-Allow-Credentials: true
Access-Control-Allow-Origin: <configured-origin>
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Origin, Content-Type, Accept, Authorization, X-Kora-CSRF-Token, X-Request-Id
```

## Rate Limiting

Default: 100 requests/second per user. Returns `429 Too Many Requests` when exceeded. Configurable via `common_site_config.yaml`.
