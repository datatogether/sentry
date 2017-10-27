package core

import (

	// "database/sql"
	// "github.com/datatogether/sqlutil"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
)

func ListUncrawlables(store datastore.Datastore, limit, offset int) ([]*Uncrawlable, error) {
	q := query.Query{
		Prefix: Uncrawlable{}.DatastoreType(),
		Limit:  limit,
		Offset: offset,
	}

	res, err := store.Query(q)
	if err != nil {
		return nil, err
	}

	uncrawlables := make([]*Uncrawlable, limit)
	i := 0
	for r := range res.Next() {
		if r.Error != nil {
			return nil, err
		}
		c, ok := r.Value.(*Uncrawlable)
		if !ok {
			return nil, ErrInvalidResponse
		}

		uncrawlables[i] = c
		i++
	}

	return uncrawlables[:i], nil
}

// func UnmarshalBoundedUncrawlables(rows *sql.Rows, limit int) ([]*Uncrawlable, error) {
// 	defer rows.Close()
// 	subuncrawlables := make([]*Uncrawlable, limit)
// 	i := 0
// 	for rows.Next() {
// 		u := &Uncrawlable{}
// 		if err := u.UnmarshalSQL(rows); err != nil {
// 			return nil, err
// 		}
// 		subuncrawlables[i] = u
// 		i++
// 	}

// 	return subuncrawlables[:i], nil
// }
