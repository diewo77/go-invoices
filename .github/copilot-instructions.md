## Billing App – AI Assistant Working Rules

Concise project-specific guidance so an AI can contribute productively. Focus on THESE patterns; avoid generic boilerplate.

### 1. Architecture Snapshot

- Entry point: `cmd/server/main.go` builds an http server using `NewApp` (landing + authenticated pages) which wraps API routes from `internal/server/router.go`.
- Layers:
  - Handlers (`internal/handlers/*`): HTTP logic (auth/session, validation, request format dual JSON/Form).
  - Services (`internal/services/*`): business orchestration (ex: setup company) – keep side‑effect logic here when adding complex flows.
  - Models (`internal/models/*`): GORM entities. Soft delete now enabled on `Product` via `DeletedAt`.
  - View (`internal/view/view.go`): centralized template rendering (layout + partials + func map). Always use `view.Render(w,r,"file.html",data)`; DO NOT parse templates ad‑hoc in handlers.
  - Middleware (`internal/middleware/*`): inject user prefs (language/theme) + auth context.
  - DB bootstrap: `internal/db/migrate.go` (AutoMigrate fallback unless `MIGRATIONS=1`).
  - Static assets: `static/` (Tailwind+DaisyUI compiled to `tailwind.css`, hashed via manifest lookup in `view.resolveAsset`).

### 2. Key Conventions

- Templates: live in `templates/`; extend `layout.html` using `{{ define "title" }}` and `{{ define "content" }}`. Use existing partial `partials/header.html`.
- Functions available in templates: `t` (i18n), `lang`, `theme`, `avgPrice`, `mul`, `asset`, `year`.
- All UI routes must use `view.Render` (ensures func map + layout + caching). If a template includes `<!DOCTYPE` it renders standalone (skip layout detection logic already in `view.Render`).
- Dual-format endpoints pattern (example: `internal/handlers/setup.go`): Accept JSON (`application/json`) OR form (`application/x-www-form-urlencoded`). Use content negotiation: if `Accept` contains `application/json` w/out `text/html`, return JSON.
- Validation: use `internal/validation` helpers to populate a `Violations` map; return 400 JSON `{error:"validation_failed"}` or re-render template with `Errors` for HTML.
- Flash messages: written to a `flash` cookie then consumed & cleared in handlers (see dashboard/index handling in `cmd/server/app.go`).
- Product codes: each `Product` has `Code` unique per `UserID` (composite unique index). Codes are uppercased server-side.
- Soft delete: prefer `db.Where(...).Delete(&models.Product{})` so `DeletedAt` auto-populates. Queries must filter out deleted rows (already done in `ProductHandler.List` with `deleted_at IS NULL`). Follow that pattern for new list queries on soft-deleted models.

### 3. Routing & Auth

## Billing App – AI Assistant Working Rules

Concise project-specific guidance so an AI can contribute productively. Focus on THESE patterns; avoid generic boilerplate.

### 1. Architecture Snapshot

- Entry point: `cmd/server/main.go` builds an http server using `NewApp` (landing + authenticated pages) which wraps API routes from `internal/server/router.go`.
- Layers:
  - Handlers (`internal/handlers/*`): HTTP logic (auth/session, validation, request format dual JSON/Form).
  - Services (`internal/services/*`): business orchestration (ex: setup company) – keep side‑effect logic here when adding complex flows.
  - Models (`internal/models/*`): GORM entities. Soft delete now enabled on `Product` via `DeletedAt`.
  - View (`internal/view/view.go`): centralized template rendering (layout + partials + func map). Always use `view.Render(w,r,"file.html",data)`; DO NOT parse templates ad‑hoc in handlers.
  - Middleware (`internal/middleware/*`): inject user prefs (language/theme) + auth context.
  - DB bootstrap: `internal/db/migrate.go` (AutoMigrate fallback unless `MIGRATIONS=1`).
  - Static assets: `static/` (Tailwind+DaisyUI compiled to `tailwind.css`, hashed via manifest lookup in `view.resolveAsset`).

### 2. Key Conventions

- Templates: live in `templates/`; extend `layout.html` using `{{ define "title" }}` and `{{ define "content" }}`. Use existing partial `partials/header.html`.
- Functions available in templates: `t` (i18n), `lang`, `theme`, `avgPrice`, `mul`, `asset`, `year`.
- All UI routes must use `view.Render` (ensures func map + layout + caching). If a template includes `<!DOCTYPE` it renders standalone (skip layout detection logic already in `view.Render`).
- Dual-format endpoints pattern (example: `internal/handlers/setup.go`): Accept JSON (`application/json`) OR form (`application/x-www-form-urlencoded`). Use content negotiation: if `Accept` contains `application/json` w/out `text/html`, return JSON.
- Validation: use `internal/validation` helpers to populate a `Violations` map; return 400 JSON `{error:"validation_failed"}` or re-render template with `Errors` for HTML.
- Flash messages: written to a `flash` cookie then consumed & cleared in handlers (see dashboard/index handling in `cmd/server/app.go`).
- Product codes: each `Product` has `Code` unique per `UserID` (composite unique index). Codes are uppercased server-side.
- Soft delete: prefer `db.Where(...).Delete(&models.Product{})` so `DeletedAt` auto-populates. Queries must filter out deleted rows (already done in `ProductHandler.List` with `deleted_at IS NULL`). Follow that pattern for new list queries on soft-deleted models.

### 3. Routing & Auth

- Public pages: `/` (marketing), `/login`, `/signup`.
- Auth-required pages: `/dashboard`, `/products`, `/setup`, profile routes. Wrap new protected endpoints with `auth.Middleware` + `auth.RequireAuth` (see router usage).
- Use `/settings` -> redirects to `/setup` (alias). Maintain consistency for future renames.

### 4. Products Module (current reference)

- Handler: `internal/handlers/product.go` supports List (pagination + search), Create (JSON/form), Update, Delete (soft).
- Query params: `q` (case-insensitive substring against name or code), `page`, `limit` (<=200). Response JSON: `{ items, total, limit, offset }`.
- VAT handling: client can send `vat_rate` either 0–100 (percent) or decimal (>1 converted) then stored 0–1. Display uses `mul .VATRate 100`.
- Front-end editing: `templates/products.html` reuses a modal for create/update; code is immutable once created.

### 5. Setup Flow

- `SetupHandler` merges JSON + form logic. Form posts redirect with 303 on success. JSON POST returns 201 or 409 if already configured. Billing address auto-copies unless user chose separate fields.
- TVA (`vat_rate` / `tva_rate`) stored as decimal fraction. Form sends percent; JSON sends decimal.

### 6. Internationalization & Theme

- Lang + theme preferences pulled from cookies/query and injected via middleware. Use `t "key"` in templates; add new keys in `internal/i18n` and reference by code only (no inline strings for repeated phrases if translatable).
- Theme toggling script is embedded in `layout.html`; new theme-dependent components must rely on DaisyUI variables, not custom CSS.

### 7. Asset Pipeline

- Tailwind + DaisyUI compiled into `static/tailwind.css` (build script `build-assets.sh`). Asset links resolved through `asset` helper (manifest hashed filename in production, query hash fallback in dev). When adding a new static file, update `manifest.json` if hashed naming is used.

### 8. Database & Migrations

- Default dev path: AutoMigrate enumerating a curated slice in `migrate.go`. Add new models to that slice.
- Production / controlled schema: set `MIGRATIONS=1` and add versioned SQL under `migrations/`. If adding columns like soft delete manually in SQL mode, remember: `ALTER TABLE products ADD COLUMN deleted_at TIMESTAMPTZ NULL; CREATE INDEX idx_products_deleted_at ON products(deleted_at);`.
- Backfill utilities (e.g. product codes) live under `cmd/server/backfill_product_codes.go` and are triggered via a flag: `go run ./cmd/server -backfill-product-codes`.

### 9. Testing Guidelines (current state)

- Prefer running tests via Makefile inside Docker dev as needed.
- Always write tests for any new behavior or public API. Submissions without tests are incomplete.
- Tests (see product tests) rely on consistent template resolution. Always ensure new templates reside in `templates/` and avoid dynamic path building in tests.
- When writing tests for dual-format endpoints: set explicit `Accept` headers to disambiguate JSON vs HTML.

### 10. Adding New Features

- Prefer: model -> migration/backfill (if needed) -> service (if cross-entity logic) -> handler -> template -> asset update.
- Maintain JSON + form parity for user-facing create/update flows unless explicitly API-only.
- Reuse helper funcs (`avgPrice`, `mul`) by extending `view.Funcs` for any new generic computation exposed to templates.

### 11. Error & Response Patterns

- JSON errors: use `httpx.JSONError(w, status, code, details)` where `code` is a stable snake_case string (e.g. `validation_failed`, `code_already_exists`).
- HTML errors: re-render same template with `Errors` map or minimal fallback string only if template rendering fails.

### 12. Security & Auth Notes

- Session parsing utilities in `internal/auth` set user context. Always check `UserIDFromContext` for protected operations; respond 401 JSON or redirect to `/login` for HTML.
- When adding uniqueness constraints (like product code), catch duplicate insert via `strings.Contains(err.Error(),"duplicate")` (DB-agnostic fallback).

### 13. Performance / Caching

- Template caching enabled except in dev (`DEV=1`). Avoid manual global template caches; extend existing mechanism.
- Static assets: ETag computed for files; long immutable cache headers in non-dev. Use `asset` helper to bust caches after content changes.

### 14. Soft Delete Pattern

- Read queries must append `Where("deleted_at IS NULL")` (or rely on GORM's default if using model methods). Current code uses explicit SQL; mirror that style for consistency.
- REST semantics: soft delete returns 200 JSON `{deleted:id}` or 303 redirect for HTML form.

### 15. Common Pitfalls (Avoid)

- Parsing templates directly (duplicate func map, breaks layout). Always use `view.Render`.
- Forgetting percent↔decimal VAT conversion on input/edit.
- Returning 302 instead of 303 after form POST (use 303 for correctness with redirects to GET).
- Introducing inline Tailwind `@apply` in template `<style>` blocks (not processed at runtime) – use plain CSS.

### Quick Examples

Render template: `view.Render(w, r, "products.html", map[string]any{"Products": ps})`
JSON error: `httpx.JSONError(w, http.StatusBadRequest, "validation_failed", v)`
Pagination params: `limit` (<=200), `page` (1-based), `q` (alnum search on name/code)

---

### 16. Invoices & PDF flow (current state + how to extend)

- Models: `internal/models/invoice.go` defines `Invoice` (Status: draft/final) and `InvoiceItem` (Product + Quantity). Totals are computed by services, not stored.
- Expected handlers (mirror setup/products patterns):
  - `GET /invoices` list (HTML/JSON) with pagination and basic search (client name/ref when available).
  - `POST /invoices` create draft invoice with items (JSON/form). Validate product existence, positive quantities.
  - `POST /invoices/finalize?id=...` sets Status to `final` and blocks edits; idempotent.
  - `GET /invoices/pdf?id=...` streams a PDF.
- Service layer (`internal/services/invoice.go` recommended):
  - `ComputeTotals(inv *models.Invoice) (ht, tva, ttc float64)` using each item's `Product.UnitPrice` and `Product.VATRate`.
  - Prevent finalize if zero items or any product is soft-deleted.
- PDF generation (`internal/pdf/generator.go` recommended):
  - `func InvoicePDF(inv models.Invoice, company models.CompanySettings, client models.Client) ([]byte, error)` using Maroto.
  - Handler sets headers: `Content-Type: application/pdf` and `Content-Disposition: attachment; filename="invoice-<id>.pdf"`.

### 17. Testing examples (patterns to mirror)

- Dual-format endpoints:
  - JSON: `Accept: application/json`, `Content-Type: application/json` (e.g., POST /setup, POST /products).
  - HTML: `Accept: text/html`, `Content-Type: application/x-www-form-urlencoded`.
- Products pagination/search:
  - Seed >60 products; GET `/products?limit=50&page=2&q=sku` with JSON Accept; expect `items` length ≤50 and `total` ≥60.
- Unique product code per user:
  - POST two JSON creates with same `code` for the same authenticated user; expect 201 then 409 `code_already_exists`.
- Soft delete hides rows:
  - POST form to `/products/delete` with an id; then GET `/products` JSON and assert the id is absent; DB shows `deleted_at` set.
- Template rendering:
  - For HTML GETs, assert 200 and presence of layout elements rendered via `view.Render` (e.g., header nav links), not just raw page content.

Feedback welcome: point out unclear patterns or missing sections and we’ll refine this guide.

### 19. Docker & Makefile dev workflow

- Dev runs in containers. Use the provided Makefile targets:
  - `make docker-dev-up` to start the dev stack (app + Postgres) with live sync.
  - `make docker-dev-logs` to tail logs; `make docker-dev-down` to stop.
  - `make docker-dev-rebuild` (or `docker-dev-nocache`) to rebuild the dev image.
- Production-like stack: `make docker-up` / `make docker-down` with `docker-build` and `docker-logs` as needed.
- Local (non-docker) helpers also exist: `make dev` for reflex hot reload, `make run` to run the server, `make test` for tests, `make build` to compile.
- Migrations: use `make migrate-up`/`migrate-down` with DATABASE_DSN; enable runtime migrations by setting `MIGRATIONS=1` for the app.

### 18. Go learning mode (maintainer is new to Go)

- Be explicit and add brief comments for non-trivial code (why/how, not just what).
- Prefer clear, idiomatic Go over cleverness; avoid unnecessary generics/reflection; keep functions small and well-named.
- When introducing patterns (contexts, middleware, GORM usage, error handling), include 1–2 lines of rationale in code comments.
- Provide minimal "how to run" and "how to test" notes in PRs. Where helpful, include tiny Go tips inline (e.g., pointer receivers, slices vs arrays).
