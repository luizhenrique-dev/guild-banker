package user

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type lookupDriver struct {
	query func(context.Context, string, []driver.NamedValue) (driver.Rows, error)
}

func (d *lookupDriver) Open(string) (driver.Conn, error)             { return &lookupConn{query: d.query}, nil }
func (d *lookupDriver) Connect(context.Context) (driver.Conn, error) { return d.Open("") }
func (d *lookupDriver) Driver() driver.Driver                        { return d }

type lookupConn struct {
	query func(context.Context, string, []driver.NamedValue) (driver.Rows, error)
}

func (c *lookupConn) Prepare(string) (driver.Stmt, error) { return nil, driver.ErrSkip }
func (c *lookupConn) Close() error                        { return nil }
func (c *lookupConn) Begin() (driver.Tx, error)           { return nil, driver.ErrSkip }
func (c *lookupConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return c.query(ctx, query, args)
}

type emptyRows struct{}

func (emptyRows) Columns() []string         { return []string{"id"} }
func (emptyRows) Close() error              { return nil }
func (emptyRows) Next([]driver.Value) error { return io.EOF }

func TestStorageGetByExternalID(t *testing.T) {
	technicalErr := errors.New("technical database failure")
	for _, tt := range []struct {
		name              string
		queryErr, wantErr error
	}{
		{name: "not found", wantErr: ErrNotFound},
		{name: "technical", queryErr: technicalErr, wantErr: technicalErr},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			db := sql.OpenDB(&lookupDriver{query: func(got context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
				assert.Equal(t, ctx, got)
				assert.Contains(t, query, "WHERE external_id = $1")
				require.Len(t, args, 1)
				assert.Equal(t, "external-sub", args[0].Value)
				return emptyRows{}, tt.queryErr
			}})
			t.Cleanup(func() { require.NoError(t, db.Close()) })
			u, err := NewStorage(sqlx.NewDb(db, "test")).GetByExternalID(ctx, "external-sub")
			assert.Nil(t, u)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.queryErr == nil {
				assert.ErrorIs(t, err, sql.ErrNoRows)
				return
			}
			assert.NotErrorIs(t, err, ErrNotFound)
		})
	}
}
