package core

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/har4s/ohmygo/dbx"
)

func connectDB(dsn string) (*dbx.DB, error) {
	db, err := dbx.Open("mysql", dsn)

	if err != nil {
		return nil, err
	}

	if err := db.DB().Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
