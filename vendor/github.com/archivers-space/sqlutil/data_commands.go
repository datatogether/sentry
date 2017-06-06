package sqlutil

import (
	"fmt"
	"github.com/gchaincl/dotsql"
)

// LoadDataCommands takes a filepath to a sql file with create & drop table commands
// and returns a DataCommands
func LoadDataCommands(sqlFilePath string) (*DataCommands, error) {
	f, err := dotsql.LoadFromFile(sqlFilePath)
	if err != nil {
		return nil, err
	}

	return &DataCommands{
		file: f,
	}, nil
}

func LoadDataString(sql string) (*DataCommands, error) {
	f, err := dotsql.LoadFromString(sql)
	if err != nil {
		return nil, err
	}

	return &DataCommands{
		file: f,
	}, nil
}

// SchemaFile is an sql file that defines a database schema
type DataCommands struct {
	file *dotsql.DotSql
}

func (d *DataCommands) Commands() []string {
	return commandsWithPrefix(d.file, "")
}

// DropAll executes the command named "drop-all" from the sql file
// this should be a command in the form:
// DROP TABLE IF EXISTS foo, bar, baz ...
func (d *DataCommands) DeleteAll(db Execable) error {
	for _, cmd := range commandsWithPrefix(d.file, "delete") {
		if _, err := d.file.Exec(db, cmd); err != nil {
			return fmt.Errorf("error executing '%s': %s", cmd, err)
		}
	}
	return nil
}

func (d *DataCommands) Reset(db Execable, tables ...string) error {
	for _, t := range tables {
		if _, err := d.file.Exec(db, fmt.Sprintf("delete-%s", t)); err != nil {
			return fmt.Errorf("error executing 'delete-%s': %s", t, err)
		}

		if _, err := d.file.Exec(db, fmt.Sprintf("insert-%s", t)); err != nil {
			return fmt.Errorf("error executing 'insert-%s': %s", t, err)
		}
	}
	return nil
}

// CreateAll executes all commands that have the prefix "create"
// func (d *DataCommands) InsertAll(db Execable) error {
// 	for _, cmd := range commandsWithPrefix(d.file, "insert") {
// 		if _, err := d.file.Exec(db, cmd); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func (d *DataCommands) ResetAll(db Execable) error {
// 	if err := d.DeleteAll(db); err != nil {
// 		return err
// 	}

// 	if err := d.InsertAll(db); err != nil {
// 		return err
// 	}

// 	return nil
// }
