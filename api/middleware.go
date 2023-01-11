package api

import (
	"net/http"
	"strings"

	"github.com/har4s/ohmygo/core"
	"github.com/har4s/ohmygo/model"
	"github.com/har4s/ohmygo/tokens"
	"github.com/har4s/ohmygo/tools/security"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"
)

// Common request context keys used by the middlewares and api handlers.
const (
	ContextUserKey string = "user"
)

// RequireAuth middleware requires a request to have
// a valid user Authorization header.
func RequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, _ := c.Get(ContextUserKey).(*model.User)
			if user == nil {
				return NewUnauthorizedError("The request requires valid user authorization token to be set.", nil)
			}

			return next(c)
		}
	}
}

// LoadAuthContext middleware reads the Authorization request header
// and loads the token related user instance into the request's context.
//
// This middleware is expected to be already registered by default for all routes.
func LoadAuthContext(app core.App) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := c.Request().Header.Get("Authorization")
			if token == "" {
				return next(c)
			}

			// the schema is not required and it is only for
			// compatibility with the defaults of some HTTP clients
			token = strings.TrimPrefix(token, "Bearer ")

			claims, _ := security.ParseUnverifiedJWT(token)
			tokenType := cast.ToString(claims["type"])

			switch tokenType {
			case tokens.TypeUser:
				user, err := app.Dao().FindUserByToken(
					token,
					app.Settings().UserAuthToken.Secret,
				)
				if err == nil && user != nil {
					c.Set(ContextUserKey, user)
				}
			}

			return next(c)
		}
	}
}

// Returns the "real" user IP from common proxy headers (or fallbackIp if none is found).
//
// The returned IP value shouldn't be trusted if not behind a trusted reverse proxy!
func realUserIp(r *http.Request, fallbackIp string) string {
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}

	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	if ipsList := r.Header.Get("X-Forwarded-For"); ipsList != "" {
		ips := strings.Split(ipsList, ",")
		// extract the rightmost ip
		for i := len(ips) - 1; i >= 0; i-- {
			ip := strings.TrimSpace(ips[i])
			if ip != "" {
				return ip
			}
		}
	}

	return fallbackIp
}
