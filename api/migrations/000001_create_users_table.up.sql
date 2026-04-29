CREATE TABLE users
(
	id          BIGSERIAL PRIMARY KEY,
	external_id TEXT        NOT NULL UNIQUE,
	name        TEXT        NOT NULL,
	email       TEXT        NOT NULL UNIQUE,
	enabled     BOOLEAN     NOT NULL        DEFAULT TRUE,
	created_at  TIMESTAMPTZ NOT NULL        DEFAULT NOW(),
	created_by  TEXT        NOT NULL,
	updated_at  TIMESTAMPTZ,
	updated_by  TEXT,
	disabled_at TIMESTAMPTZ,
	disabled_by TEXT
);
