package migrations

import (
	"path/filepath"
	"runtime"

	"github.com/har4s/ohmygo/dbx"
	"github.com/har4s/ohmygo/tools/migrate"
)

var Migrations migrate.MigrationsList

// Register is a short alias for `Migrations.Register()`
// that is usually used in external/user defined migrations.
func Register(
	up func(db dbx.Builder) error,
	down func(db dbx.Builder) error,
	optFilename ...string,
) {
	var optFiles []string
	if len(optFilename) > 0 {
		optFiles = optFilename
	} else {
		_, path, _, _ := runtime.Caller(1)
		optFiles = append(optFiles, filepath.Base(path))
	}
	Migrations.Register(up, down, optFiles...)
}
