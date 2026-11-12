package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// WithAdvisoryLock coordinates cooperating processes through one database session.
func WithAdvisoryLock(db *sqlx.DB, name string, update func() error) (err error) {
	if db == nil {
		return fmt.Errorf("Database not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := db.Connx(ctx)
	if err != nil {
		return fmt.Errorf("Open database lock connection: %w", err)
	}
	defer conn.Close()
	var acquired sql.NullInt64
	if err := conn.GetContext(ctx, &acquired, "SELECT GET_LOCK(?, 0)", name); err != nil {
		// The server may have acquired the lock before the response was lost.
		_ = conn.Raw(func(any) error { return driver.ErrBadConn })
		return fmt.Errorf("Acquire database lock %q: %w", name, err)
	}
	if !acquired.Valid {
		return fmt.Errorf("Database failed to acquire lock %q", name)
	}
	if acquired.Int64 != 1 {
		return fmt.Errorf("Database operation %q is already running", name)
	}
	defer func() {
		releaseCtx, releaseCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer releaseCancel()
		var released sql.NullInt64
		releaseErr := conn.GetContext(releaseCtx, &released, "SELECT RELEASE_LOCK(?)", name)
		if releaseErr == nil && released.Valid && released.Int64 == 1 {
			return
		}
		if releaseErr == nil {
			releaseErr = fmt.Errorf("Lock was not owned by this connection")
		}
		// Discard the physical connection so a held lock cannot return to the pool.
		_ = conn.Raw(func(any) error { return driver.ErrBadConn })
		err = errors.Join(err, fmt.Errorf("Release database lock %q: %w", name, releaseErr))
	}()
	return update()
}
