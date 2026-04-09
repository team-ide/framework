package odbc

import (
	"database/sql"
)

func GetDriverName() string {
	return "odbc"
}

func Open(dsn string) (db *sql.DB, err error) {
	db, err = sql.Open(GetDriverName(), dsn)
	if err != nil {
		return
	}
	return
}
