// abstractions for working with postgres databases
package sqlutil

import (
	"database/sql"
	"fmt"
	"time"
)

// Scannable unifies both *sql.Row & *sql.Rows, functions can accept
// Scannable & work with both
type Scannable interface {
	Scan(...interface{}) error
}

// Querable unifies both *sql.DB & *sql.Tx for querying purposes
type Queryable interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// Execable ugrades a read-only interface to be able to
// execute sql DDL statements
type Execable interface {
	Queryable
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type Transactable interface {
	Execable
	Begin() (*sql.Tx, error)
}

// Uniform Database connector
func ConnectToDb(driverName, url string, db *sql.DB) error {
	for i := 0; i < 1000; i++ {
		conn, err := SetupConnection(driverName, url)
		if err != nil {
			fmt.Println(err)
			time.Sleep(time.Second)
			continue
		}

		*db = *conn
		break
	}
	return nil
}

// Sets up a connection with a given postgres db connection string
func SetupConnection(driverName, connString string) (db *sql.DB, err error) {
	db, err = sql.Open(driverName, connString)
	if err != nil {
		return
	}
	if err = db.Ping(); err != nil {
		return
	}
	return
}

// EnsureSeedData runs "EnsureTables", and then injects seed data for any newly-created tables
func EnsureSeedData(db *sql.DB, schemaFilepath, dataFilepath string, tables ...string) (created []string, err error) {
	// test query to check for database schema existence
	sc, err := LoadSchemaCommands(schemaFilepath)
	if err != nil {
		return nil, err
	}

	created, err = sc.Create(db, tables...)
	if err != nil {
		return created, err
	}

	if len(created) > 0 {
		dc, err := LoadDataCommands(dataFilepath)
		if err != nil {
			return created, err
		}

		if err := dc.Reset(db, created...); err != nil {
			return created, err
		}
	}

	return created, nil
}

// EnsureTables checks for table existence, creating them from the schema file if not,
// returning a slice of table names that were created
func EnsureTables(db *sql.DB, schemaFilepath string, tables ...string) ([]string, error) {
	sc, err := LoadSchemaCommands(schemaFilepath)
	if err != nil {
		return nil, fmt.Errorf("error loading schema file: %s", err)
	}
	return sc.Create(db, tables...)
}
