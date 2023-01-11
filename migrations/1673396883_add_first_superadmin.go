package migrations

import (
	"github.com/har4s/ohmygo/dao"
	"github.com/har4s/ohmygo/dbx"
	"github.com/har4s/ohmygo/model"
)

func init() {
	Register(func(db dbx.Builder) error {
		user := &model.User{
			Email:        "admin@example.com",
			IsAdmin:      true,
			IsSuperadmin: true,
		}
		user.SetPassword("admin")
		user.RefreshLastResetSentAt()

		return dao.New(db).SaveUser(user)
	}, func(db dbx.Builder) error {
		d := dao.New(db)
		user, err := d.FindUserByEmail("admin@example.com")
		if err != nil {
			return err
		}

		return d.DeleteUser(user)
	})
}
