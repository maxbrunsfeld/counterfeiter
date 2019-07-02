package sql

import (
	"database/sql"
)

//go:generate counterfeiter . DB

type DB interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}
