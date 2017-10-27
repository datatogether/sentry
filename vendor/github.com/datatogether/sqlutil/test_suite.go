package sqlutil

import (
	"database/sql"
)

type TestSuite struct {
	DB      *sql.DB
	Schema  *SchemaCommands
	Data    *DataCommands
	Cascade []string
}

type TestSuiteOpts struct {
	DriverName      string
	ConnString      string
	SchemaPath      string
	SchemaSqlString string
	DataPath        string
	DataSqlString   string
	Cascade         []string
}

func InitTestSuite(o *TestSuiteOpts) (*TestSuite, error) {
	ts := &TestSuite{
		Cascade: o.Cascade,
	}

	db, err := SetupConnection(o.DriverName, o.ConnString)
	if err != nil {
		return nil, err
	}
	ts.DB = db

	if o.SchemaPath != "" && o.DataPath != "" {
		sf, err := LoadSchemaCommands(o.SchemaPath)
		if err != nil {
			return nil, err
		}

		ts.Schema = sf

		df, err := LoadDataCommands(o.DataPath)
		if err != nil {
			return nil, err
		}
		ts.Data = df
	} else {
		sf, err := LoadSchemaString(o.SchemaSqlString)
		if err != nil {
			return nil, err
		}
		ts.Schema = sf

		df, err := LoadDataString(o.DataSqlString)
		if err != nil {
			return nil, err
		}
		ts.Data = df
	}

	// TODO - make this work & be awesome
	// if err := ts.InitializeDatabase(db); err != nil {
	//  return nil, nil, nil, err
	// }

	// if err := ts.ResetAll(db); err != nil {
	//  return nil, nil, nil, err
	// }

	if err := ts.Schema.DropAllCreate(ts.DB, ts.Cascade...); err != nil {
		return nil, err
	}

	if err := ts.Data.Reset(ts.DB, ts.Cascade...); err != nil {
		return nil, err
	}

	return ts, nil
}

// CreateAll executes all commands that have the prefix "create"
// TODO - need some form of table-order infrerence
// func (ts *TestSuite) CreateAll() error {
//  for _, cmd := range commandsWithPrefix(s.file, "create") {
//    if _, err := s.file.Exec(ts.DB, cmd); err != nil {
//      return fmt.Errorf("errors executing %s: %s", cmd, err.Error())
//    }
//  }
//  return nil
// }
