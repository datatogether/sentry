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
		// if err := initializeDatabase(appDB); err != nil {
		// 	fmt.Println(err.Error())
		// }
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

// drops test data tables & re-inserts base data from sql/test_data.sql, based on
// passed in table names
// func InsertTestData(db *sql.DB, tables ...string) error {
// 	schema, err := dotsql.LoadFromFile()
// 	if err != nil {
// 		return err
// 	}
// 	for _, t := range tables {
// 		if _, err := schema.Exec(db, fmt.Sprintf("insert-%s", t)); err != nil {
// 			err = fmt.Errorf("error insert-%s: %s", t, err.Error())
// 			return err
// 		}
// 	}
// 	return nil
// }
