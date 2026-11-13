BEGIN;
-- Refuse rollback while cancelled invitations exist, preserving their history.
ALTER TABLE guild_invitation DROP CONSTRAINT guild_invitation_status_check;
ALTER TABLE guild_invitation ADD CONSTRAINT guild_invitation_status_check
 CHECK (status IN ('PENDING', 'ACCEPTED', 'DECLINED'));
ALTER TABLE user_account DROP COLUMN last_access_at;
COMMIT;
