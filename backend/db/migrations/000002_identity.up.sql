CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE organisations (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name       text NOT NULL,
    -- The slug is the organisation's subdomain, so it must be a valid, non-reserved hostname label.
    slug       text NOT NULL UNIQUE
               CHECK (slug ~ '^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$'
                      AND slug NOT IN ('www', 'api', 'admin', 'app')),
    kind       text NOT NULL
               CHECK (kind IN ('producer', 'processor', 'transporter', 'exporter', 'importer', 'certifier', 'regulator')),
    status     text NOT NULL DEFAULT 'pending'
               CHECK (status IN ('pending', 'verified', 'rejected')),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE users (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email         citext NOT NULL UNIQUE,
    password_hash text NOT NULL,
    name          text NOT NULL,
    status        text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'disabled')),
    created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE roles (
    id   int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name text NOT NULL UNIQUE
);

-- Permission names are "resource:action".
CREATE TABLE permissions (
    id   int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name text NOT NULL UNIQUE CHECK (name ~ '^[a-z_]+:[a-z_]+$')
);

CREATE TABLE role_permissions (
    role_id       int NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    permission_id int NOT NULL REFERENCES permissions (id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- "This user has this role in this organisation."
-- A NULL organisation is a platform-wide scope (superadmin).
CREATE TABLE userscope (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role_id         int NOT NULL REFERENCES roles (id),
    organisation_id uuid REFERENCES organisations (id) ON DELETE CASCADE,
    created_at      timestamptz NOT NULL DEFAULT now(),
    UNIQUE NULLS NOT DISTINCT (user_id, organisation_id)
);

INSERT INTO roles (name) VALUES ('superadmin'), ('org_admin'), ('member');

INSERT INTO permissions (name) VALUES
    ('org:verify'), ('org:read'), ('user:manage'),
    ('passport:issue'), ('passport:read'),
    ('handover:create'), ('handover:accept'),
    ('transform:create'),
    ('certificate:issue'), ('certificate:revoke'),
    ('flag:create'), ('flag:resolve'),
    ('template:manage');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON
    r.name = 'superadmin'
    OR (r.name = 'org_admin' AND p.name NOT IN ('org:verify', 'template:manage', 'flag:resolve'))
    OR (r.name = 'member' AND p.name IN ('passport:issue', 'passport:read', 'handover:create',
                                         'handover:accept', 'transform:create', 'flag:create'));
