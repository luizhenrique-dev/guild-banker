package user

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// NormalizeStoredEmails lowercases legacy account emails without merging accounts.
func NormalizeStoredEmails(ctx context.Context, db *sqlx.DB) error {
	const query = `UPDATE user_account
        SET email = LOWER(BTRIM(email))
        WHERE email IS DISTINCT FROM LOWER(BTRIM(email))`
	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("normalize stored account emails: %w", profileError(err))
	}
	return nil
}
