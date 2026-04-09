package sqlite3

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func GetDriverName() string {
	return "sqlite3"
}

func GetDialect() string {
	return "sqlite3"
}

func GetDSN(databasePath string) string {
	dsn := databasePath
	return dsn
}

func Open(dsn string) (db *sql.DB, err error) {
	db, err = sql.Open(GetDriverName(), dsn)
	if err != nil {
		return
	}
	return
}
