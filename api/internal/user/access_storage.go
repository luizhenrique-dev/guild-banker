package user

import (
	"context"
	"fmt"
	"time"
)

func (s *postgresStorage) RecordAccess(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	const query = `UPDATE user_account
	 SET last_access_at = GREATEST(last_access_at, clock_timestamp())
	 WHERE id = $1 AND enabled`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("record account access: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("count accessed accounts: %w", err)
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}
