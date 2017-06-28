# sql_datastore
--
    import "github.com/datatogether/sql_datastore"

sql_datastore is an experimental implementation of the ipfs datastore interface
for sql databases. It's very much a work in progress, born out of a somewhat
special use case of needing to be able to store data in a number of different
places, with the datastore interface as a lowest-common-denominator.

The goal is not a fully-expressive sql database operated through the datastore
interface, this is not possible, or even desired. Instead, this package focuses
on doing the kinds of things one would want to do with a key-value datastore,
requiring implementers to provide a standard set of queries and parameters to
glue everything together. Whenever the datastore interface is not expressive
enough, one can always fall back to standard SQL work.

This implementation leads to a great deal of required boilerplate code to
implement. In the future this package could be expanded to become syntax-aware,
accepting a table name & schema definition for registered models. From here the
sql_datastore package could construct default queries that could be overridden
using the current SQLQuery & SQLParams methods. Before that happens, it's worth
noting that the datastore interface may undergo changes in the near future.

## Usage

```go
var DefaultStore = NewDatastore(nil)
```
Package Level Datastore. Be sure to call SetDB before using!

#### func  Register

```go
func Register(models ...Model) error
```

#### func  SetDB

```go
func SetDB(db *sql.DB)
```
SetDB sets the DefaultStore's DB, must be called for DefaultStore to work

#### type Cmd

```go
type Cmd int
```

Cmd represents a set of standardized SQL queries these abstractions define a
common set of commands that a model can provide to sql_datastore for execution

```go
const (
	// Unknown as default, errored state
	CmdUnknown Cmd = iota
	// starting with DDL statements:
	// CREATE TABLE query
	CmdCreateTable
	// ALTER TABLE statement
	CmdAlterTable
	// DROP TABLE statement
	CmdDropTable
	// SELECT statement that should return a single result
	CmdSelectOne
	// INSERT a single row
	CmdInsertOne
	// UPDATE a single row
	CmdUpdateOne
	// DELETE a single row
	CmdDeleteOne
	// Check if a single row exists
	CmdExistsOne
	// List Query, can return many rows. List is special
	// in that LIMIT & OFFSET params are provided by the query
	// method. See note in Datastore.Query()
	CmdList
)
```

#### type Datastore

```go
type Datastore struct {
	// DB is the underlying DB handler
	// it should be safe for use outside of the
	// Datastore itself
	DB *sql.DB
}
```

Datastore

#### func  NewDatastore

```go
func NewDatastore(db *sql.DB) *Datastore
```
NewDatastore creates a datastore instance Datastores should be pointers.

#### func (*Datastore) Batch

```go
func (ds *Datastore) Batch() (datastore.Batch, error)
```
Batch commands are currently not supported

#### func (Datastore) Delete

```go
func (ds Datastore) Delete(key datastore.Key) error
```
Delete a value from the store

#### func (Datastore) Get

```go
func (ds Datastore) Get(key datastore.Key) (value interface{}, err error)
```
Get a model from the store

#### func (Datastore) Has

```go
func (ds Datastore) Has(key datastore.Key) (exists bool, err error)
```
Check to see if key exists in the db

#### func (Datastore) Put

```go
func (ds Datastore) Put(key datastore.Key, value interface{}) error
```
Put a model in the store

#### func (Datastore) Query

```go
func (ds Datastore) Query(q query.Query) (query.Results, error)
```
Ok, this is nothing more than a first step. In the future it seems datastore
will need to construct these queries, which will require more info (tablename,
expected response schema) from the model. Currently it's required that the
passed-in prefix be equal to DatastoreType() which query will use to determine
what model to ask for a ListCmd

#### func (*Datastore) Register

```go
func (ds *Datastore) Register(models ...Model) error
```
Register one or more models that will be used by this datastore. Must be called
before a model can be manipulated by the store

#### type Model

```go
type Model interface {
	// DatastoreType must return the "type" of object, which is a consistent
	// name for the object being stored. DatastoreType works in conjunction
	// with GetId to construct the key for storage.
	// Since SQL doesn't support the "pathing" aspect of keys, any path
	// values are ignored
	DatastoreType() string
	// GetId should return the cannonical ID for the object.
	GetId() string

	// NewSQLModel must allocate & return a new instance of the
	// model with id set such that GetId returns the passed-in id string
	NewSQLModel(id string) Model

	// SQLQuery gives the datastore the query to execute for a given command type
	// As an example, if CmdSelectOne is passed in, something like
	// "SELECT * FROM table WHERE id = $1"
	// should be returned
	SQLQuery(Cmd) string
	// SQLParams gives the datastore the required params for a given command type
	// As an example if CmdSelectOne is passed in, something like
	// []interface{}{m.id}
	// should be returned.
	SQLParams(Cmd) []interface{}
	// UnmarshalSQL takes it's que from UnmarshalJSON, with the difference that
	// the result of Unmarshaling should be assigned to the receiver.
	UnmarshalSQL(sqlutil.Scannable) error
}
```

Model is the interface that must be implemented to work with sql_datastore.
There are some fairly heavy constraints here. For a working example checkout:
https://github.com/datatogether/archive and have a look at primer.go
