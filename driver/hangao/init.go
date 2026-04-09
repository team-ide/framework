package hangao

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"net/url"
)

func GetDriverName() string {
	return "postgres"
}

func GetDialect() string {
	return "hangao"
}

func GetDSN(user string, password string, host string, port int, database string) string {
	password = url.PathEscape(password)
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable", user, password, host, port, database)
	return dsn
}

func Open(dsn string) (db *sql.DB, err error) {
	db, err = sql.Open(GetDriverName(), dsn)
	if err != nil {
		return
	}
	return
}
