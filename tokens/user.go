package tokens

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/har4s/ohmygo/core"
	"github.com/har4s/ohmygo/model"
	"github.com/har4s/ohmygo/tools/security"
)

// NewUserAuthToken generates and returns a new user authentication token.
func NewUserAuthToken(app core.App, user *model.User) (string, error) {
	return security.NewToken(
		jwt.MapClaims{"id": user.Id, "type": TypeUser},
		(user.TokenKey + app.Settings().UserAuthToken.Secret),
		app.Settings().UserAuthToken.Duration,
	)
}

// NewUserResetPasswordToken generates and returns a new user password reset request token.
func NewUserResetPasswordToken(app core.App, user *model.User) (string, error) {
	return security.NewToken(
		jwt.MapClaims{"id": user.Id, "type": TypeUser, "email": user.Email},
		(user.TokenKey + app.Settings().UserPasswordResetToken.Secret),
		app.Settings().UserPasswordResetToken.Duration,
	)
}
