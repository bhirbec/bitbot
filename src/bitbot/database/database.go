package database

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	"bitbot/errorutils"
)

// TODO: remove this struct?
type DB struct {
	*sql.DB
}

func Open(name, host, port, user, pwd string) *DB {
	source := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8", user, pwd, host, port, name)
	db, err := sql.Open("mysql", source)
	errorutils.PanicOnError(err)
	return &DB{db}
}
