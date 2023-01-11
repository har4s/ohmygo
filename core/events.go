package core

import (
	"github.com/har4s/ohmygo/dao"
	"github.com/har4s/ohmygo/model"
	"github.com/har4s/ohmygo/model/settings"
	"github.com/labstack/echo/v4"
)

// -------------------------------------------------------------------
// Serve events data
// -------------------------------------------------------------------

type BootstrapEvent struct {
	App App
}

type ServeEvent struct {
	App    App
	Router *echo.Echo
}

type ApiErrorEvent struct {
	HttpContext echo.Context
	Error       error
}

// -------------------------------------------------------------------
// Model DAO events data
// -------------------------------------------------------------------

type ModelEvent struct {
	Dao   *dao.Dao
	Model model.Model
}

// -------------------------------------------------------------------
// Settings API events data
// -------------------------------------------------------------------

type SettingsListEvent struct {
	HttpContext      echo.Context
	RedactedSettings *settings.Settings
}

type SettingsUpdateEvent struct {
	HttpContext echo.Context
	OldSettings *settings.Settings
	NewSettings *settings.Settings
}

// -------------------------------------------------------------------
// User API events data
// -------------------------------------------------------------------

type UserAuthEvent struct {
	HttpContext echo.Context
	User        *model.User
	Token       string
}
