package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// InTransaction publishes changes only after every operation succeeds.
func InTransaction(db *sqlx.DB, update func(*sqlx.Tx) error) error {
	if db == nil {
		return fmt.Errorf("Database not initialized")
	}
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("Begin transaction: %w", err)
	}
	defer tx.Rollback()
	if err := update(tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("Commit transaction: %w", err)
	}
	return nil
}
