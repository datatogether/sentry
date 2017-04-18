// abstractions for working with postgres databases
package archive

import (
	"database/sql"
	_ "github.com/lib/pq"
)

// this interface unifies both *sql.Row & *sql.Rows
type sqlScannable interface {
	Scan(...interface{}) error
}

// sqlQuerable unifies both *sql.DB & *sql.Tx for querying purposes
type sqlQueryable interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

type sqlExecable interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type sqlQueryExecable interface {
	sqlQueryable
	sqlExecable
}

// Sets up a connection with a given postgres db connection string
func SetupConnection(connString string) (db *sql.DB, err error) {
	db, err = sql.Open("postgres", connString)
	if err != nil {
		return
	}
	if err = db.Ping(); err != nil {
		return
	}
	return
}
