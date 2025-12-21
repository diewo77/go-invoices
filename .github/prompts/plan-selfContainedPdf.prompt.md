Self-contained PDF renderer plan (Maroto) + signing hook

Goals

- Replace stubbed pdf generator with an in-process renderer (Maroto) producing real invoice PDFs.
- Support optional logo embedding (local path or HTTP/HTTPS URL fetch + register image reader).
- Add a post-processing hook to allow later signing (since pdfcpu lacks a public signing API); document options.

Decisions

- Renderer: github.com/go-pdf/fpdf (pure Go, no CGO). Version: v0.8.0 (or latest stable).
- Signing: pdfcpu currently has validation only; no public Sign API. Provide a PostProcess func([]byte) ([]byte, error) hook so a caller can inject an external signer (or swap to UniPDF if we want built-in signing later).

Data contracts (unchanged)

- Keep InvoiceData, InvoiceItem, ClientData, CompanyData as-is, but add fields:
  - CompanyData.LogoURL (already present) used for optional embedding.
  - InvoiceData.PostProcess func([]byte) ([]byte, error) `json:"-"` to avoid JSON leaks (optional, defaults to nil).

Renderer behavior

- Validate: must have at least 1 item; error otherwise.
- Defaults: if InvoiceNumber/Date/DueDate empty, use "N/A".
- Layout (A4 portrait, margins 15mm):
  - Header: company name/address, invoice title/number, date/due date.
  - Client block: name/address/email.
  - Optional logo: if LogoURL set and reachable (file:// or http/https), fetch/read into RegisterImageOptionsReader; place top-right (e.g., x=170mm, width=25mm).
  - Items table: columns Description (90), Qty (20), Unit Price (40), Total (40). Compute line total if Total==0.
  - Totals: Subtotal (HT, recomputed if Total==0), VAT, Grand Total (fallback Total+VAT if GrandTotal==0). Currency format simple "€ %.2f".
- Output: buffer bytes. If PostProcess != nil, call it on the rendered bytes; return its result.

Signing path (future)

- Since pdfcpu cannot sign today, users can:
  - Inject a PostProcess that calls an external signer (CLI/service) or a custom PKCS7 signer.
  - Or switch to UniPDF/Unidoc for native signing.

Tests

- Add a pdf module test that asserts: non-empty bytes; starts with %PDF; contains expected text chunks (Invoice, company/client names) via simple substring; optional skip logo fetch in unit test.
- Existing app tests (invoice finalize/PDF) should pass unchanged.

Integration

- No handler changes needed; they already build InvoiceData. Optionally populate LogoURL if available; leave PostProcess nil for now.

Tasks to implement (when editing)

- Update pdf/go.mod to require github.com/johnfercher/maroto/v2 and add go.sum via go mod tidy.
- Replace pdf/generator.go with the Maroto implementation + PostProcess hook.
- Add unit test in pdf module.
- Run go test ./... in billing-app after tidy.
