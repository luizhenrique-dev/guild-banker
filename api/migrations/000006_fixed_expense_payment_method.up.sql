BEGIN;
ALTER TABLE fixed_expense ADD COLUMN payment_method TEXT;
UPDATE fixed_expense SET payment_method = 'CREDIT_CARD';
ALTER TABLE fixed_expense ALTER COLUMN payment_method SET NOT NULL;
ALTER TABLE fixed_expense ADD CONSTRAINT fixed_expense_payment_method_check
 CHECK (payment_method IN ('CREDIT_CARD', 'DEBIT_CARD', 'PIX', 'AUTOMATIC_DEBIT', 'VOUCHER'));
COMMIT;
