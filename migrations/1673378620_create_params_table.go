package migrations

import "github.com/har4s/ohmygo/dbx"

func init() {
	Register(func(db dbx.Builder) error {
		_, tablesErr := db.NewQuery(`
		CREATE TABLE {{params}} (
			[[id]] VARCHAR(255) NOT NULL,
			[[key]] VARCHAR(255) NOT NULL,
			[[value]] TEXT NULL DEFAULT NULL,
			[[created]] TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			[[updated]] TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY ([[id]]),
			UNIQUE KEY ([[key]])
		);
		`).Execute()
		if tablesErr != nil {
			return tablesErr
		}

		return nil
	}, func(db dbx.Builder) error {
		if _, err := db.DropTable("params").Execute(); err != nil {
			return err
		}

		return nil
	})
}
