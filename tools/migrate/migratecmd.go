package migrate

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/har4s/ohmygo/core"
	"github.com/har4s/ohmygo/tools/inflector"
)

// Options defines optional struct to customize the default plugin behavior.
type Options struct {
	// Dir specifies the migrations directory.
	Dir string

	// Automigrate specifies whether to enable automigrations.
	Automigrate bool
}

type MigrateCmd struct {
	app     core.App
	options *Options
}

func NewMigrateCmd(app core.App) *MigrateCmd {
	m := &MigrateCmd{
		app:     app,
		options: &Options{},
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil
	}

	m.options.Dir = filepath.Join(wd, "./migrations")

	return m
}

func (m *MigrateCmd) MigrateCreateHandler(args []string, interactive bool) error {
	if len(args) < 1 {
		return fmt.Errorf("missing migration file name")
	}

	name := args[0]
	dir := m.options.Dir

	resultFilePath := path.Join(
		dir,
		fmt.Sprintf("%d_%s.%s", time.Now().Unix(), inflector.Snakecase(name), "go"),
	)

	if interactive {
		confirm := false
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Do you really want to create migration %q?", resultFilePath),
		}
		survey.AskOne(prompt, &confirm)
		if !confirm {
			fmt.Println("The command has been cancelled")
			return nil
		}
	}

	template, templateErr := m.goBlankTemplate()

	if templateErr != nil {
		return fmt.Errorf("failed to resolve create template: %v", templateErr)
	}

	// ensure that the migrations dir exist
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	// save the migration file
	if err := os.WriteFile(resultFilePath, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to save migration file %q: %v", resultFilePath, err)
	}

	if interactive {
		fmt.Printf("Successfully created file %q\n", resultFilePath)
	}

	return nil
}
