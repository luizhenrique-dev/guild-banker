BEGIN;
ALTER TABLE fixed_expense ADD COLUMN responsible_user_id BIGINT;
ALTER TABLE fixed_expense ADD CONSTRAINT fixed_expense_responsible_member_fk
 FOREIGN KEY (guild_id, responsible_user_id) REFERENCES guild_member (guild_id, user_id)
 ON DELETE SET NULL (responsible_user_id);
CREATE INDEX idx_fixed_expense_responsible_member ON fixed_expense (guild_id, responsible_user_id)
 WHERE responsible_user_id IS NOT NULL;
COMMIT;
