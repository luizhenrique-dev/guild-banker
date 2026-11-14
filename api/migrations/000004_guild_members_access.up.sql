BEGIN;
ALTER TABLE user_account ADD COLUMN last_access_at TIMESTAMPTZ;
ALTER TABLE guild_invitation DROP CONSTRAINT guild_invitation_status_check;
ALTER TABLE guild_invitation ADD CONSTRAINT guild_invitation_status_check
 CHECK (status IN ('PENDING', 'ACCEPTED', 'DECLINED', 'CANCELLED'));
COMMIT;
