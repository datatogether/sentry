# sqlutil
--
    import "github.com/datatogether/sqlutil"

utils for working with dotsql structs

abstractions for working with postgres databases

## Usage

#### func  ConnectToDb

```go
func ConnectToDb(driverName, url string, db *sql.DB) error
```
Uniform Database connector

#### func  EnsureSeedData

```go
func EnsureSeedData(db *sql.DB, schemaFilepath, dataFilepath string, tables ...string) (err error)
```
EnsureSeedData runs "EnsureTables", and then injects seed data for any
newly-created tables

#### func  EnsureTables

```go
func EnsureTables(db *sql.DB, schemaFilepath string, tables ...string) ([]string, error)
```
EnsureTables checks for table existence, creating them from the schema file if
not, returning a slice of table names that were created

#### func  SetupConnection

```go
func SetupConnection(driverName, connString string) (db *sql.DB, err error)
```
Sets up a connection with a given postgres db connection string

#### type DataCommands

```go
type DataCommands struct {
}
```

SchemaFile is an sql file that defines a database schema

#### func  LoadDataCommands

```go
func LoadDataCommands(sqlFilePath string) (*DataCommands, error)
```
LoadDataCommands takes a filepath to a sql file with create & drop table
commands and returns a DataCommands

#### func  LoadDataString

```go
func LoadDataString(sql string) (*DataCommands, error)
```

#### func (*DataCommands) Commands

```go
func (d *DataCommands) Commands() []string
```

#### func (*DataCommands) DeleteAll

```go
func (d *DataCommands) DeleteAll(db Execable) error
```
DropAll executes the command named "drop-all" from the sql file this should be a
command in the form: DROP TABLE IF EXISTS foo, bar, baz ...

#### func (*DataCommands) Reset

```go
func (d *DataCommands) Reset(db Execable, tables ...string) error
```

#### type Execable

```go
type Execable interface {
	Queryable
	Exec(query string, args ...interface{}) (sql.Result, error)
}
```

Execable ugrades a read-only interface to be able to execute sql DDL statements

#### type Queryable

```go
type Queryable interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}
```

Querable unifies both *sql.DB & *sql.Tx for querying purposes

#### type Scannable

```go
type Scannable interface {
	Scan(...interface{}) error
}
```

Scannable unifies both *sql.Row & *sql.Rows, functions can accept Scannable &
work with both

#### type SchemaCommands

```go
type SchemaCommands struct {
}
```

SchemaCommands is an sql file that defines a database schema

#### func  LoadSchemaCommands

```go
func LoadSchemaCommands(sqlFilePath string) (*SchemaCommands, error)
```
LoadSchemaCommands takes a filepath to a sql file with create & drop table
commands and returns a SchemaCommands

#### func  LoadSchemaString

```go
func LoadSchemaString(sql string) (*SchemaCommands, error)
```

#### func (*SchemaCommands) Create

```go
func (s *SchemaCommands) Create(db Execable, tables ...string) ([]string, error)
```
Create tables if they don't already exist

#### func (*SchemaCommands) DropAll

```go
func (s *SchemaCommands) DropAll(db Execable) error
```
DropAll executes the command named "drop-all" from the sql file this should be a
command in the form: DROP TABLE IF EXISTS foo, bar, baz ...

#### func (*SchemaCommands) DropAllCreate

```go
func (s *SchemaCommands) DropAllCreate(db Execable, tables ...string) error
```

#### type TestSuite

```go
type TestSuite struct {
	DB      *sql.DB
	Schema  *SchemaCommands
	Data    *DataCommands
	Cascade []string
}
```


#### func  InitTestSuite

```go
func InitTestSuite(o *TestSuiteOpts) (*TestSuite, error)
```

#### type TestSuiteOpts

```go
type TestSuiteOpts struct {
	DriverName      string
	ConnString      string
	SchemaPath      string
	SchemaSqlString string
	DataPath        string
	DataSqlString   string
	Cascade         []string
}
```


#### type Transactable

```go
type Transactable interface {
	Execable
	Begin() (*sql.Tx, error)
}
```
