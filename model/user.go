package model

import (
	"errors"

	"github.com/har4s/ohmygo/tools/security"
	"github.com/har4s/ohmygo/tools/types"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	BaseModel

	Email           string         `db:"email" json:"email"`
	TokenKey        string         `db:"tokenKey" json:"-"`
	PasswordHash    string         `db:"passwordHash" json:"-"`
	IsAdmin         bool           `db:"isAdmin" json:"-"`
	IsSuperadmin    bool           `db:"isSuperadmin" json:"-"`
	LastResetSentAt types.DateTime `db:"lastResetSentAt" json:"-"`
}

// TableName returns the User model SQL table name.
func (m *User) TableName() string {
	return "users"
}

// ValidatePassword validates a plain password against the model's password.
func (m *User) ValidatePassword(password string) bool {
	bytePassword := []byte(password)
	bytePasswordHash := []byte(m.PasswordHash)

	// comparing the password with the hash
	err := bcrypt.CompareHashAndPassword(bytePasswordHash, bytePassword)

	// nil means it is a match
	return err == nil
}

// SetPassword sets cryptographically secure string to `model.Password`.
//
// Additionally this method also resets the LastResetSentAt and the TokenKey fields.
func (m *User) SetPassword(password string) error {
	if password == "" {
		return errors.New("the provided plain password is empty")
	}

	// hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 13)
	if err != nil {
		return err
	}

	m.PasswordHash = string(hashedPassword)
	m.LastResetSentAt = types.DateTime{} // reset

	// invalidate previously issued tokens
	return m.RefreshTokenKey()
}

// RefreshTokenKey generates and sets new random token key.
func (m *User) RefreshTokenKey() error {
	m.TokenKey = security.RandomString(50)
	return nil
}

// RefreshLastResetSentAt updates the user LastResetSentAt field with the current datetime.
func (m *User) RefreshLastResetSentAt() {
	m.LastResetSentAt = types.NowDateTime()
}
