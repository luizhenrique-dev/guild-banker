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

CREATE TABLE fixed_expense
(
	id         BIGSERIAL PRIMARY KEY,
	user_id    BIGINT         NOT NULL REFERENCES user_account (id) ON DELETE CASCADE,
	name       TEXT           NOT NULL,
	amount     NUMERIC(12, 2) NOT NULL,
	due_day    INTEGER        NOT NULL,
	category   TEXT           NOT NULL,
	status     TEXT           NOT NULL DEFAULT 'ACTIVE',
	created_at TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
	created_by TEXT           NOT NULL,
	updated_at TIMESTAMPTZ,
	updated_by TEXT,
	CONSTRAINT chk_fixed_expense_due_day CHECK (due_day BETWEEN 1 AND 31),
	CONSTRAINT chk_fixed_expense_amount_positive CHECK (amount > 0),
	CONSTRAINT chk_fixed_expense_status CHECK (status IN ('ACTIVE', 'PAUSED', 'CANCELLED'))
);

CREATE INDEX idx_fixed_expense_user_id_status ON fixed_expense (user_id, status);

CREATE TYPE transaction_type AS ENUM ('EXPENSE', 'INCOME');
CREATE TYPE transaction_status AS ENUM ('ACTIVE', 'CANCELLED');
CREATE TYPE transaction_source AS ENUM ('MANUAL', 'IMPORT');
CREATE TYPE transaction_visibility AS ENUM ('PRIVATE', 'PUBLIC');

CREATE TABLE transaction_category
(
	name TEXT PRIMARY KEY
);

INSERT INTO transaction_category (name)
VALUES ('GROCERY'),
	   ('HOUSING'),
	   ('UTILITIES'),
	   ('SUBSCRIPTIONS'),
	   ('INSURANCE'),
	   ('EDUCATION'),
	   ('TRANSPORTATION'),
	   ('HEALTH'),
	   ('PERSONAL_CARE'),
	   ('TAXES'),
	   ('OTHER'),
	   ('FOOD_AND_DINING'),
	   ('ENTERTAINMENT'),
	   ('SHOPPING'),
	   ('PETS'),
	   ('TRAVEL'),
	   ('INVESTMENTS');

CREATE TABLE transaction
(
	id              BIGSERIAL PRIMARY KEY,
	type            transaction_type       NOT NULL,
	description     TEXT           NOT NULL,
	amount          NUMERIC(19, 2) NOT NULL,
	category        TEXT           NOT NULL,
	status          transaction_status     NOT NULL DEFAULT 'ACTIVE',
	source          transaction_source     NOT NULL DEFAULT 'MANUAL',
	visibility      transaction_visibility NOT NULL DEFAULT 'PUBLIC',
	occurred_at     TIMESTAMPTZ    NOT NULL,
	user_account_id BIGINT         NOT NULL REFERENCES user_account (id),
	guild_id        BIGINT         NOT NULL REFERENCES guild (id),
	created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
	updated_at      TIMESTAMPTZ,
	created_by      TEXT           NOT NULL,
	updated_by      TEXT,
	CONSTRAINT chk_transactions_amount_positive CHECK (amount > 0),
	CONSTRAINT fk_transactions_category FOREIGN KEY (category) REFERENCES transaction_category (name)
);

CREATE INDEX idx_transactions_guild_occurred ON transaction (guild_id, occurred_at DESC, id DESC);
CREATE INDEX idx_transactions_guild_category ON transaction (guild_id, category);
CREATE INDEX idx_transactions_guild_type ON transaction (guild_id, type);
CREATE INDEX idx_transactions_guild_status ON transaction (guild_id, status);
CREATE INDEX idx_transactions_user_occurred ON transaction (user_account_id, occurred_at DESC);
CREATE INDEX idx_transactions_guild_visibility_user ON transaction (guild_id, visibility, user_account_id);

CREATE TYPE import_batch_status AS ENUM ('PENDING_REVIEW', 'COMPLETED', 'CANCELLED');
CREATE TYPE import_item_status AS ENUM ('READY', 'DUPLICATE', 'DISCARDED');

CREATE TABLE import_batch
(
	id              BIGSERIAL PRIMARY KEY,
	guild_id        BIGINT              NOT NULL REFERENCES guild (id),
	user_account_id BIGINT              NOT NULL REFERENCES user_account (id),
	file_name       TEXT                NOT NULL,
	source_bank     TEXT                NOT NULL,
	status          import_batch_status NOT NULL DEFAULT 'PENDING_REVIEW',
	created_at      TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
	created_by      TEXT                NOT NULL,
	updated_at      TIMESTAMPTZ,
	updated_by      TEXT
);

CREATE INDEX idx_import_batch_guild_user ON import_batch (guild_id, user_account_id);

CREATE TABLE import_item
(
	id                  BIGSERIAL PRIMARY KEY,
	import_batch_id     BIGINT           NOT NULL REFERENCES import_batch (id) ON DELETE CASCADE,
	user_account_id     BIGINT           NOT NULL REFERENCES user_account (id),
	occurred_at         TIMESTAMPTZ      NOT NULL,
	description         TEXT             NOT NULL,
	amount              NUMERIC(19, 2)   NOT NULL,
	type                transaction_type NOT NULL,
	category            TEXT             NOT NULL REFERENCES transaction_category (name),
	bank_category       TEXT             NOT NULL,
	card_last4          TEXT             NOT NULL,
	installment_current INTEGER,
	installment_total   INTEGER,
	status              import_item_status NOT NULL DEFAULT 'READY',
	transaction_id      BIGINT REFERENCES transaction (id),
	created_at          TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
	created_by          TEXT             NOT NULL,
	updated_at          TIMESTAMPTZ,
	updated_by          TEXT,
	CONSTRAINT chk_import_item_amount_positive CHECK (amount > 0),
	CONSTRAINT chk_import_item_card_last4 CHECK (char_length(card_last4) = 4),
	CONSTRAINT chk_import_item_installment CHECK (
		(installment_current IS NULL AND installment_total IS NULL)
			OR (installment_current > 0 AND installment_total > 0 AND installment_current <= installment_total)
		)
);

CREATE INDEX idx_import_item_import_batch ON import_item (import_batch_id);
CREATE INDEX idx_import_item_dedup_support ON import_item (user_account_id, occurred_at, amount, card_last4);

