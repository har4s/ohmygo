package dao

import (
	"errors"

	"github.com/har4s/ohmygo/dbx"
	"github.com/har4s/ohmygo/model"
	"github.com/har4s/ohmygo/tools/list"
	"github.com/har4s/ohmygo/tools/security"
)

// UserQuery returns a new User select query.
func (dao *Dao) UserQuery() *dbx.SelectQuery {
	return dao.ModelQuery(&model.User{})
}

// FindUserById finds the user with the provided id.
func (dao *Dao) FindUserById(id string) (*model.User, error) {
	model := &model.User{}

	err := dao.UserQuery().
		AndWhere(dbx.HashExp{"id": id}).
		Limit(1).
		One(model)

	if err != nil {
		return nil, err
	}

	model.MarkAsNotNew()

	return model, nil
}

// FindUserByEmail finds the user with the provided email address.
func (dao *Dao) FindUserByEmail(email string) (*model.User, error) {
	model := &model.User{}

	err := dao.UserQuery().
		AndWhere(dbx.HashExp{"email": email}).
		Limit(1).
		One(model)

	if err != nil {
		return nil, err
	}

	model.MarkAsNotNew()

	return model, nil
}

// FindUserByToken finds the user associated with the provided JWT token.
//
// Returns an error if the JWT token is invalid or expired.
func (dao *Dao) FindUserByToken(token string, baseTokenKey string) (*model.User, error) {
	// @todo consider caching the unverified claims
	unverifiedClaims, err := security.ParseUnverifiedJWT(token)
	if err != nil {
		return nil, err
	}

	// check required claims
	id, _ := unverifiedClaims["id"].(string)
	if id == "" {
		return nil, errors.New("missing or invalid token claims")
	}

	user, err := dao.FindUserById(id)
	if err != nil || user == nil {
		return nil, err
	}

	verificationKey := user.TokenKey + baseTokenKey

	// verify token signature
	if _, err := security.ParseJWT(token, verificationKey); err != nil {
		return nil, err
	}

	user.MarkAsNotNew()

	return user, nil
}

// TotalUsers returns the number of existing user records.
func (dao *Dao) TotalUsers() (int, error) {
	var total int

	err := dao.UserQuery().Select("count(*)").Row(&total)

	return total, err
}

// IsUserEmailUnique checks if the provided email address is not
// already in use by other users.
func (dao *Dao) IsUserEmailUnique(email string, excludeIds ...string) bool {
	if email == "" {
		return false
	}

	query := dao.UserQuery().Select("count(*)").
		AndWhere(dbx.HashExp{"email": email}).
		Limit(1)

	if uniqueExcludeIds := list.NonzeroUniques(excludeIds); len(uniqueExcludeIds) > 0 {
		query.AndWhere(dbx.NotIn("id", list.ToInterfaceSlice(uniqueExcludeIds)...))
	}

	var exists bool

	return query.Row(&exists) == nil && !exists
}

// DeleteUser deletes the provided User model.
//
// Returns an error if there is only 1 user.
func (dao *Dao) DeleteUser(user *model.User) error {
	total, err := dao.TotalUsers()
	if err != nil {
		return err
	}

	if total == 1 {
		return errors.New("you cannot delete the only existing user")
	}

	return dao.Delete(user)
}

// SaveUser upserts the provided User model.
func (dao *Dao) SaveUser(user *model.User) error {
	return dao.Save(user)
}
