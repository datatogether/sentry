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

sql_datastore reconciles the key-value orientation of the datastore interface
with the tables/relational orientation of SQL databases through the concept of a
"Model". Model is a bit of an unfortunate name, as it implies this package is an
ORM, which isn't a design goal.

Annnnnnnnyway, the important patterns of this approach are:

    1. The Model interface defines how to get stuff into and out of SQL
    2. All Models that will be interacted with must be "Registered" to the store.
       Registered Models map to a datastore.Key Type.
    3. All Get/Put/Delete/Has/Query to sql_datastore must map to a single Model

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
Register a number of models to the DefaultStore

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
common set of commands that a model can provide to sql_datastore for execution.

```go
const (
	// Unknown as default, errored state. CmdUnknown should never
	// be intentionally passed to... anything.
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

Datastore implements the ipfs datastore interface for SQL databases

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

#### type FilterKeyTypeEq

```go
type FilterKeyTypeEq string
```

FilterTypeEq filters for a specific key Type (which should match a registerd
model on the sql_datastore.Datastore) FilterTypeEq is a string that specifies
the key type we're after

#### func (FilterKeyTypeEq) Filter

```go
func (f FilterKeyTypeEq) Filter(e query.Entry) bool
```
Filter satisfies the query.Filter interface TODO - make this work properly for
the sake of other datastores

#### func (FilterKeyTypeEq) Key

```go
func (f FilterKeyTypeEq) Key() datastore.Key
```
Key return s FilterKeyTypeEq formatted as a datastore.Key

#### func (FilterKeyTypeEq) String

```go
func (f FilterKeyTypeEq) String() string
```
Satisfy the Stringer interface

#### type Model

```go
type Model interface {
	// DatastoreType must return the "type" of object, which is a consistent
	// name for the object being stored. DatastoreType works in conjunction
	// with GetId to construct the key for storage.
	DatastoreType() string
	// GetId should return the standalone cannonical ID for the object.
	GetId() string

	// Key is a methoda that traditionally combines DatastoreType() and GetId()
	// to form a key that can be provided to Get & Has commands
	// eg:
	// func (m) Key() datastore.Key {
	// 	return datastore.NewKey(fmt.Sprintf("%s:%s", m.DatastoreType(), m.GetId()))
	// }
	// in examples of "submodels" of another model it makes sense to leverage the
	// POSIX structure of keys. for example:
	// func (m) Key() datastore.Key {
	// 	return datastore.NewKey(fmt.Sprintf("%s:%s/%s", m.DatastoreType(), m.ParentId(), m.GetId()))
	// }
	Key() datastore.Key

	// NewSQLModel must allocate & return a new instance of the
	// model with id set such that GetId returns the passed-in Key
	// NewSQLModel will be passed keys for creation of new blank models
	NewSQLModel(key datastore.Key) Model

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
https://github.com/archivers-space/archive and have a look at primer.go

#### type OrderBy

```go
type OrderBy string
```

Order a query by a field

#### func (OrderBy) Sort

```go
func (o OrderBy) Sort([]query.Entry)
```
satisfy datastore.Order interface, this is a no-op b/c sql sorting will happen
at query-time TODO - In the future this should be generalized to facilitate
supplying sql_datastore.OrderBy orders to other datastores, providing parity

#### func (OrderBy) String

```go
func (o OrderBy) String() string
```
String value, used to inject the field name istself as a SQL query param

#### type OrderByDesc

```go
type OrderByDesc string
```

Order a query by a field, descending

#### func (OrderByDesc) Sort

```go
func (o OrderByDesc) Sort([]query.Entry)
```
satisfy datastore.Order interface, this is a no-op b/c sql sorting will happen
at query-time TODO - In the future this should be generalized to facilitate
supplying sql_datastore.OrderBy orders to other datastores, providing parity

#### func (OrderByDesc) String

```go
func (o OrderByDesc) String() string
```
String value, used to inject the field name istself as a SQL query param
