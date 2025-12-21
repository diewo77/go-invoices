## Plan: Create /clients page

I need to create the `/clients` feature. Here's the detailed plan for what needs to be added:

### 1. Create [internal/handlers/client.go](billing-app/internal/handlers/client.go)

Create a `ClientHandler` following the same dual-format (JSON/HTML) pattern as `ProductHandler`:

- `List` — GET with pagination (`page`, `limit`) and search (`q`) on `nom`, `email`, `siret`
- `Create` — POST (JSON or form) with validation on `nom` (required)
- `Update` — POST/PUT/PATCH to `/clients/update?id=...`
- `Delete` — POST/DELETE to `/clients/delete?id=...`

Key imports: `auth`, `httpx`, `models`, `validation`, `view`, `gorm`

### 2. Create [templates/clients.html](billing-app/templates/clients.html)

Mirror the structure of [products.html](billing-app/templates/products.html):

- Header with title "Clients" and "Nouveau" button
- Stats cards (Total clients)
- Filter input for search
- Table with columns: Nom, Contact, Email, Téléphone, Actions (Edit/Delete)
- Modal dialog for create/edit form with fields: `nom`, `nom_commercial`, `contact`, `email`, `telephone`, `site_web`, `siren`, `siret`, `tva_intra`, address fields

### 3. Update [internal/server/router.go](billing-app/internal/server/router.go#L72-L89)

Add client routes after the invoice routes:

```go
// Client endpoints
ch := handlers.NewClientHandler(db)
mux.Handle("/clients", auth.Middleware(auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        ch.List(w, r)
    case http.MethodPost:
        ch.Create(w, r)
    default:
        w.Header().Set("Allow", "GET,POST")
        httpx.JSONError(w, http.StatusMethodNotAllowed, "method_not_allowed", nil)
    }
}))))
mux.Handle("/clients/delete", auth.Middleware(auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ch.Delete(w, r) }))))
mux.Handle("/clients/update", auth.Middleware(auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ch.Update(w, r) }))))
```

### 4. Update [templates/partials/header.html](billing-app/templates/partials/header.html#L22-L27)

Add clients link to navigation after products:

```html
<a href="/clients" class="nav-link">{{ t "nav_clients" }}</a>
```

### 5. Update [i18n/i18n.go](i18n/i18n.go)

Add translation keys for both `fr` and `en`:

```go
// French
"nav_clients": "Clients",
"clients.title": "Clients",
"clients.new": "Nouveau client",
"clients.empty": "Aucun client.",
"clients.name": "Nom / Raison sociale",
"clients.contact": "Contact",
"clients.email": "Email",
"clients.phone": "Téléphone",

// English
"nav_clients": "Clients",
"clients.title": "Clients",
"clients.new": "New client",
"clients.empty": "No clients yet.",
"clients.name": "Name / Company",
"clients.contact": "Contact",
"clients.email": "Email",
"clients.phone": "Phone",
```

### Further Considerations

1. **Address handling**: Should the form include inline address fields or use a separate address modal?
2. **SIRET/SIREN validation**: Add specific French business identifier validation rules?
3. **Delete behavior**: Hard delete or soft delete (add `DeletedAt` to Client model)?
