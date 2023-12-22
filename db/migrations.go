// SPDX-FileCopyrightText: Chris Waldon <christopher.waldon.dev@gmail.com>
//
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
)

type migration struct {
	upQuery   string
	downQuery string
	postHook  func(*sql.Tx) error
}

var (
	//go:embed sql/1_add_project_ids.up.sql
	migration1Up string
	//go:embed sql/1_add_project_ids.down.sql
	migration1Down string
)

var migrations = [...]migration{
	0: {
		upQuery: `CREATE TABLE schema_migrations (version uint64, dirty bool);
		INSERT INTO schema_migrations (version, dirty) VALUES (0, 0);`,
		downQuery: `DROP TABLE schema_migrations;`,
	},
	1: {
		upQuery:   migration1Up,
		downQuery: migration1Down,
		postHook:  generateAndInsertProjectIDs,
	},
}

// Migrate runs all pending migrations
func Migrate(db *sql.DB) error {
	version := getSchemaVersion(db)
	for nextMigration := version + 1; nextMigration < len(migrations); nextMigration++ {
		if err := runMigration(db, nextMigration); err != nil {
			return fmt.Errorf("migrations failed: %w", err)
		}
		if version := getSchemaVersion(db); version != nextMigration {
			return fmt.Errorf("migration did not update version (expected %d, got %d)", nextMigration, version)
		}
	}
	return nil
}

// runMigration runs a single migration inside a transaction, updates the schema
// version and commits the transaction if successful, and rolls back the
// transaction if unsuccessful.
func runMigration(db *sql.DB, migrationIdx int) (err error) {
	current := migrations[migrationIdx]
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed opening transaction for migration %d: %w", migrationIdx, err)
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
		}
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("failed rolling back: %w due to: %w", rbErr, err)
			}
		}
	}()
	if len(current.upQuery) > 0 {
		if _, err := tx.Exec(current.upQuery); err != nil {
			return fmt.Errorf("failed running migration %d: %w", migrationIdx, err)
		}
	}
	if current.postHook != nil {
		if err := current.postHook(tx); err != nil {
			return fmt.Errorf("failed running posthook for migration %d: %w", migrationIdx, err)
		}
	}
	return updateSchemaVersion(tx, migrationIdx)
}

// undoMigration rolls the single most recent migration back inside a
// transaction, updates the schema version and commits the transaction if
// successful, and rolls back the transaction if unsuccessful.
//
//lint:ignore U1000 Will be used when #34 is implemented (https://todo.sr.ht/~amolith/willow/34)
func undoMigration(db *sql.DB, migrationIdx int) (err error) {
	current := migrations[migrationIdx]
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed opening undo transaction for migration %d: %w", migrationIdx, err)
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
		}
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("failed rolling back: %w due to: %w", rbErr, err)
			}
		}
	}()
	if len(current.downQuery) > 0 {
		if _, err := tx.Exec(current.downQuery); err != nil {
			return fmt.Errorf("failed undoing migration %d: %w", migrationIdx, err)
		}
	}
	return updateSchemaVersion(tx, migrationIdx-1)
}

// getSchemaVersion returns the schema version from the database
func getSchemaVersion(db *sql.DB) int {
	row := db.QueryRowContext(context.Background(), `SELECT version FROM schema_migrations LIMIT 1;`)
	var version int
	if err := row.Scan(&version); err != nil {
		version = -1
	}
	return version
}

// updateSchemaVersion sets the version to the provided int
func updateSchemaVersion(tx *sql.Tx, version int) error {
	if version < 0 {
		// Do not try to use the schema_migrations table in a schema version where it doesn't exist
		return nil
	}
	_, err := tx.Exec(`UPDATE schema_migrations SET version = @version;`, sql.Named("version", version))
	return err
}
