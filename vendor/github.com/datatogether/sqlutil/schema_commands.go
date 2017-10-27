package sqlutil

import (
	"fmt"
	"github.com/gchaincl/dotsql"
)

// LoadSchemaCommands takes a filepath to a sql file with create & drop table commands
// and returns a SchemaCommands
func LoadSchemaCommands(sqlFilePath string) (*SchemaCommands, error) {
	f, err := dotsql.LoadFromFile(sqlFilePath)
	if err != nil {
		return nil, err
	}

	return &SchemaCommands{
		file: f,
	}, nil
}

func LoadSchemaString(sql string) (*SchemaCommands, error) {
	f, err := dotsql.LoadFromString(sql)
	if err != nil {
		return nil, err
	}

	return &SchemaCommands{
		file: f,
	}, nil
}

// SchemaCommands is an sql file that defines a database schema
type SchemaCommands struct {
	file *dotsql.DotSql
}

// DropAll executes the command named "drop-all" from the sql file
// this should be a command in the form:
// DROP TABLE IF EXISTS foo, bar, baz ...
func (s *SchemaCommands) DropAll(db Execable) error {
	_, err := s.file.Exec(db, "drop-all")
	if err != nil {
		fmt.Errorf("error executing 'drop-all': %s", err.Error())
	}
	return nil
}

// Create tables if they don't already exist
func (s *SchemaCommands) Create(db Execable, tables ...string) ([]string, error) {
	created := []string{}
	for _, table := range tables {
		exists := false
		if err := db.QueryRow(fmt.Sprintf("select exists(select 1 from %s limit 1)", table)).Scan(&exists); err == nil || exists {
			continue
		}
		cmd := fmt.Sprintf("create-%s", table)
		if _, err := s.file.Exec(db, cmd); err != nil {
			return created, fmt.Errorf("error executing '%s': %s", cmd, err.Error())
		}
		created = append(created, table)
	}
	return created, nil
}

func (s *SchemaCommands) DropAllCreate(db Execable, tables ...string) error {
	if err := s.DropAll(db); err != nil {
		return err
	}
	if _, err := s.Create(db, tables...); err != nil {
		return err
	}
	return nil
}

// // InitializeDatabase drops everything and calls create on all tables
// func (s *SchemaCommands) InitializeDatabase(db Execable,) error {
// 	// // test query to check for database schema existence
// 	// var exists bool
// 	// if err = db.QueryRow("select exists(select * from primers limit 1)").Scan(&exists); err == nil {
// 	//   return nil
// 	// }

// 	if err := s.DropAll(db); err != nil {
// 		return err
// 	}

// 	if err := s.CreateAll(db); err != nil {
// 		return err
// 	}

// 	return nil
// }
