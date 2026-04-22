CREATE TABLE users
(
	id          BIGSERIAL PRIMARY KEY,
	external_id TEXT        NOT NULL UNIQUE,
	name        TEXT        NOT NULL,
	email       TEXT        NOT NULL UNIQUE,
	created_at  TIMESTAMPTZ NOT NULL        DEFAULT NOW(),
	created_by  TEXT        NOT NULL,
	enabled     BOOLEAN     NOT NULL        DEFAULT TRUE,
	updated_at  TIMESTAMPTZ,
	disabled_at TIMESTAMPTZ,
	disabled_by TEXT
);
