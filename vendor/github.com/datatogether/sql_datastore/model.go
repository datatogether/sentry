package sql_datastore

import (
	"github.com/datatogether/sqlutil"
)

// Model is the interface that must be implemented to work
// with sql_datastore. There are some fairly heavy constraints here.
// For a working example checkout: https://github.com/archivers-space/archive
// and have a look at primer.go
type Model interface {
	// DatastoreType must return the "type" of object, which is a consistent
	// name for the object being stored. DatastoreType works in conjunction
	// with GetId to construct the key for storage.
	// Since SQL doesn't support the "pathing" aspect of keys, any path
	// values are ignored
	DatastoreType() string
	// GetId should return the cannonical ID for the object.
	GetId() string

	// While not explicitly required by this package, most implementations
	// will want to have a "Key" method that combines DatastoreType() and GetId()
	// to form a key that can be provided to Get & Has commands
	// eg:
	// func (m) Key() datastore.Key {
	// 	return datastore.NewKey(fmt.Sprintf("%s:%s", m.DatastoreType(), m.GetId()))
	// }

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
