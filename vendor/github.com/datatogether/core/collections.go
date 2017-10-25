package core

import (
	"database/sql"
	"fmt"
	"github.com/datatogether/sql_datastore"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
)

func ListCollections(store datastore.Datastore, limit, offset int) ([]*Collection, error) {
	q := query.Query{
		Prefix: Collection{}.DatastoreType(),
		Limit:  limit,
		Offset: offset,
	}

	res, err := store.Query(q)
	if err != nil {
		return nil, err
	}

	collections := make([]*Collection, limit)
	i := 0
	for r := range res.Next() {
		if r.Error != nil {
			return nil, err
		}

		c, ok := r.Value.(*Collection)
		if !ok {
			return nil, ErrInvalidResponse
		}

		collections[i] = c
		i++
	}

	return collections[:i], nil
}

func CollectionsByCreator(store datastore.Datastore, creator, orderby string, limit, offset int) ([]*Collection, error) {
	sqls, ok := store.(*sql_datastore.Datastore)
	if !ok {
		return nil, fmt.Errorf("collections for creator only works with SQL datastores for now")
	}

	rows, err := sqls.DB.Query(qCollectionsByCreator, limit, offset, orderby, creator)
	if err != nil {
		return nil, err
	}

	return unmarshalBoundedCollections(rows, limit)
}

// unmarshalBoundedCollections turns a standard sql.Rows of Collection results into a *Collection slice
func unmarshalBoundedCollections(rows *sql.Rows, limit int) ([]*Collection, error) {
	defer rows.Close()
	collections := make([]*Collection, limit)
	i := 0
	for rows.Next() {
		c := &Collection{}
		if err := c.UnmarshalSQL(rows); err != nil {
			return nil, err
		}
		collections[i] = c
		i++
	}

	return collections[:i], nil
}
