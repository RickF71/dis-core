-- 000_init.sql
-- Base DIS-Core structures: domains, bootstrap_files, canon.

CREATE TABLE IF NOT EXISTS domains (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    parent_id   TEXT,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ,
    CONSTRAINT fk_domains_parent
        FOREIGN KEY (parent_id) REFERENCES domains(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS bootstrap_files (
    id          SERIAL PRIMARY KEY,
    domain_id   TEXT,
    path        TEXT NOT NULL,
    content     TEXT,
    created_at  TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS canon (
    id          SERIAL PRIMARY KEY,
    domain_id   TEXT,
    path        TEXT NOT NULL,
    content     TEXT,
    created_at  TIMESTAMPTZ DEFAULT now()
);
