package forms

import (
	"errors"

	"github.com/har4s/ohmygo/core"
	"github.com/har4s/ohmygo/dao"
	"github.com/har4s/ohmygo/model"
	"github.com/har4s/ohmygo/validation"
	"github.com/har4s/ohmygo/validation/is"
)

// UserLogin is an user email/pass login form.
type UserLogin struct {
	app core.App
	dao *dao.Dao

	Identity string `form:"identity" json:"identity"`
	Password string `form:"password" json:"password"`
}

// NewUserLogin creates a new [UserLogin] form initialized with
// the provided [core.App] instance.
//
// If you want to submit the form as part of a transaction,
// you can change the default Dao via [SetDao()].
func NewUserLogin(app core.App) *UserLogin {
	return &UserLogin{
		app: app,
		dao: app.Dao(),
	}
}

// Validate makes the form validatable by implementing [validation.Validatable] interface.
func (form *UserLogin) Validate() error {
	return validation.ValidateStruct(form,
		validation.Field(&form.Identity, validation.Required, validation.Length(1, 255), is.EmailFormat),
		validation.Field(&form.Password, validation.Required, validation.Length(1, 255)),
	)
}

// Submit validates and submits the user form.
// On success returns the authorized user model.
func (form *UserLogin) Submit() (*model.User, error) {
	if err := form.Validate(); err != nil {
		return nil, err
	}

	user, err := form.dao.FindUserByEmail(form.Identity)
	if err != nil {
		return nil, err
	}

	if user.ValidatePassword(form.Password) {
		return user, nil
	}

	return nil, errors.New("invalid login credentials")
}
