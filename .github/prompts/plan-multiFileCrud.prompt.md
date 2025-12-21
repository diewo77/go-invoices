## Plan: Multi-file CRUD Structure for Products & Invoices

I'll provide the complete file contents that need to be created. You'll need to:

1. Create the directories: `templates/products/` and `templates/invoices/`
2. Create the template files
3. Update the handlers and router

---

### **Products Module**

**Create `templates/products/index.html`:**

```html
{{ define "title" }}Produits - Billing App{{ end }} {{ define "content" }}
<section class="mx-auto max-w-7xl px-6 py-10 space-y-10">
  {{ template "page-header" (dict "Title" "Produits" "Subtitle" "Gestion du
  catalogue & tarifs" "BackLink" "/dashboard" "BackText" "Dashboard"
  "ActionText" (if (not .NoCompany) "Nouveau" "") "ActionLink" "/products/new")
  }} {{ template "errors-alert" . }} {{ if .NoCompany }}
  <div class="alert alert-warning">
    <span>Aucune société configurée.</span>
    <a href="/setup" class="btn btn-sm btn-outline ml-auto">Configurer</a>
  </div>
  {{ end }}

  <!-- Stats -->
  <div class="grid gap-4 sm:grid-cols-3">
    {{ template "stat-card" (dict "Label" "Total produits" "Value" (or (len
    .Products) 0)) }} {{ template "stat-card" (dict "Label" "Prix moyen" "Value"
    (if .Products (printf "%.2f€" (avgPrice .Products)) "—")) }} {{ template
    "stat-card" (dict "Label" "Dernière création" "Value" (if .Products ((index
    .Products 0).CreatedAt.Format "2006-01-02") "—")) }}
  </div>

  <!-- Table Card -->
  <div class="card bg-base-100 border border-base-200">
    <div
      class="p-4 flex flex-col gap-4 md:flex-row md:items-center md:justify-between"
    >
      <h2 class="text-lg font-semibold">Catalogue</h2>
      {{ template "search-filter" (dict "Action" "/products" "AriaLabel"
      "Filtrer les produits" "Query" .Query) }}
    </div>
    <div class="overflow-x-auto">
      <table class="table table-sm">
        <thead>
          <tr>
            <th>Nom / Code</th>
            <th>Type</th>
            <th>Unité</th>
            <th class="text-right">Prix HT (€)</th>
            <th class="text-right">TVA (%)</th>
            <th class="text-right">Actions</th>
          </tr>
        </thead>
        <tbody>
          {{ if .Products }}{{ range .Products }}
          <tr>
            <td>
              <a href="/products/{{ .ID }}" class="link link-hover font-medium"
                >{{ .Name }}</a
              >
              <span class="block text-[10px] tracking-wide opacity-60 font-mono"
                >{{ .Code }}</span
              >
            </td>
            <td>{{ if .ProductType }}{{ .ProductType.Name }}{{ end }}</td>
            <td>{{ if .UnitType }}{{ .UnitType.Symbol }}{{ end }}</td>
            <td class="text-right">{{ printf "%.2f" .UnitPrice }}</td>
            <td class="text-right">{{ printf "%.0f" (mul .VATRate 100) }}%</td>
            <td class="text-right">
              <div class="flex justify-end gap-2">
                <a href="/products/{{ .ID }}" class="btn btn-ghost btn-xs"
                  >Voir</a
                >
                <a href="/products/{{ .ID }}/edit" class="btn btn-ghost btn-xs"
                  >Éditer</a
                >
                <form
                  method="post"
                  action="/products/delete"
                  onsubmit="return confirm('Supprimer ce produit ?');"
                >
                  <input type="hidden" name="id" value="{{ .ID }}" />
                  <button type="submit" class="btn btn-error btn-xs">
                    Supprimer
                  </button>
                </form>
              </div>
            </td>
          </tr>
          {{ end }}{{ else }}
          <tr>
            <td colspan="6" class="text-center py-8 opacity-60 text-sm">
              Aucun produit
            </td>
          </tr>
          {{ end }}
        </tbody>
      </table>
    </div>
  </div>
</section>
{{ end }}
```

---

**Create `templates/products/form.html`:**

```html
{{ define "title" }}{{ if .Product.ID }}Modifier{{ else }}Nouveau{{ end }}
produit - Billing App{{ end }} {{ define "content" }}
<section class="mx-auto max-w-2xl px-6 py-10 space-y-8">
  {{ template "page-header" (dict "Title" (if .Product.ID "Modifier le produit"
  "Nouveau produit") "Subtitle" (if .Product.ID .Product.Name "Créer un nouveau
  produit") "BackLink" "/products" "BackText" "Retour") }} {{ template
  "errors-alert" . }}

  <div class="card bg-base-100 border border-base-200">
    <div class="card-body">
      <form
        method="POST"
        action="{{ if .Product.ID }}/products/update?id={{ .Product.ID }}{{ else }}/products{{ end }}"
        class="space-y-6"
      >
        {{ if .Product.ID }}<input
          type="hidden"
          name="id"
          value="{{ .Product.ID }}"
        />{{ end }}

        <h3 class="font-semibold text-lg border-b pb-2">
          Informations produit
        </h3>
        <div class="grid gap-4 sm:grid-cols-2">
          <label class="form-control sm:col-span-2">
            <span class="label-text">Nom *</span>
            <input
              name="name"
              type="text"
              class="input input-bordered"
              required
              value="{{ .Product.Name }}"
            />
            {{ if .Errors.name }}<span class="text-error text-xs mt-1"
              >{{ .Errors.name }}</span
            >{{ end }}
          </label>
          <label class="form-control">
            <span class="label-text">Code *</span>
            <input
              name="code"
              type="text"
              class="input input-bordered font-mono uppercase"
              maxlength="40"
              placeholder="SKU-001"
              value="{{ .Product.Code }}"
              {{
              if
              .Product.ID
              }}readonly
              class="input input-bordered font-mono uppercase bg-base-200"
              {{
              end
              }}
              {{
              if
              not
              .Product.ID
              }}required{{
              end
              }}
            />
            {{ if .Errors.code }}<span class="text-error text-xs mt-1"
              >{{ .Errors.code }}</span
            >{{ end }} {{ if .Product.ID }}<span class="text-xs opacity-60 mt-1"
              >Le code ne peut pas être modifié.</span
            >{{ end }}
          </label>
          <label class="form-control">
            <span class="label-text">Prix HT (€) *</span>
            <input name="unit_price" type="number" step="0.01" min="0"
            class="input input-bordered" required value="{{ if .Product.ID }}{{
            printf "%.2f" .Product.UnitPrice }}{{ end }}" /> {{ if
            .Errors.unit_price }}<span class="text-error text-xs mt-1"
              >{{ .Errors.unit_price }}</span
            >{{ end }}
          </label>
        </div>

        <div class="grid gap-4 sm:grid-cols-3">
          <label class="form-control">
            <span class="label-text">TVA (%)</span>
            <input name="vat_rate" type="number" step="0.01" min="0" max="100"
            class="input input-bordered" value="{{ if .Product.ID }}{{ printf
            "%.0f" (mul .Product.VATRate 100) }}{{ else }}20{{ end }}" />
            <span class="text-xs opacity-60 mt-1"
              >Pourcentage (ex: 20 pour 20%)</span
            >
          </label>
          <label class="form-control">
            <span class="label-text">Type de produit *</span>
            <select
              name="product_type_id"
              class="select select-bordered"
              required
            >
              <option value="">— Sélectionner —</option>
              {{ range .ProductTypes }}
              <option
                value="{{ .ID }}"
                {{
                if
                eq
                .ID
                $.Product.ProductTypeID
                }}selected{{
                end
                }}
              >
                {{ .Name }}
              </option>
              {{ end }}
            </select>
            {{ if .Errors.product_type_id }}<span
              class="text-error text-xs mt-1"
              >Type requis</span
            >{{ end }}
          </label>
          <label class="form-control">
            <span class="label-text">Unité de mesure *</span>
            <select name="unit_type_id" class="select select-bordered" required>
              <option value="">— Sélectionner —</option>
              {{ range .UnitTypes }}
              <option
                value="{{ .ID }}"
                {{
                if
                eq
                .ID
                $.Product.UnitTypeID
                }}selected{{
                end
                }}
              >
                {{ .Name }} ({{ .Symbol }})
              </option>
              {{ end }}
            </select>
            {{ if .Errors.unit_type_id }}<span class="text-error text-xs mt-1"
              >Unité requise</span
            >{{ end }}
          </label>
        </div>

        <div class="flex justify-end gap-3 pt-4">
          <a href="/products" class="btn btn-ghost">Annuler</a>
          <button type="submit" class="btn btn-primary">
            {{ if .Product.ID }}Enregistrer{{ else }}Créer{{ end }}
          </button>
        </div>
      </form>
    </div>
  </div>
</section>
{{ end }}
```

---

**Create `templates/products/show.html`:**

```html
{{ define "title" }}{{ .Product.Name }} - Billing App{{ end }} {{ define
"content" }}
<section class="mx-auto max-w-3xl px-6 py-10 space-y-8">
  {{ template "page-header" (dict "Title" .Product.Name "Subtitle" .Product.Code
  "BackLink" "/products" "BackText" "Produits" "ActionText" "Modifier"
  "ActionLink" (printf "/products/%d/edit" .Product.ID)) }}

  <div class="grid gap-6 md:grid-cols-2">
    <!-- Détails produit -->
    <div class="card bg-base-100 border border-base-200">
      <div class="card-body">
        <h2 class="card-title text-lg">Détails</h2>
        <dl class="grid grid-cols-2 gap-4 mt-4">
          <div>
            <dt class="text-xs uppercase tracking-wide opacity-60">Code</dt>
            <dd class="font-mono mt-1">{{ .Product.Code }}</dd>
          </div>
          <div>
            <dt class="text-xs uppercase tracking-wide opacity-60">Nom</dt>
            <dd class="font-medium mt-1">{{ .Product.Name }}</dd>
          </div>
          <div>
            <dt class="text-xs uppercase tracking-wide opacity-60">Type</dt>
            <dd class="mt-1">
              {{ if .Product.ProductType }}{{ .Product.ProductType.Name }}{{
              else }}—{{ end }}
            </dd>
          </div>
          <div>
            <dt class="text-xs uppercase tracking-wide opacity-60">Unité</dt>
            <dd class="mt-1">
              {{ if .Product.UnitType }}{{ .Product.UnitType.Name }} ({{
              .Product.UnitType.Symbol }}){{ else }}—{{ end }}
            </dd>
          </div>
        </dl>
      </div>
    </div>

    <!-- Tarification -->
    <div class="card bg-base-100 border border-base-200">
      <div class="card-body">
        <h2 class="card-title text-lg">Tarification</h2>
        <dl class="grid grid-cols-2 gap-4 mt-4">
          <div>
            <dt class="text-xs uppercase tracking-wide opacity-60">Prix HT</dt>
            <dd class="text-2xl font-bold mt-1">
              {{ printf "%.2f" .Product.UnitPrice }} €
            </dd>
          </div>
          <div>
            <dt class="text-xs uppercase tracking-wide opacity-60">TVA</dt>
            <dd class="text-2xl font-bold mt-1">
              {{ printf "%.0f" (mul .Product.VATRate 100) }}%
            </dd>
          </div>
          <div>
            <dt class="text-xs uppercase tracking-wide opacity-60">Prix TTC</dt>
            <dd class="text-lg font-semibold mt-1 text-primary">
              {{ printf "%.2f" (mul .Product.UnitPrice (add 1 .Product.VATRate))
              }} €
            </dd>
          </div>
          <div>
            <dt class="text-xs uppercase tracking-wide opacity-60">Devise</dt>
            <dd class="mt-1">{{ or .Product.Currency "EUR" }}</dd>
          </div>
        </dl>
      </div>
    </div>
  </div>

  <!-- Actions -->
  <div class="flex justify-between items-center pt-4">
    <form
      method="POST"
      action="/products/delete"
      onsubmit="return confirm('Êtes-vous sûr de vouloir supprimer ce produit ?')"
    >
      <input type="hidden" name="id" value="{{ .Product.ID }}" />
      <button type="submit" class="btn btn-error btn-outline">Supprimer</button>
    </form>
    <div class="flex gap-3">
      <a href="/products" class="btn btn-ghost">Retour</a>
      <a href="/products/{{ .Product.ID }}/edit" class="btn btn-primary"
        >Modifier</a
      >
    </div>
  </div>

  <!-- Métadonnées -->
  <div class="text-xs opacity-50 text-right">
    Créé le {{ .Product.CreatedAt.Format "02/01/2006 à 15:04" }} {{ if ne
    .Product.UpdatedAt .Product.CreatedAt }}• Modifié le {{
    .Product.UpdatedAt.Format "02/01/2006 à 15:04" }}{{ end }}
  </div>
</section>
{{ end }}
```

---

### **Invoices Module**

**Create `templates/invoices/index.html`:**

```html
{{ define "title" }}Factures - Billing App{{ end }} {{ define "content" }}
<section class="mx-auto max-w-7xl px-6 py-10 space-y-10">
  {{ template "page-header" (dict "Title" "Factures" "Subtitle" "Gestion des
  factures" "BackLink" "/dashboard" "BackText" "Dashboard" "ActionText" (if (not
  .NoCompany) "Nouvelle" "") "ActionLink" "/invoices/new") }} {{ template
  "errors-alert" . }} {{ if .NoCompany }}
  <div class="alert alert-warning">
    <span>Aucune société configurée.</span>
    <a href="/setup" class="btn btn-sm btn-outline ml-auto">Configurer</a>
  </div>
  {{ end }}

  <!-- Stats -->
  <div class="grid gap-4 sm:grid-cols-3">
    {{ template "stat-card" (dict "Label" "Total factures" "Value" (or (len
    .Invoices) 0)) }} {{ template "stat-card" (dict "Label" "Brouillons" "Value"
    .DraftCount) }} {{ template "stat-card" (dict "Label" "Finalisées" "Value"
    .FinalCount) }}
  </div>

  <!-- Table Card -->
  <div class="card bg-base-100 border border-base-200">
    <div
      class="p-4 flex flex-col gap-4 md:flex-row md:items-center md:justify-between"
    >
      <h2 class="text-lg font-semibold">Liste des factures</h2>
      {{ template "search-filter" (dict "Action" "/invoices" "AriaLabel"
      "Filtrer les factures" "Query" .Query) }}
    </div>
    <div class="overflow-x-auto">
      <table class="table">
        <thead>
          <tr>
            <th>#</th>
            <th>Client</th>
            <th>Date</th>
            <th>Statut</th>
            <th class="text-right">Articles</th>
            <th class="text-right">Actions</th>
          </tr>
        </thead>
        <tbody>
          {{ if .Invoices }}{{ range .Invoices }}
          <tr>
            <td class="font-mono font-medium">
              <a href="/invoices/{{ .ID }}" class="link link-hover"
                >#{{ .ID }}</a
              >
            </td>
            <td>{{ if .Client }}{{ .Client.Nom }}{{ else }}—{{ end }}</td>
            <td>{{ .CreatedAt.Format "02/01/2006" }}</td>
            <td>
              {{ if eq .Status "draft" }}
              <span class="badge badge-warning badge-sm">Brouillon</span>
              {{ else }}
              <span class="badge badge-success badge-sm">Finalisée</span>
              {{ end }}
            </td>
            <td class="text-right">{{ len .Items }}</td>
            <td class="text-right">
              <div class="flex justify-end gap-2">
                <a href="/invoices/{{ .ID }}" class="btn btn-ghost btn-xs"
                  >Voir</a
                >
                {{ if eq .Status "final" }}
                <a
                  href="/invoices/pdf?id={{ .ID }}"
                  class="btn btn-ghost btn-xs"
                  >PDF</a
                >
                {{ end }}
              </div>
            </td>
          </tr>
          {{ end }}{{ else }}
          <tr>
            <td colspan="6" class="text-center py-8 opacity-60 text-sm">
              Aucune facture
            </td>
          </tr>
          {{ end }}
        </tbody>
      </table>
    </div>
  </div>
</section>
{{ end }}
```

---

**Create `templates/invoices/form.html`:**

```html
{{ define "title" }}Nouvelle facture - Billing App{{ end }} {{ define "content"
}}
<section class="mx-auto max-w-3xl px-6 py-10 space-y-8">
  {{ template "page-header" (dict "Title" "Nouvelle facture" "Subtitle" "Créer
  une facture brouillon" "BackLink" "/invoices" "BackText" "Retour") }} {{
  template "errors-alert" . }}

  <div class="card bg-base-100 border border-base-200">
    <div class="card-body">
      <form method="POST" action="/invoices" class="space-y-6">
        <h3 class="font-semibold text-lg border-b pb-2">Client</h3>
        <label class="form-control">
          <span class="label-text">Sélectionner un client *</span>
          <select name="client_id" class="select select-bordered" required>
            <option value="">— Choisir un client —</option>
            {{ range .Clients }}
            <option value="{{ .ID }}">
              {{ .Nom }}{{ if .NomCommercial }} ({{ .NomCommercial }}){{ end }}
            </option>
            {{ end }}
          </select>
          {{ if .Errors.client_id }}<span class="text-error text-xs mt-1"
            >Client requis</span
          >{{ end }}
        </label>

        <h3 class="font-semibold text-lg border-b pb-2">Articles</h3>
        <div class="space-y-4" id="items-container">
          <div class="grid gap-4 sm:grid-cols-3 items-end">
            <label class="form-control sm:col-span-2">
              <span class="label-text">Produit *</span>
              <select name="product_id" class="select select-bordered" required>
                <option value="">— Sélectionner un produit —</option>
                {{ range .Products }}
                <option value="{{ .ID }}">
                  {{ .Name }} ({{ printf "%.2f" .UnitPrice }}€)
                </option>
                {{ end }}
              </select>
            </label>
            <label class="form-control">
              <span class="label-text">Quantité *</span>
              <input
                name="quantity"
                type="number"
                min="1"
                value="1"
                class="input input-bordered"
                required
              />
            </label>
          </div>
        </div>
        <p class="text-xs opacity-60">
          Pour ajouter plusieurs articles, créez la facture puis modifiez-la.
        </p>

        <div class="flex justify-end gap-3 pt-4">
          <a href="/invoices" class="btn btn-ghost">Annuler</a>
          <button type="submit" class="btn btn-primary">Créer brouillon</button>
        </div>
      </form>
    </div>
  </div>
</section>
{{ end }}
```

---

**Create `templates/invoices/show.html`:**

```html
{{ define "title" }}Facture #{{ .Invoice.ID }} - Billing App{{ end }} {{ define
"content" }}
<section class="mx-auto max-w-4xl px-6 py-10 space-y-8">
  {{ template "page-header" (dict "Title" (printf "Facture #%d" .Invoice.ID)
  "Subtitle" (if (eq .Invoice.Status "draft") "Brouillon" "Finalisée")
  "BackLink" "/invoices" "BackText" "Factures") }}

  <div class="grid gap-6 lg:grid-cols-3">
    <!-- Infos générales -->
    <div class="card bg-base-100 border border-base-200">
      <div class="card-body">
        <h2 class="card-title text-lg">Informations</h2>
        <dl class="space-y-3 mt-4">
          <div>
            <dt class="text-xs uppercase tracking-wide opacity-60">Statut</dt>
            <dd class="mt-1">
              {{ if eq .Invoice.Status "draft" }}
              <span class="badge badge-warning">Brouillon</span>
              {{ else }}
              <span class="badge badge-success">Finalisée</span>
              {{ end }}
            </dd>
          </div>
          <div>
            <dt class="text-xs uppercase tracking-wide opacity-60">Date</dt>
            <dd class="mt-1">{{ .Invoice.CreatedAt.Format "02/01/2006" }}</dd>
          </div>
          <div>
            <dt class="text-xs uppercase tracking-wide opacity-60">Client</dt>
            <dd class="mt-1 font-medium">
              {{ if .Client }}<a
                href="/clients/{{ .Client.ID }}"
                class="link link-primary"
                >{{ .Client.Nom }}</a
              >{{ else }}—{{ end }}
            </dd>
          </div>
        </dl>
      </div>
    </div>

    <!-- Totaux -->
    <div class="card bg-base-100 border border-base-200 lg:col-span-2">
      <div class="card-body">
        <h2 class="card-title text-lg">Totaux</h2>
        <div class="grid grid-cols-3 gap-4 mt-4">
          <div class="text-center p-4 bg-base-200 rounded-lg">
            <div class="text-xs uppercase opacity-60">Total HT</div>
            <div class="text-2xl font-bold">{{ printf "%.2f" .TotalHT }} €</div>
          </div>
          <div class="text-center p-4 bg-base-200 rounded-lg">
            <div class="text-xs uppercase opacity-60">TVA</div>
            <div class="text-2xl font-bold">
              {{ printf "%.2f" .TotalVAT }} €
            </div>
          </div>
          <div
            class="text-center p-4 bg-primary text-primary-content rounded-lg"
          >
            <div class="text-xs uppercase opacity-80">Total TTC</div>
            <div class="text-2xl font-bold">
              {{ printf "%.2f" .TotalTTC }} €
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- Articles -->
  <div class="card bg-base-100 border border-base-200">
    <div class="card-body">
      <h2 class="card-title text-lg">Articles</h2>
      <div class="overflow-x-auto mt-4">
        <table class="table table-sm">
          <thead>
            <tr>
              <th>Produit</th>
              <th class="text-right">Prix unitaire</th>
              <th class="text-right">Quantité</th>
              <th class="text-right">TVA</th>
              <th class="text-right">Total HT</th>
            </tr>
          </thead>
          <tbody>
            {{ range .Invoice.Items }}
            <tr>
              <td class="font-medium">
                {{ if .Product }}{{ .Product.Name }}{{ else }}Produit supprimé{{
                end }}
              </td>
              <td class="text-right">
                {{ if .Product }}{{ printf "%.2f" .Product.UnitPrice }} €{{ end
                }}
              </td>
              <td class="text-right">{{ .Quantity }}</td>
              <td class="text-right">
                {{ if .Product }}{{ printf "%.0f" (mul .Product.VATRate 100)
                }}%{{ end }}
              </td>
              <td class="text-right font-medium">
                {{ if .Product }}{{ printf "%.2f" (mul .Product.UnitPrice
                .Quantity) }} €{{ end }}
              </td>
            </tr>
            {{ else }}
            <tr>
              <td colspan="5" class="text-center opacity-60">Aucun article</td>
            </tr>
            {{ end }}
          </tbody>
        </table>
      </div>
    </div>
  </div>

  <!-- Actions -->
  <div class="flex justify-between items-center pt-4">
    <div>
      {{ if eq .Invoice.Status "draft" }}
      <form
        method="POST"
        action="/invoices/finalize?id={{ .Invoice.ID }}"
        class="inline"
      >
        <button
          type="submit"
          class="btn btn-success"
          onclick="return confirm('Finaliser cette facture ? Cette action est irréversible.')"
        >
          Finaliser
        </button>
      </form>
      {{ end }}
    </div>
    <div class="flex gap-3">
      <a href="/invoices" class="btn btn-ghost">Retour</a>
      {{ if eq .Invoice.Status "final" }}
      <a href="/invoices/pdf?id={{ .Invoice.ID }}" class="btn btn-primary"
        >Télécharger PDF</a
      >
      {{ end }}
    </div>
  </div>
</section>
{{ end }}
```

---

### **Handler Updates**

Add these methods to `internal/handlers/product.go` (after the `List` method):

```go
// Show displays a single product's details.
func (h *ProductHandler) Show(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	if id <= 0 {
		http.Redirect(w, r, "/products", http.StatusSeeOther)
		return
	}

	var product models.Product
	if err := h.DB.Preload("ProductType").Preload("UnitType").Where("id = ? AND deleted_at IS NULL", id).First(&product).Error; err != nil {
		http.Redirect(w, r, "/products", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := view.Render(w, r, "products/show.html", map[string]any{"Product": product}); err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
	}
}

// New renders the create product form.
func (h *ProductHandler) New(w http.ResponseWriter, r *http.Request) {
	var pts []models.ProductType
	var uts []models.UnitType
	_ = h.DB.Order("name asc").Find(&pts).Error
	_ = h.DB.Order("name asc").Find(&uts).Error

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := view.Render(w, r, "products/form.html", map[string]any{
		"Product":      models.Product{},
		"ProductTypes": pts,
		"UnitTypes":    uts,
	}); err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
	}
}

// Edit renders the edit product form.
func (h *ProductHandler) Edit(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	if id <= 0 {
		http.Redirect(w, r, "/products", http.StatusSeeOther)
		return
	}

	var product models.Product
	if err := h.DB.Preload("ProductType").Preload("UnitType").Where("id = ? AND deleted_at IS NULL", id).First(&product).Error; err != nil {
		http.Redirect(w, r, "/products", http.StatusSeeOther)
		return
	}

	var pts []models.ProductType
	var uts []models.UnitType
	_ = h.DB.Order("name asc").Find(&pts).Error
	_ = h.DB.Order("name asc").Find(&uts).Error

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := view.Render(w, r, "products/form.html", map[string]any{
		"Product":      product,
		"ProductTypes": pts,
		"UnitTypes":    uts,
	}); err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
	}
}
```

Add these methods to `internal/handlers/invoice.go`:

```go
// Show displays a single invoice's details.
func (h *InvoiceHandler) Show(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	if id <= 0 {
		http.Redirect(w, r, "/invoices", http.StatusSeeOther)
		return
	}

	var inv models.Invoice
	if err := h.DB.Preload("Items.Product").First(&inv, id).Error; err != nil {
		http.Redirect(w, r, "/invoices", http.StatusSeeOther)
		return
	}

	var client models.Client
	_ = h.DB.First(&client, inv.ClientID).Error

	ht, vat, ttc := h.Svc.ComputeTotals(&inv)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := view.Render(w, r, "invoices/show.html", map[string]any{
		"Invoice":  inv,
		"Client":   client,
		"TotalHT":  ht,
		"TotalVAT": vat,
		"TotalTTC": ttc,
	}); err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
	}
}

// New renders the create invoice form.
func (h *InvoiceHandler) New(w http.ResponseWriter, r *http.Request) {
	uid, _ := auth.UserIDFromContext(r.Context())

	var clients []models.Client
	_ = h.DB.Where("user_id = ?", uid).Order("nom asc").Find(&clients).Error

	var company models.CompanySettings
	if err := h.DB.Select("id").First(&company).Error; err != nil {
		_ = view.Render(w, r, "invoices/form.html", map[string]any{"NoCompany": true})
		return
	}

	var products []models.Product
	_ = h.DB.Where("company_id = ? AND deleted_at IS NULL", company.ID).Order("name asc").Find(&products).Error

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := view.Render(w, r, "invoices/form.html", map[string]any{
		"Clients":  clients,
		"Products": products,
	}); err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
	}
}
```

---

### **Router Updates**

Update the products routes in `router.go`:

```go
// Product endpoints
ph := handlers.NewProductHandler(db)
mux.Handle("/products", auth.Middleware(auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		ph.List(w, r)
		return
	}
	if r.Method == http.MethodPost {
		ph.Create(w, r)
		return
	}
	w.Header().Set("Allow", "GET,POST")
	httpx.JSONError(w, http.StatusMethodNotAllowed, "method_not_allowed", nil)
}))))
mux.Handle("/products/new", auth.Middleware(auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ph.New(w, r) }))))
mux.Handle("/products/show", auth.Middleware(auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ph.Show(w, r) }))))
mux.Handle("/products/edit", auth.Middleware(auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ph.Edit(w, r) }))))
mux.Handle("/products/delete", auth.Middleware(auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ph.Delete(w, r) }))))
mux.Handle("/products/update", auth.Middleware(auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ph.Update(w, r) }))))

// Friendly URL: /products/{id} and /products/{id}/edit
mux.Handle("/products/", auth.Middleware(auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/products/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Redirect(w, r, "/products", http.StatusSeeOther)
		return
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil || id <= 0 {
		http.Redirect(w, r, "/products", http.StatusSeeOther)
		return
	}
	q := r.URL.Query()
	q.Set("id", strconv.Itoa(id))
	r.URL.RawQuery = q.Encode()

	if len(parts) >= 2 && parts[1] == "edit" {
		ph.Edit(w, r)
		return
	}
	ph.Show(w, r)
}))))
```

Add invoice routes:

```go
// Invoice endpoints
mux.Handle("/invoices/new", auth.Middleware(auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ih.New(w, r) }))))
mux.Handle("/invoices/show", auth.Middleware(auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ih.Show(w, r) }))))

// Friendly URL: /invoices/{id}
mux.Handle("/invoices/", auth.Middleware(auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/invoices/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Redirect(w, r, "/invoices", http.StatusSeeOther)
		return
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil || id <= 0 {
		http.Redirect(w, r, "/invoices", http.StatusSeeOther)
		return
	}
	q := r.URL.Query()
	q.Set("id", strconv.Itoa(id))
	r.URL.RawQuery = q.Encode()
	ih.Show(w, r)
}))))
```

---

### **Update template paths in existing handlers**

Change `"products.html"` → `"products/index.html"` and `"invoices.html"` → `"invoices/index.html"` in the List methods.

---

### **Add `add` template function**

In `view/view.go`, add to the FuncMap:

```go
"add": func(a, b float64) float64 { return a + b },
```

---

### **Delete old files**

After creating the new structure:

- Delete `templates/products.html`
- Delete `templates/invoices.html`
- Delete `static/js/products.js` (no longer needed)
