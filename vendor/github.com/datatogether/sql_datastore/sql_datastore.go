// sql_datastore is an experimental implementation of the ipfs datastore
// interface for sql databases. It's very much a work in progress,
// born out of a somewhat special use case of needing to be able to store
// data in a number of different places, with the datastore interface as
// a lowest-common-denominator.
//
// The goal is not a fully-expressive sql database operated through
// the datastore interface, this is not possible, or even desired.
// Instead, this package focuses on doing the kinds of things one
// would want to do with a key-value datastore, requiring implementers
// to provide a standard set of queries and parameters to glue everything
// together. Whenever the datastore interface is not expressive enough,
// one can always fall back to standard SQL work.
//
// sql_datastore reconciles the key-value orientation of the datastore
// interface with the tables/relational orientation of SQL databases
// through the concept of a "Model". Model is a bit of an unfortunate name,
// as it implies this package is an ORM, which isn't a design goal.
//
// Annnnnnnnyway, the important patterns of this approach are:
//     1. The Model interface defines how to get stuff into and out of SQL
//     2. All Models that will be interacted with must be "Registered" to the store.
//        Registered Models map to a datastore.Key Type.
//     3. All Get/Put/Delete/Has/Query to sql_datastore must map to a single Model
//
// This implementation leads to a great deal of required boilerplate code
// to implement. In the future this package could be expanded to become
// syntax-aware, accepting a table name & schema definition for registered
// models. From here the sql_datastore package could construct default queries
// that could be overridden using the current SQLQuery & SQLParams methods.
// Before that happens, it's worth noting that the datastore interface may
// undergo changes in the near future.
package sql_datastore

import (
	"database/sql"
	"fmt"
	datastore "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
)

// Package Level Datastore. Be sure to call SetDB before using!
var DefaultStore = NewDatastore(nil)

// SetDB sets the DefaultStore's DB, must be called for DefaultStore
// to work
func SetDB(db *sql.DB) {
	DefaultStore.DB = db
}

// Register a number of models to the DefaultStore
func Register(models ...Model) error {
	return DefaultStore.Register(models...)
}

// Datastore implements the ipfs datastore interface for SQL databases
type Datastore struct {
	// DB is the underlying DB handler
	// it should be safe for use outside of the
	// Datastore itself
	DB *sql.DB
	// Slice of models that have been registered to this
	// datastore
	models []Model
}

// NewDatastore creates a datastore instance
// Datastores should be pointers.
func NewDatastore(db *sql.DB) *Datastore {
	return &Datastore{DB: db}
}

// Register one or more models that will be used by this datastore.
// Must be called before a model can be manipulated by the store
func (ds *Datastore) Register(models ...Model) error {
	for _, model := range models {
		// TODO - sanity check to make sure the model behaves.
		// return error if not
		ds.models = append(ds.models, model)
	}
	return nil
}

// Put a model in the store
func (ds Datastore) Put(key datastore.Key, value interface{}) error {
	sqlModelValue, ok := value.(Model)
	if !ok {
		return fmt.Errorf("value is not a valid sql model")
	}

	exists, err := ds.hasModel(sqlModelValue)
	if err != nil {
		return err
	}

	if exists {
		return ds.exec(sqlModelValue, CmdUpdateOne)
	} else {
		return ds.exec(sqlModelValue, CmdInsertOne)
	}
}

// Get a model from the store
func (ds Datastore) Get(key datastore.Key) (value interface{}, err error) {
	m, err := ds.modelForKey(key)
	if err != nil {
		return nil, err
	}

	row, err := ds.queryRow(m, CmdSelectOne)
	if err != nil {
		return nil, err
	}

	v := m.NewSQLModel(key)
	if err := v.UnmarshalSQL(row); err != nil {
		return nil, err
	}
	return v, nil
}

// Check to see if key exists in the db
func (ds Datastore) Has(key datastore.Key) (exists bool, err error) {
	m, err := ds.modelForKey(key)
	if err != nil {
		return false, err
	}

	row, err := ds.queryRow(m, CmdExistsOne)
	if err != nil {
		return false, err
	}

	err = row.Scan(&exists)
	return
}

// Delete a value from the store
func (ds Datastore) Delete(key datastore.Key) error {
	m, err := ds.modelForKey(key)
	if err != nil {
		return err
	}

	return ds.exec(m, CmdDeleteOne)
}

// Ok, this is nothing more than a first step. In the future
// it seems datastore will need to construct these queries, which
// will require more info (tablename, expected response schema) from
// the model.
// Currently it's required that the passed-in prefix be equal to DatastoreType()
// which query will use to determine what model to ask for a ListCmd
func (ds Datastore) Query(q query.Query) (query.Results, error) {
	var rows *sql.Rows

	// TODO - support KeysOnly
	if q.KeysOnly {
		return nil, fmt.Errorf("sql datastore doesn't support keysonly querying")
	}

	// determine what type of model we're querying for
	m, err := ds.modelForQuery(q)
	if err != nil {
		return nil, err
	}

	// here we attach query information to a new model
	// TODO - currently this is basing off of the prefix, might need
	// to do smarter things in relation to filters?
	m = m.NewSQLModel(datastore.NewKey(q.Prefix))

	// This is totally janky, but will work for now. It's expected that
	// the returned CmdList will have at least 2 bindvars:
	// $1 : LIMIT value
	// $2 : OFFSET value
	// if compatible Orders are provided to the query, the order string
	// will be provided as a third bindvar:
	// $3 : ORDERBY value
	// From there it can provide zero or more additional bindvars to
	// organize the query, which should be returned by the SQLParams method
	// TODO - this seems to hint at a need for some sort of Controller-like
	// pattern in userland. Have a think.
	os := orderString(q)
	if os != "" {
		rows, err = ds.query(m, CmdList, q.Limit, q.Offset, os)
		if err != nil {
			return nil, err
		}
	} else {
		rows, err = ds.query(m, CmdList, q.Limit, q.Offset)
		if err != nil {
			return nil, err
		}
	}

	// TODO - should this be q.Limit or query.NormalBufferSize?
	reschan := make(chan query.Result, q.Limit)
	go func() {
		defer close(reschan)

		for rows.Next() {

			model := m.NewSQLModel(datastore.NewKey(q.Prefix))
			if err := model.UnmarshalSQL(rows); err != nil {
				reschan <- query.Result{
					Error: err,
				}
			}

			reschan <- query.Result{
				Entry: query.Entry{
					Key:   m.GetId(),
					Value: model,
				},
			}

		}
	}()

	return query.ResultsWithChan(q, reschan), nil
}

// modelForQuery determines what type of model this query should return
func (ds Datastore) modelForQuery(q query.Query) (Model, error) {
	for _, f := range q.Filters {
		switch t := f.(type) {
		case FilterKeyTypeEq:
			return ds.modelForKey(t.Key())
		}
	}
	return ds.modelForKey(datastore.NewKey(fmt.Sprintf("/%s:", q.Prefix)))
}

// orderString generates orders from any OrderBy or OrderByDesc values
func orderString(q query.Query) string {
	if len(q.Orders) == 0 {
		return ""
	}
	orders := []string{}
	for _, o := range q.Orders {
		// TODO - should this cross-correct by casting based on
		// the outer order?
		switch ot := o.(type) {
		case query.OrderByValue:
			switch obv := ot.TypedOrder.(type) {
			case OrderBy:
				orders = append(orders, obv.String())
			case OrderByDesc:
				orders = append(orders, obv.String())
			}
		case query.OrderByValueDescending:
			switch obv := ot.TypedOrder.(type) {
			case OrderBy:
				orders = append(orders, obv.String())
			case OrderByDesc:
				orders = append(orders, obv.String())
			}
		}
	}
	os := ""
	for i, o := range orders {
		if i == 0 {
			os += o
		} else {
			os += fmt.Sprintf(", %s", o)
		}
	}
	return os
}

// Batch commands are currently not supported
func (ds *Datastore) Batch() (datastore.Batch, error) {
	return nil, datastore.ErrBatchUnsupported
}

// for a given key, determine what kind of Model we're looking for
func (ds Datastore) modelForKey(key datastore.Key) (Model, error) {
	for _, m := range ds.models {
		if m.DatastoreType() == key.Type() {
			// return a model with "ID" set to the key param
			return m.NewSQLModel(key), nil
		}
	}
	return nil, fmt.Errorf("no usable model found for key, did you call register on the model?: %s", key.String())
}

// does this datastore
func (ds Datastore) hasModel(m Model) (exists bool, err error) {
	row, err := ds.queryRow(m, CmdExistsOne)
	if err != nil {
		return false, err
	}
	err = row.Scan(&exists)
	return
}

// execute a Cmd against a given model
func (ds Datastore) exec(m Model, t Cmd) error {
	if ds.DB == nil {
		return fmt.Errorf("datastore has no DB")
	}
	query, params, err := ds.prepQuery(m, t)
	if err != nil {
		return err
	}
	_, err = ds.DB.Exec(query, params...)
	return err
}

// query for a single row, given a type of command and model
func (ds Datastore) queryRow(m Model, t Cmd) (*sql.Row, error) {
	if ds.DB == nil {
		return nil, fmt.Errorf("datastore has no DB")
	}
	query, params, err := ds.prepQuery(m, t)
	if err != nil {
		return nil, err
	}
	return ds.DB.QueryRow(query, params...), nil
}

// run a query against the db for a given command and model, with optionally prebound
// arguments derived from the query
func (ds Datastore) query(m Model, t Cmd, prebind ...interface{}) (*sql.Rows, error) {
	if ds.DB == nil {
		return nil, fmt.Errorf("datastore has no DB")
	}
	query, params, err := ds.prepQuery(m, t)
	if err != nil {
		return nil, err
	}

	return ds.DB.Query(query, append(prebind, params...)...)
}

// prepare a query, grabbing the command sql & params from the model
func (ds Datastore) prepQuery(m Model, t Cmd) (string, []interface{}, error) {
	query := m.SQLQuery(t)
	if query == "" {
		// TODO - make Cmd satisfy stringer, provide better error
		return "", nil, fmt.Errorf("missing required command: %d", t)
	}
	params := m.SQLParams(t)
	return query, params, nil
}
