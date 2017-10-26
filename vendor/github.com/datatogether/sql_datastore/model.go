package sql_datastore

import (
	"github.com/datatogether/sqlutil"
	datastore "github.com/ipfs/go-datastore"
)

// Model is the interface that must be implemented to work
// with sql_datastore. There are some fairly heavy constraints here.
// For a working example checkout: https://github.com/archivers-space/archive
// and have a look at primer.go
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
