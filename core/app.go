package core

import (
	"github.com/har4s/ohmygo/dao"
	"github.com/har4s/ohmygo/dbx"
	"github.com/har4s/ohmygo/model/settings"
	"github.com/har4s/ohmygo/tools/filesystem"
	"github.com/har4s/ohmygo/tools/hook"
	"github.com/har4s/ohmygo/tools/mailer"
	"github.com/har4s/ohmygo/tools/store"
)

// App defines the main app interface.
type App interface {
	// DB returns the default app database instance.
	DB() *dbx.DB

	// Dao returns the default app Dao instance.
	Dao() *dao.Dao

	// IsDebug returns whether the app is in debug mode
	// (showing more detailed error logs, executed sql statements, etc.).
	IsDebug() bool

	// Settings returns the loaded app settings.
	Settings() *settings.Settings

	// Cache returns the app internal cache store.
	Cache() *store.Store[any]

	// NewMailClient creates and returns a configured app mail client.
	NewMailClient() mailer.Mailer

	// NewFilesystem creates and returns a configured filesystem.System instance.
	//
	// NB! Make sure to call `Close()` on the returned result
	// after you are done working with it.
	NewFilesystem() (*filesystem.System, error)

	// RefreshSettings reinitializes and reloads the stored application settings.
	RefreshSettings() error

	// IsBootstrapped checks if the application was initialized
	// (aka. whether Bootstrap() was called).
	IsBootstrapped() bool

	// Bootstrap takes care for initializing the application
	// (open db connections, load settings, etc.).
	//
	// It will call ResetBootstrapState() if the application was already bootstrapped.
	Bootstrap() error

	// ResetBootstrapState takes care for releasing initialized app resources
	// (eg. closing db connections).
	ResetBootstrapState() error

	// ---------------------------------------------------------------
	// App event hooks
	// ---------------------------------------------------------------

	// OnBeforeBootstrap hook is triggered before initializing the base
	// application resources (eg. before db open and initial settings load).
	OnBeforeBootstrap() *hook.Hook[*BootstrapEvent]

	// OnAfterBootstrap hook is triggered after initializing the base
	// application resources (eg. after db open and initial settings load).
	OnAfterBootstrap() *hook.Hook[*BootstrapEvent]

	// OnBeforeServe hook is triggered before serving the internal router (echo),
	// allowing you to adjust its options and attach new routes.
	OnBeforeServe() *hook.Hook[*ServeEvent]

	// OnBeforeApiError hook is triggered right before sending an error API
	// response to the client, allowing you to further modify the error data
	// or to return a completely different API response (using [hook.StopPropagation]).
	OnBeforeApiError() *hook.Hook[*ApiErrorEvent]

	// OnAfterApiError hook is triggered right after sending an error API
	// response to the client.
	// It could be used to log the final API error in external services.
	OnAfterApiError() *hook.Hook[*ApiErrorEvent]

	// ---------------------------------------------------------------
	// Dao event hooks
	// ---------------------------------------------------------------

	// OnModelBeforeCreate hook is triggered before inserting a new
	// entry in the DB, allowing you to modify or validate the stored data.
	OnModelBeforeCreate() *hook.Hook[*ModelEvent]

	// OnModelAfterCreate hook is triggered after successfully
	// inserting a new entry in the DB.
	OnModelAfterCreate() *hook.Hook[*ModelEvent]

	// OnModelBeforeUpdate hook is triggered before updating existing
	// entry in the DB, allowing you to modify or validate the stored data.
	OnModelBeforeUpdate() *hook.Hook[*ModelEvent]

	// OnModelAfterUpdate hook is triggered after successfully updating
	// existing entry in the DB.
	OnModelAfterUpdate() *hook.Hook[*ModelEvent]

	// OnModelBeforeDelete hook is triggered before deleting an
	// existing entry from the DB.
	OnModelBeforeDelete() *hook.Hook[*ModelEvent]

	// OnModelAfterDelete is triggered after successfully deleting an
	// existing entry from the DB.
	OnModelAfterDelete() *hook.Hook[*ModelEvent]

	// ---------------------------------------------------------------
	// User API event hooks
	// ---------------------------------------------------------------

	// OnUserAuthRequest hook is triggered on each successful API User
	// authentication request (sign-in, token refresh, etc.).
	//
	// Could be used to additionally validate or modify the
	// authenticated user data and token.
	OnUserAuthRequest() *hook.Hook[*UserAuthEvent]
}
