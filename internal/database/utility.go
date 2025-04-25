package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/gommon/log"
)

// HandleTransaction ensures that a transaction is committed or rolled back properly.
func HandleTransaction(tx *sqlx.Tx, err *error) {
	if p := recover(); p != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			log.Errorf("Failed to rollback transaction: %v\n", rollbackErr)
		}
		panic(p) // Re-panic after rollback
	} else if *err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			log.Errorf("Failed to rollback transaction: %v\n", rollbackErr)
		}
	} else {
		commitErr := tx.Commit()
		if commitErr != nil {
			log.Errorf("Failed to commit transaction: %v\n", commitErr)
			*err = fmt.Errorf("commit failed: %w", commitErr)
		}
	}
}
