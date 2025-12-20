# Présentation du projet “Billing App”

## Contexte et objectifs  
Billing App est une application web **from scratch**, développée en Go 1.24.5, qui permet de gérer la facturation d’une petite entreprise : catalogue de produits, création de factures en brouillon, calcul automatique de la TVA, validation finale et génération de PDF. L’ensemble de la solution est packagé dans des conteneurs Docker (Go + PostgreSQL) pour un déploiement et un développement simplifiés.

## Fonctionnalités principales  
- **Catalogue de produits** : définition du nom, du prix unitaire et du taux de TVA de chaque produit.  
- **Gestion des factures** :  
  - Création de factures en **statut “draft”** (brouillon).  
  - Ajout/suppression de lignes (produit + quantité).  
  - Calcul ligne par ligne : montant HT, TVA, TTC.  
  - Passage en statut **“final”** pour bloquer la facture.  
- **Export PDF** : génération d’une facture au format PDF via Maroto (rendu Go, sans dépendances natives).  
- **Interface web** : rendu serveur avec `html/template` et `net/http`, sans framework front-end externe.  
- **Containerisation** : orchestration via Docker Compose :  
  - Service **db** (PostgreSQL 15) avec healthcheck et données persistées.  
  - Service **app** (binaire Go statiquement compilé).

### API Produits – Liste et pagination (JSON)

`GET /products` avec `Accept: application/json` retourne la forme suivante :

```
{
  "items": [
    { "id": 1, "name": "Nom", "code": "SKU1", "unit_price": 12.5, "vat_rate": 0.2, ... },
    ...
  ],
  "total": 123,        // nombre total d'éléments correspondant (hors pagination)
  "limit": 50,         // taille de page (max 200)
  "offset": 0          // offset calculé depuis page (page 1 => offset 0)
}
```

Paramètres de requête pris en charge :

- `q` : filtre texte (insensible à la casse) sur `name` et `code`.
- `page` : numéro de page (1 par défaut).
- `limit` : taille de page (50 par défaut, maximum 200).

Notes :

- Les produits supprimés (soft delete) sont exclus des résultats.
- Les codes produit sont uniques par utilisateur (composite `(user_id, code)`).

### Modélisation: types de produits et unités de mesure (1→N)

Tables de référence (1) liées aux produits (N) :

- `product_types` (ex: Vente de marchandises, Prestation de services, Abonnement)
- `unit_types` (ex: pièce «pc», heure «h», kilogramme «kg», mètre «m»)

Relation 1→N:

- `products.product_type_id` → `product_types.id`
- `products.unit_type_id` → `unit_types.id`

En JSON/HTML, un produit référence son type et son unité par identifiant. En base, des index et clés étrangères garantissent l’intégrité; la suppression d’un type ne supprime pas les produits (FK en `ON DELETE SET NULL`).

## Nouveautés récentes


### Linting and CI

- Linter: golangci-lint configured via `.golangci.yml`.
- Local: install with `scripts/install-golangci-lint.sh` and run `make lint RUN_LOCAL=1` to see issues without failing.
- CI: GitHub Actions runs build, tests, and lints on pushes/PRs to `main`.

billing-app/
├─ cmd/server/main.go         # point d’entrée : lecture de la config, routes, ServeMux
├─ internal/
│  ├─ db/migrate.go           # connexion Postgres + AutoMigrate (GORM) avec retry
│  ├─ models/models.go        # définitions GORM : Product, Invoice, InvoiceItem
│  ├─ handlers/invoice.go     # handlers HTTP (ListProducts, Create/Update/Validate, PDF)
│  ├─ pdf/generator.go        # logique de génération PDF (Maroto v2)
│  └─ templates/embed.go      # //go:embed pour HTML templates
├─ templates/                 # fichiers HTML (layout, form, liste…)
├─ static/                    # CSS, JS et images (servis par FileServer)
├─ Dockerfile                 # multi-stage build pour compiler le serveur Go
├─ docker-compose.yml         # orchestration app + db
├─ go.mod, go.sum             # définition du module Go
└─ README.md                  # documentation et instructions d’installation

## Stack technique  
- **Langage** : Go 1.24.5 (modules, embed)  
- **BDD** : PostgreSQL 15 (containerisé)  
- **ORM** : GORM  
- **PDF** : Maroto (v2)  
- **Templates** : `html/template` + `net/http`  
- **Orchestration** : Docker Compose (healthcheck, dépendances)  

## Endpoint d'initialisation (Setup) – Mode double JSON / Formulaire

L'endpoint `/setup` permet de créer la configuration initiale de l'entreprise.

Modes supportés :

- **JSON API** :
  - `GET /setup` retourne `{ "configured": bool }`.
  - `HEAD /setup` retourne seulement l'en-tête `X-Setup-Configured: true|false`.
  - `POST /setup` (Content-Type: `application/json`) crée la configuration et répond `201` + objet `CompanySettings`.
  - Si déjà configuré, `POST` renvoie `409` et un JSON d'erreur.

- **Formulaire HTML** :
  - `POST /setup` (Content-Type: `application/x-www-form-urlencoded`) accepte les champs `company,address,address2,postal_code,city,country,siret,tva(oui/non),tva_rate(%)`.
  - Succès : redirection `303 See Other` vers `/`.
  - Conflit (déjà configuré) : redirection `303` vers `/`.

Conversion TVA : `tva_rate` (%) envoyé dans le formulaire est converti en ratio décimal côté serveur (ex: 20 devient 0.20). En JSON, envoyer directement la valeur décimale dans `vat_rate`.

Documentation OpenAPI : `/openapi.yaml` expose le schéma incluant les deux modes (JSON + form-data). 

Avantages :
- Un seul handler pour interface web et clients API.
- HEAD rapide pour health-check ou front minimal.
- Compatibilité future avec clients SPA ou CLI.

Bonnes pratiques :
- Après initialisation, sécuriser ultérieurement `/setup` (ex: le désactiver ou ajouter une auth) pour éviter reconfiguration indésirable.
- Sur front JS, préférer `GET /setup` ou `HEAD /setup` pour déterminer si le formulaire doit être affiché.

## Mode développement avec rechargement automatique (Docker Compose Watch)

Pour un cycle de développement rapide sans reconstruire l'image à chaque sauvegarde :

1. Un stage `dev` a été ajouté au `Dockerfile` avec l'outil `reflex` (watcher Go).
2. Un fichier `docker-compose.dev.yml` configure `develop.watch` pour :
  - synchroniser les fichiers modifiés instantanément dans le conteneur (`action: sync`),
  - déclencher un rebuild uniquement quand `go.mod` ou `go.sum` changent.

Lancer l'environnement dev :

```bash
docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
```

Activer le seeding en développement:
```bash
DB_SEED=1 docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
```

Activer les migrations SQL versionnées (recommandé en prod):

```bash
MIGRATIONS=1 docker compose up --build

En dev via Docker Compose Watch:

```bash
MIGRATIONS=1 docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
```

Notes migrations:
- La variable `MIGRATIONS=1` déclenche l'exécution des scripts dans `./migrations` au démarrage de l'app (et dans `dev.sh`, un run initial `--migrate-only` aide à échouer tôt en cas d'erreur).
- L'image finale embarque le dossier `migrations/` (copié dans le Dockerfile), donc l'exécution est possible également en production.
- En développement, vous pouvez laisser `MIGRATIONS` non défini pour conserver le fallback `AutoMigrate`.
```

En local, si vous souhaitez regénérer le schéma rapidement sans versionnage strict, laissez `MIGRATIONS` vide et GORM AutoMigrate fera le minimum. Pour un changement irréversible de structure ou en équipe, utilisez toujours `MIGRATIONS=1`.

Arrêter :

```bash
docker compose -f docker-compose.yml -f docker-compose.dev.yml down
```

Travaillez normalement dans l'éditeur : à chaque modification `.go`, `reflex` redémarre le serveur quasi instantanément.

Notes :
- En production on utilise toujours l'image finale minimaliste (stage Alpine), pas le stage `dev`.
- Si vous ajoutez des variables d'environnement, mettez-les dans les deux fichiers compose si nécessaires.
- Exclure les gros dossiers (vendor, node_modules…) pour accélérer la synchronisation.

---

> Vous pouvez coller ce texte dans votre README ou dans un commentaire de projet VS Code, puis utiliser GitHub Copilot pour générer les handlers, les modèles et les templates à partir de cette structure !  