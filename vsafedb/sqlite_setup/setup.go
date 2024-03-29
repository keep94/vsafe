// Package sqlite_setup sets up a sqlite database for vsafe app
package sqlite_setup

import (
	"database/sql"
)

// SetUpTables creates all needed tables in database for the vsafe app.
func SetUpTables(tx *sql.Tx) error {
	_, err := tx.Exec("create table if not exists user (id INTEGER PRIMARY KEY AUTOINCREMENT, owner INTEGER, name TEXT, key TEXT, checksum TEXT)")
	if err != nil {
		return err
	}
	_, err = tx.Exec("create unique index if not exists user_name_idx on user (name)")
	if err != nil {
		return err
	}
	_, err = tx.Exec("create table if not exists category (id INTEGER PRIMARY KEY AUTOINCREMENT, owner INTEGER, name TEXT)")
	if err != nil {
		return err
	}
	_, err = tx.Exec("create table if not exists entry (id INTEGER PRIMARY KEY AUTOINCREMENT, owner INTEGER, url TEXT, title TEXT, desc TEXT, uname TEXT, password TEXT, special TEXT, categories TEXT)")
	return err
}
