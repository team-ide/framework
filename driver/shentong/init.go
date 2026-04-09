package shentong

import (
	"database/sql"
	"fmt"
	_ "github.com/team-ide/framework/driver/shentong/go-aci"
	"net/url"
)

func GetDriverName() string {
	return "aci"
}

func GetDialect() string {
	return "shentong"
}

func GetDSN(user string, password string, host string, port int, database string) string {
	password = url.PathEscape(password)
	dsn := fmt.Sprintf("%s/%s@%s:%d/%s", user, password, host, port, database)
	return dsn
}

func Open(dsn string) (db *sql.DB, err error) {
	db, err = sql.Open(GetDriverName(), dsn)
	if err != nil {
		return
	}
	return
}
