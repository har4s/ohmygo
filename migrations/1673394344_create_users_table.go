package migrations

import "github.com/har4s/ohmygo/dbx"

func init() {
	Register(func(db dbx.Builder) error {
		_, tablesErr := db.NewQuery(`
		CREATE TABLE {{users}} (
			[[id]] VARCHAR(255) NOT NULL,
			[[email]] VARCHAR(255) NOT NULL,
			[[tokenKey]] VARCHAR(255) NOT NULL,
			[[passwordHash]] VARCHAR(255) NOT NULL,
			[[isAdmin]] BOOLEAN NOT NULL DEFAULT FALSE,
			[[isSuperadmin]] BOOLEAN NOT NULL DEFAULT FALSE,
			[[lastResetSentAt]] TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
			[[created]] TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			[[updated]] TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY ([[id]]),
			UNIQUE KEY ([[email]]),
			UNIQUE KEY ([[tokenKey]])
		);
		`).Execute()
		if tablesErr != nil {
			return tablesErr
		}

		return nil
	}, func(db dbx.Builder) error {
		if _, err := db.DropTable("users").Execute(); err != nil {
			return err
		}

		return nil
	})
}
