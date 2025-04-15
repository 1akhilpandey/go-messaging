package migrate

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Migrate applies database migrations in the specified direction ("up" or "down").
// The databaseURL should be provided in the format required by the sqlite3 driver, e.g. "sqlite3://chatapp.db".
// If no direction is provided, it defaults to "up".
func Migrate(databaseURL string, directions ...string) error {
	dir := "up"
	if len(directions) > 0 {
		dir = directions[0]
	}
	m, err := migrate.New(
		"file://db/migrate/sqlite/",
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize migration: %w", err)
	}
	switch dir {
	case "up":
		err = m.Up()
		if err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("migration up failed: %w", err)
		}
	case "down":
		err = m.Down()
		if err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("migration down failed: %w", err)
		}
	default:
		return fmt.Errorf("unknown migration direction: %s", dir)
	}
	return nil
}
