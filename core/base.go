package core

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/har4s/ohmygo/dao"
	"github.com/har4s/ohmygo/dbx"
	"github.com/har4s/ohmygo/model"
	"github.com/har4s/ohmygo/model/settings"
	"github.com/har4s/ohmygo/tools/filesystem"
	"github.com/har4s/ohmygo/tools/hook"
	"github.com/har4s/ohmygo/tools/mailer"
	"github.com/har4s/ohmygo/tools/store"
)

const (
	DefaultDataMaxOpenConns int = 100
	DefaultDataMaxIdleConns int = 20
	DefaultLogsMaxOpenConns int = 10
	DefaultLogsMaxIdleConns int = 2
)

type BaseApp struct {
	// configurable parameters
	isDebug          bool
	databaseUrl      string
	dataMaxOpenConns int
	dataMaxIdleConns int
	logsMaxOpenConns int
	logsMaxIdleConns int

	// internals
	cache    *store.Store[any]
	settings *settings.Settings
	dao      *dao.Dao

	// app event hooks
	onBeforeBootstrap *hook.Hook[*BootstrapEvent]
	onAfterBootstrap  *hook.Hook[*BootstrapEvent]
	onBeforeServe     *hook.Hook[*ServeEvent]
	onBeforeApiError  *hook.Hook[*ApiErrorEvent]
	onAfterApiError   *hook.Hook[*ApiErrorEvent]

	// dao event hooks
	onModelBeforeCreate *hook.Hook[*ModelEvent]
	onModelAfterCreate  *hook.Hook[*ModelEvent]
	onModelBeforeUpdate *hook.Hook[*ModelEvent]
	onModelAfterUpdate  *hook.Hook[*ModelEvent]
	onModelBeforeDelete *hook.Hook[*ModelEvent]
	onModelAfterDelete  *hook.Hook[*ModelEvent]

	// settings api event hooks
	onSettingsListRequest         *hook.Hook[*SettingsListEvent]
	onSettingsBeforeUpdateRequest *hook.Hook[*SettingsUpdateEvent]
	onSettingsAfterUpdateRequest  *hook.Hook[*SettingsUpdateEvent]

	// user api event hooks
	onUserAuthRequest *hook.Hook[*UserAuthEvent]
}

// BaseAppConfig defines a BaseApp configuration option
type BaseAppConfig struct {
	IsDebug          bool
	DatabaseURL      string
	DataMaxOpenConns int // default to 100
	DataMaxIdleConns int // default 20
	LogsMaxOpenConns int // default to 10
	LogsMaxIdleConns int // default to 2
}

// NewBaseApp creates and returns a new BaseApp instance
// configured with the provided arguments.
//
// To initialize the app, you need to call `app.Bootstrap()`.
func NewBaseApp(config *BaseAppConfig) *BaseApp {
	app := &BaseApp{
		isDebug:          config.IsDebug,
		databaseUrl:      config.DatabaseURL,
		dataMaxOpenConns: config.DataMaxOpenConns,
		dataMaxIdleConns: config.DataMaxIdleConns,
		logsMaxOpenConns: config.LogsMaxOpenConns,
		logsMaxIdleConns: config.LogsMaxIdleConns,
		cache:            store.New[any](nil),
		settings:         settings.New(),

		// app event hooks
		onBeforeBootstrap: &hook.Hook[*BootstrapEvent]{},
		onAfterBootstrap:  &hook.Hook[*BootstrapEvent]{},
		onBeforeServe:     &hook.Hook[*ServeEvent]{},
		onBeforeApiError:  &hook.Hook[*ApiErrorEvent]{},
		onAfterApiError:   &hook.Hook[*ApiErrorEvent]{},

		// dao event hooks
		onModelBeforeCreate: &hook.Hook[*ModelEvent]{},
		onModelAfterCreate:  &hook.Hook[*ModelEvent]{},
		onModelBeforeUpdate: &hook.Hook[*ModelEvent]{},
		onModelAfterUpdate:  &hook.Hook[*ModelEvent]{},
		onModelBeforeDelete: &hook.Hook[*ModelEvent]{},
		onModelAfterDelete:  &hook.Hook[*ModelEvent]{},

		// settings API event hooks
		onSettingsListRequest:         &hook.Hook[*SettingsListEvent]{},
		onSettingsBeforeUpdateRequest: &hook.Hook[*SettingsUpdateEvent]{},
		onSettingsAfterUpdateRequest:  &hook.Hook[*SettingsUpdateEvent]{},

		// user api event hooks
		onUserAuthRequest: &hook.Hook[*UserAuthEvent]{},
	}

	app.registerDefaultHooks()

	return app
}

// IsBootstrapped checks if the application was initialized
// (aka. whether Bootstrap() was called).
func (app *BaseApp) IsBootstrapped() bool {
	return app.dao != nil && app.settings != nil
}

// Bootstrap initializes the application
// (aka. create data dir, open db connections, load settings, etc.).
//
// It will call ResetBootstrapState() if the application was already bootstrapped.
func (app *BaseApp) Bootstrap() error {
	event := &BootstrapEvent{app}

	if err := app.OnBeforeBootstrap().Trigger(event); err != nil {
		return err
	}

	// clear resources of previous core state (if any)
	if err := app.ResetBootstrapState(); err != nil {
		return err
	}

	if err := app.initDB(); err != nil {
		return err
	}

	if err := app.OnAfterBootstrap().Trigger(event); err != nil && app.IsDebug() {
		log.Println(err)
	}

	return nil
}

// ResetBootstrapState takes care for releasing initialized app resources
// (eg. closing db connections).
func (app *BaseApp) ResetBootstrapState() error {
	if app.Dao() != nil {
		if err := app.Dao().ConcurrentDB().(*dbx.DB).Close(); err != nil {
			return err
		}
		if err := app.Dao().NonconcurrentDB().(*dbx.DB).Close(); err != nil {
			return err
		}
	}

	app.dao = nil
	app.settings = nil

	return nil
}

// DB returns the default app database instance.
func (app *BaseApp) DB() *dbx.DB {
	if app.Dao() == nil {
		return nil
	}

	db, ok := app.Dao().DB().(*dbx.DB)
	if !ok {
		return nil
	}

	return db
}

// Dao returns the default app Dao instance.
func (app *BaseApp) Dao() *dao.Dao {
	return app.dao
}

// IsDebug returns whether the app is in debug mode
// (showing more detailed error logs, executed sql statements, etc.).
func (app *BaseApp) IsDebug() bool {
	return app.isDebug
}

// Settings returns the loaded app settings.
func (app *BaseApp) Settings() *settings.Settings {
	return app.settings
}

// Cache returns the app internal cache store.
func (app *BaseApp) Cache() *store.Store[any] {
	return app.cache
}

// NewMailClient creates and returns a new SMTP or Sendmail client
// based on the current app settings.
func (app *BaseApp) NewMailClient() mailer.Mailer {
	if app.Settings().Smtp.Enabled {
		return &mailer.SmtpClient{
			Host:       app.Settings().Smtp.Host,
			Port:       app.Settings().Smtp.Port,
			Username:   app.Settings().Smtp.Username,
			Password:   app.Settings().Smtp.Password,
			Tls:        app.Settings().Smtp.Tls,
			AuthMethod: app.Settings().Smtp.AuthMethod,
		}
	}

	return &mailer.Sendmail{}
}

// NewFilesystem creates a new local or S3 filesystem instance
// based on the current app settings.
//
// NB! Make sure to call `Close()` on the returned result
// after you are done working with it.
func (app *BaseApp) NewFilesystem() (*filesystem.System, error) {
	if app.settings.S3.Enabled {
		return filesystem.NewS3(
			app.settings.S3.Bucket,
			app.settings.S3.Region,
			app.settings.S3.Endpoint,
			app.settings.S3.AccessKey,
			app.settings.S3.Secret,
			app.settings.S3.ForcePathStyle,
		)
	}

	// fallback to local filesystem
	return filesystem.NewLocal("")
}

// RefreshSettings reinitializes and reloads the stored application settings.
func (app *BaseApp) RefreshSettings() error {
	if app.settings == nil {
		app.settings = settings.New()
	}

	storedSettings, err := app.Dao().FindSettings()
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// no settings were previously stored
	if storedSettings == nil {
		return app.Dao().SaveSettings(app.settings)
	}

	// load the settings from the stored param into the app ones
	if err := app.settings.Merge(storedSettings); err != nil {
		return err
	}

	return nil
}

// -------------------------------------------------------------------
// App event hooks
// -------------------------------------------------------------------

func (app *BaseApp) OnBeforeBootstrap() *hook.Hook[*BootstrapEvent] {
	return app.onBeforeBootstrap
}

func (app *BaseApp) OnAfterBootstrap() *hook.Hook[*BootstrapEvent] {
	return app.onAfterBootstrap
}

func (app *BaseApp) OnBeforeServe() *hook.Hook[*ServeEvent] {
	return app.onBeforeServe
}

func (app *BaseApp) OnBeforeApiError() *hook.Hook[*ApiErrorEvent] {
	return app.onBeforeApiError
}

func (app *BaseApp) OnAfterApiError() *hook.Hook[*ApiErrorEvent] {
	return app.onAfterApiError
}

// -------------------------------------------------------------------
// Dao event hooks
// -------------------------------------------------------------------

func (app *BaseApp) OnModelBeforeCreate() *hook.Hook[*ModelEvent] {
	return app.onModelBeforeCreate
}

func (app *BaseApp) OnModelAfterCreate() *hook.Hook[*ModelEvent] {
	return app.onModelAfterCreate
}

func (app *BaseApp) OnModelBeforeUpdate() *hook.Hook[*ModelEvent] {
	return app.onModelBeforeUpdate
}

func (app *BaseApp) OnModelAfterUpdate() *hook.Hook[*ModelEvent] {
	return app.onModelAfterUpdate
}

func (app *BaseApp) OnModelBeforeDelete() *hook.Hook[*ModelEvent] {
	return app.onModelBeforeDelete
}

func (app *BaseApp) OnModelAfterDelete() *hook.Hook[*ModelEvent] {
	return app.onModelAfterDelete
}

// -------------------------------------------------------------------
// Settings API event hooks
// -------------------------------------------------------------------

func (app *BaseApp) OnSettingsListRequest() *hook.Hook[*SettingsListEvent] {
	return app.onSettingsListRequest
}

func (app *BaseApp) OnSettingsBeforeUpdateRequest() *hook.Hook[*SettingsUpdateEvent] {
	return app.onSettingsBeforeUpdateRequest
}

func (app *BaseApp) OnSettingsAfterUpdateRequest() *hook.Hook[*SettingsUpdateEvent] {
	return app.onSettingsAfterUpdateRequest
}

// -------------------------------------------------------------------
// User API event hooks
// -------------------------------------------------------------------

func (app *BaseApp) OnUserAuthRequest() *hook.Hook[*UserAuthEvent] {
	return app.onUserAuthRequest
}

// -------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------

func (app *BaseApp) initDB() error {
	maxOpenConns := DefaultDataMaxOpenConns
	maxIdleConns := DefaultDataMaxIdleConns
	if app.dataMaxOpenConns > 0 {
		maxOpenConns = app.dataMaxOpenConns
	}
	if app.dataMaxIdleConns > 0 {
		maxIdleConns = app.dataMaxIdleConns
	}

	concurrentDB, err := connectDB(app.databaseUrl)
	if err != nil {
		return err
	}
	concurrentDB.DB().SetMaxOpenConns(maxOpenConns)
	concurrentDB.DB().SetMaxIdleConns(maxIdleConns)
	concurrentDB.DB().SetConnMaxIdleTime(5 * time.Minute)

	nonconcurrentDB, err := connectDB(app.databaseUrl)
	if err != nil {
		return err
	}
	nonconcurrentDB.DB().SetMaxOpenConns(1)
	nonconcurrentDB.DB().SetMaxIdleConns(1)
	nonconcurrentDB.DB().SetConnMaxIdleTime(5 * time.Minute)

	if app.IsDebug() {
		nonconcurrentDB.QueryLogFunc = func(ctx context.Context, t time.Duration, sql string, rows *sql.Rows, err error) {
			color.HiBlack("[%.2fms] %v\n", float64(t.Milliseconds()), sql)
		}
		concurrentDB.QueryLogFunc = nonconcurrentDB.QueryLogFunc

		nonconcurrentDB.ExecLogFunc = func(ctx context.Context, t time.Duration, sql string, result sql.Result, err error) {
			color.HiBlack("[%.2fms] %v\n", float64(t.Milliseconds()), sql)
		}
		concurrentDB.ExecLogFunc = nonconcurrentDB.ExecLogFunc
	}

	app.dao = app.createDaoWithHooks(concurrentDB, nonconcurrentDB)

	return nil
}

func (app *BaseApp) createDaoWithHooks(concurrentDB, nonconcurrentDB dbx.Builder) *dao.Dao {
	d := dao.NewMultiDB(concurrentDB, nonconcurrentDB)

	d.BeforeCreateFunc = func(eventDao *dao.Dao, m model.Model) error {
		return app.OnModelBeforeCreate().Trigger(&ModelEvent{eventDao, m})
	}

	d.AfterCreateFunc = func(eventDao *dao.Dao, m model.Model) {
		err := app.OnModelAfterCreate().Trigger(&ModelEvent{eventDao, m})
		if err != nil && app.isDebug {
			log.Println(err)
		}
	}

	d.BeforeUpdateFunc = func(eventDao *dao.Dao, m model.Model) error {
		return app.OnModelBeforeUpdate().Trigger(&ModelEvent{eventDao, m})
	}

	d.AfterUpdateFunc = func(eventDao *dao.Dao, m model.Model) {
		err := app.OnModelAfterUpdate().Trigger(&ModelEvent{eventDao, m})
		if err != nil && app.isDebug {
			log.Println(err)
		}
	}

	d.BeforeDeleteFunc = func(eventDao *dao.Dao, m model.Model) error {
		return app.OnModelBeforeDelete().Trigger(&ModelEvent{eventDao, m})
	}

	d.AfterDeleteFunc = func(eventDao *dao.Dao, m model.Model) {
		err := app.OnModelAfterDelete().Trigger(&ModelEvent{eventDao, m})
		if err != nil && app.isDebug {
			log.Println(err)
		}
	}

	return d
}

func (app *BaseApp) registerDefaultHooks() {
}
