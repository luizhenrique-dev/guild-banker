CREATE TABLE user_account
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

CREATE TABLE guild
(
	id           BIGSERIAL PRIMARY KEY,
	name         TEXT        NOT NULL UNIQUE,
	display_name TEXT        NOT NULL,
	enabled      BOOLEAN     NOT NULL        DEFAULT TRUE,
	created_at   TIMESTAMPTZ NOT NULL        DEFAULT NOW(),
	created_by   TEXT        NOT NULL,
	updated_at   TIMESTAMPTZ,
	updated_by   TEXT,
	disabled_at  TIMESTAMPTZ,
	disabled_by  TEXT
);

CREATE TABLE guild_member
(
	guild_id   BIGINT      NOT NULL REFERENCES guild (id) ON DELETE CASCADE,
	user_id    BIGINT      NOT NULL REFERENCES user_account (id) ON DELETE CASCADE,
	invited_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	invited_by BIGINT      NOT NULL REFERENCES user_account (id),
	PRIMARY KEY (guild_id, user_id)
);

CREATE UNIQUE INDEX idx_guild_name_lower ON guild (LOWER(name));
