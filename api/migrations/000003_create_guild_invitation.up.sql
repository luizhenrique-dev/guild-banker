CREATE TABLE guild_invitation (
 id BIGSERIAL PRIMARY KEY,
 guild_id BIGINT NOT NULL REFERENCES guild(id) ON DELETE CASCADE,
 recipient_email TEXT NOT NULL,
 invited_by BIGINT NOT NULL REFERENCES user_account(id),
 status TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING','ACCEPTED','DECLINED')),
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 responded_at TIMESTAMPTZ,
 responded_by BIGINT REFERENCES user_account(id)
);
CREATE UNIQUE INDEX idx_guild_invitation_pending ON guild_invitation(guild_id, recipient_email) WHERE status='PENDING';
CREATE INDEX idx_guild_invitation_recipient ON guild_invitation(recipient_email, created_at, id) WHERE status='PENDING';
