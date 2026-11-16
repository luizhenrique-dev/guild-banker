BEGIN;
ALTER TABLE fixed_expense DROP COLUMN payment_method;
COMMIT;
