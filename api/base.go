package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/har4s/ohmygo/core"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func InitApi(app core.App) (*echo.Echo, error) {
	e := echo.New()
	e.Debug = app.IsDebug()

	// default middlewares
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())
	e.Use(LoadAuthContext(app))

	// custom error handler
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		var apiErr *ApiError

		switch v := err.(type) {
		case *echo.HTTPError:
			if v.Internal != nil && app.IsDebug() {
				log.Println(v.Internal)
			}
			msg := fmt.Sprintf("%v", v.Message)
			apiErr = NewApiError(v.Code, msg, v)
		case *ApiError:
			if app.IsDebug() && v.RawData() != nil {
				log.Println(v.RawData())
			}
			apiErr = v
		default:
			if err != nil && app.IsDebug() {
				log.Println(err)
			}
			apiErr = NewBadRequestError("", err)
		}

		event := &core.ApiErrorEvent{
			HttpContext: c,
			Error:       apiErr,
		}

		// send error response
		hookErr := app.OnBeforeApiError().Trigger(event, func(e *core.ApiErrorEvent) error {
			// @see https://github.com/labstack/echo/issues/608
			if e.HttpContext.Request().Method == http.MethodHead {
				return e.HttpContext.NoContent(apiErr.Code)
			}

			return e.HttpContext.JSON(apiErr.Code, apiErr)
		})

		// truly rare case; eg. client already disconnected
		if hookErr != nil && app.IsDebug() {
			log.Println(hookErr)
		}

		app.OnAfterApiError().Trigger(event)
	}

	bindAuthApi(app, e)

	// trigger the custom BeforeServe hook for the created api router
	// allowing users to further adjust its options or register new routes
	serveEvent := &core.ServeEvent{
		App:    app,
		Router: e,
	}
	if err := app.OnBeforeServe().Trigger(serveEvent); err != nil {
		return nil, err
	}

	return e, nil
}
