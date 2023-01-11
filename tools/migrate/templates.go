package migrate

import (
	"fmt"
	"path/filepath"
)

// -------------------------------------------------------------------
// Go templates
// -------------------------------------------------------------------

func (p *MigrateCmd) goBlankTemplate() (string, error) {
	const template = `package %s

import "github.com/har4s/ohmygo/dbx"

func init() {
	Register(func(db dbx.Builder) error {
		// add up queries...
		return nil
	}, func(db dbx.Builder) error {
		// add down queries...
		return nil
	})
}
`

	return fmt.Sprintf(template, filepath.Base(p.options.Dir)), nil
}
