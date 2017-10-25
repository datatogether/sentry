package sql_datastore

import (
	"fmt"
	"github.com/ipfs/go-datastore/query"
)

// Order a query by a field
type OrderBy string

// String value, used to inject the field name istself as a SQL query param
func (o OrderBy) String() string {
	return string(o)
}

// satisfy datastore.Order interface, this is a no-op b/c sql sorting
// will happen at query-time
// TODO -  In the future this should be generalized to facilitate supplying
// sql_datastore.OrderBy orders to other datastores, providing parity
func (o OrderBy) Sort([]query.Entry) {}

// Order a query by a field, descending
type OrderByDesc string

// String value, used to inject the field name istself as a SQL query param
func (o OrderByDesc) String() string {
	return fmt.Sprintf("%s DESC", o)
}

// satisfy datastore.Order interface, this is a no-op b/c sql sorting
// will happen at query-time
// TODO -  In the future this should be generalized to facilitate supplying
// sql_datastore.OrderBy orders to other datastores, providing parity
func (o OrderByDesc) Sort([]query.Entry) {}
