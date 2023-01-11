package api

import (
	"net/http"

	"github.com/har4s/ohmygo/core"
	"github.com/har4s/ohmygo/forms"
	"github.com/har4s/ohmygo/model"
	"github.com/har4s/ohmygo/tokens"
	"github.com/labstack/echo/v4"
)

// bindAuthApi registers the auth api endpoints and the corresponding handlers.
func bindAuthApi(app core.App, rg *echo.Echo) {
	api := authApi{app: app}
	subGroup := rg.Group("/auth")
	subGroup.POST("/login", api.authWithPassword)
	subGroup.POST("/refresh", api.authRefresh, RequireAuth())
	subGroup.POST("/logout", api.logout, RequireAuth())
	subGroup.GET("/me", api.currentUser, RequireAuth())
}

type authApi struct {
	app core.App
}

func (api *authApi) authResponse(c echo.Context, user *model.User) error {
	token, tokenErr := tokens.NewUserAuthToken(api.app, user)
	if tokenErr != nil {
		return NewBadRequestError("Failed to create auth token.", tokenErr)
	}

	event := &core.UserAuthEvent{
		HttpContext: c,
		User:        user,
		Token:       token,
	}

	return api.app.OnUserAuthRequest().Trigger(event, func(e *core.UserAuthEvent) error {
		return e.HttpContext.JSON(200, map[string]any{
			"token": e.Token,
			"user":  e.User,
		})
	})
}

func (api *authApi) authWithPassword(c echo.Context) error {
	form := forms.NewUserLogin(api.app)
	if readErr := c.Bind(form); readErr != nil {
		return NewBadRequestError("An error occurred while loading the submitted data.", readErr)
	}

	user, submitErr := form.Submit()
	if submitErr != nil {
		return NewBadRequestError("Failed to authenticate.", submitErr)
	}

	return api.authResponse(c, user)
}

func (api *authApi) authRefresh(c echo.Context) error {
	user, _ := c.Get(ContextUserKey).(*model.User)
	if user == nil {
		return NewNotFoundError("Missing auth user context.", nil)
	}

	// destroy previous tokens
	user.RefreshTokenKey()
	api.app.Dao().Save(user)

	return api.authResponse(c, user)
}

func (api *authApi) logout(c echo.Context) error {
	user, _ := c.Get(ContextUserKey).(*model.User)
	if user == nil {
		return NewNotFoundError("Missing auth user context.", nil)
	}

	// destroy previous tokens
	user.RefreshTokenKey()
	api.app.Dao().Save(user)

	return c.NoContent(http.StatusOK)
}

func (api *authApi) currentUser(c echo.Context) error {
	user, _ := c.Get(ContextUserKey).(*model.User)
	if user == nil {
		return NewNotFoundError("Missing auth user context.", nil)
	}

	return c.JSON(200, map[string]any{
		"user": user,
	})
}
