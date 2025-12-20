CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

CREATE TABLE addresses (
    id SERIAL PRIMARY KEY,
    ligne1 TEXT NOT NULL,
    ligne2 TEXT,
    code_postal TEXT NOT NULL,
    ville TEXT NOT NULL,
    pays TEXT NOT NULL,
    type TEXT,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    nom TEXT,
    prenom TEXT,
    address_id INTEGER REFERENCES addresses(id),
    role_id INTEGER REFERENCES roles(id),
    permissions TEXT,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

CREATE TABLE company_settings (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    raison_sociale TEXT NOT NULL,
    nom_commercial TEXT NOT NULL,
    siren TEXT NOT NULL,
    siret TEXT NOT NULL,
    code_naf TEXT NOT NULL,
    tva DOUBLE PRECISION,
    rcs TEXT,
    greffe TEXT,
    rm TEXT,
    dept_rm TEXT,
    capital DOUBLE PRECISION DEFAULT 0,
    activite_principale TEXT,
    agrement_sap BOOLEAN NOT NULL DEFAULT false,
    date_creation TIMESTAMP,
    type_imposition TEXT NOT NULL,
    type_declarant TEXT NOT NULL DEFAULT 'Déclarant 1',
    frequence_urssaf TEXT NOT NULL,
    redevable_tva BOOLEAN NOT NULL DEFAULT false,
    date_premiere_declaration TIMESTAMP,
    forme_juridique TEXT NOT NULL,
    regime_fiscal TEXT NOT NULL,
    address_id INTEGER REFERENCES addresses(id),
    billing_address_id INTEGER REFERENCES addresses(id),
    telephone TEXT,
    email TEXT,
    site_web TEXT,
    iban TEXT,
    logo_url TEXT,
    mentions_legales TEXT,
    rc_pro TEXT,
    tva_intra TEXT,
    code_ape TEXT,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);
