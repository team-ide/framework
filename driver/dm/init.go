package dm

import (
	"database/sql"
	"fmt"
	"net/url"
)

func GetDriverName() string {
	return "dm"
}

func GetDialect() string {
	return "dm"
}

func GetDSN(user string, password string, host string, port int, schema string) string {
	address := host
	if port > 0 {
		address = fmt.Sprintf("%s:%d", host, port)
	}
	password = url.PathEscape(password)
	dsn := fmt.Sprintf("dm://%s:%s@%s?charset=utf8", user, password, address)
	if schema != "" {
		dsn += "&schema=" + schema
	}
	return dsn
}

func Open(dsn string) (db *sql.DB, err error) {
	db, err = sql.Open(GetDriverName(), dsn)
	if err != nil {
		return
	}
	return
}
