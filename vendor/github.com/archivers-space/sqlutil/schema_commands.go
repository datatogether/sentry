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

func (s *SchemaCommands) Create(db Execable, tables ...string) error {
	for _, t := range tables {
		cmd := fmt.Sprintf("create-%s", t)
		if _, err := s.file.Exec(db, cmd); err != nil {
			return fmt.Errorf("error executing '%s': %s", cmd, err.Error())
		}
	}
	return nil
}

func (s *SchemaCommands) DropAllCreate(db Execable, tables ...string) error {
	if err := s.DropAll(db); err != nil {
		return err
	}
	if err := s.Create(db, tables...); err != nil {
		return err
	}
	return nil
}

// InitializeDatabase drops everything and calls create on all tables
// WARNING - THIS ZAPS WHATEVER DB IT'S GIVEN. DO NOT CALL THIS SHIT.
// used for testing only, returns a teardown func
// func (s *SchemaCommands) InitializeDatabase(db Execable) error {
// 	// TODO - infer table names from de-prefixed create commands,
// 	// use this to check for data existence
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
