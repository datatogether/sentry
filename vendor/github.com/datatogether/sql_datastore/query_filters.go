package sql_datastore

import (
	"fmt"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
)

// FilterTypeEq filters for a specific key Type (which should match a registerd model on the sql_datastore.Datastore)
// FilterTypeEq is a string that specifies the key type we're after
type FilterKeyTypeEq string

// Key return s FilterKeyTypeEq formatted as a datastore.Key
func (f FilterKeyTypeEq) Key() datastore.Key {
	return datastore.NewKey(fmt.Sprintf("/%s:", f.String()))
}

// Satisfy the Stringer interface
func (f FilterKeyTypeEq) String() string {
	return string(f)
}

// Filter satisfies the query.Filter interface
// TODO - make this work properly for the sake of other datastores
func (f FilterKeyTypeEq) Filter(e query.Entry) bool {
	return true
}
