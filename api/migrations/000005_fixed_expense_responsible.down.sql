BEGIN;
ALTER TABLE fixed_expense DROP COLUMN responsible_user_id;
COMMIT;
