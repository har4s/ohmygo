package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fatih/color"
	"github.com/har4s/ohmygo/api"
	"github.com/har4s/ohmygo/core"
	"github.com/har4s/ohmygo/dbx"
	"github.com/har4s/ohmygo/migrations"
	"github.com/har4s/ohmygo/tools/migrate"
	"github.com/labstack/echo/v4/middleware"
)

// starting web server.
func serve(app core.App, port int) {
	allowedOrigins := []string{"*"} // todo: get from settings
	httpAddr := fmt.Sprintf("127.0.0.1:%d", port)

	router, err := api.InitApi(app)
	if err != nil {
		panic(err)
	}

	// configure cors
	router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		Skipper:      middleware.DefaultSkipper,
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}))

	// start http server
	// ---
	mainAddr := httpAddr

	serverConfig := &http.Server{
		ReadTimeout:       5 * time.Minute,
		ReadHeaderTimeout: 30 * time.Second,
		// WriteTimeout: 60 * time.Second, // breaks sse!
		Handler: router,
		Addr:    mainAddr,
	}

	schema := "http"
	bold := color.New(color.Bold).Add(color.FgGreen)
	bold.Printf("> Server started at: %s\n", color.CyanString("%s://%s", schema, serverConfig.Addr))

	// start HTTP server
	if serveErr := serverConfig.ListenAndServe(); serveErr != http.ErrServerClosed {
		log.Fatalln(serveErr)
	}
}

func main() {
	env := NewEnv()
	app := core.NewBaseApp(&core.BaseAppConfig{
		IsDebug:     env.IsDebug,
		DatabaseURL: env.DatabaseURL,
	})

	if err := app.Bootstrap(); err != nil {
		panic(err)
	}

	if !env.SkipMigrations {
		if err := runMigrations(app); err != nil {
			panic(err)
		}
	}

	if err := app.RefreshSettings(); err != nil {
		color.Yellow("=====================================")
		color.Yellow("WARNING: Settings load error! \n%v", err)
		color.Yellow("Fallback to the application defaults.")
		color.Yellow("=====================================")
	}

	// cmd := migrate.NewMigrateCmd(app)
	// cmd.MigrateCreateHandler([]string{""}, true)
	serve(app, env.Port)
}

/**
* Helpers
 */

type migrationsConnection struct {
	DB             *dbx.DB
	MigrationsList migrate.MigrationsList
}

func runMigrations(app core.App) error {
	connections := []migrationsConnection{
		{
			DB:             app.DB(),
			MigrationsList: migrations.Migrations,
		},
	}

	for _, c := range connections {
		runner, err := migrate.NewRunner(c.DB, c.MigrationsList)
		if err != nil {
			return err
		}

		if _, err := runner.Up(); err != nil {
			return err
		}
	}

	return nil
}
