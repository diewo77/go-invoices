# Plan d'extraction des packages

Voici les candidats idéaux pour être extraits en **packages indépendants** (modules réutilisables), classés du plus facile au plus complexe :

## État d'avancement

- [x] **`pdf`** (Génération de PDF)
- [x] **`httpx`** (Helpers HTTP)
- [x] **`i18n`** (Traduction)
- [x] **`validation`** (Validation)
- [x] **`auth`** (Authentification)
- [x] **`view`** (Moteur de rendu)

---

### 1. Les "Outils Génériques" (Faciles à extraire)

Ces fonctionnalités ne dépendent pas de votre "business" (produits, factures) et pourraient servir dans n'importe quel autre projet Go.

- **`internal/pdf`** (Génération de PDF)

  - **Pourquoi ?** La logique de dessin de PDF (gofpdf ou autre) est technique. Vous pouvez en faire un package qui prend des données brutes (structs simples) et sort des bytes, sans rien savoir de votre base de données.
  - **Nom suggéré :** `github.com/diewo77/billing-app/pkg/pdf`

- **`internal/httpx`** (Helpers HTTP)

  - **Pourquoi ?** Vos fonctions pour répondre en JSON (`JSONError`, `RenderJSON`) ou gérer les headers sont purement utilitaires.
  - **Nom suggéré :** `github.com/diewo77/billing-app/pkg/httpx`

- **`internal/i18n`** (Traduction)

  - **Pourquoi ?** Le système de chargement des fichiers de langue et la fonction `t("key")` sont universels.
  - **Nom suggéré :** `github.com/diewo77/billing-app/pkg/i18n`

- **`internal/validation`** (Validation)
  - **Pourquoi ?** Vérifier si un email est valide ou si un champ est vide est un besoin partout.
  - **Nom suggéré :** `github.com/diewo77/billing-app/pkg/validation`

### 2. Les "Modules Fonctionnels" (Intermédiaires)

Ces parties contiennent de la logique métier mais peuvent être isolées si on définit bien les frontières.

- **`internal/auth`** (Authentification)

  - **Ce qu'on peut sortir :** La gestion des sessions, le hachage de mot de passe, et les middlewares de vérification.
  - **Attention :** Il faut souvent le découpler du modèle `User` spécifique (en utilisant une interface `UserProvider` par exemple).
  - **Nom suggéré :** `github.com/diewo77/billing-app/pkg/auth`

- **`internal/view`** (Moteur de rendu)
  - **Ce qu'on peut sortir :** La logique qui charge les templates HTML, gère le cache et les `FuncMap`.
  - **Nom suggéré :** `github.com/diewo77/billing-app/pkg/view`

### Ce qu'il vaut mieux garder dans l'application principale (`internal/`)

- **`internal/handlers`** : C'est le "ciment" qui lie tout ensemble (routes HTTP spécifiques à votre app).
- **`internal/services`** (sauf si très générique) : La logique "Créer une facture" est le cœur de votre app `billing-app`.
- **`internal/models`** : Vos structures de données (GORM).
