package core

import (
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
)

func ListCustomCrawls(store datastore.Datastore, limit, offset int) ([]*CustomCrawl, error) {
	q := query.Query{
		Prefix: CustomCrawl{}.DatastoreType(),
		Limit:  limit,
		Offset: offset,
	}

	res, err := store.Query(q)
	if err != nil {
		return nil, err
	}

	uncrawlables := make([]*CustomCrawl, limit)
	i := 0
	for r := range res.Next() {
		if r.Error != nil {
			return nil, err
		}
		c, ok := r.Value.(*CustomCrawl)
		if !ok {
			return nil, ErrInvalidResponse
		}

		uncrawlables[i] = c
		i++
	}

	return uncrawlables[:i], nil
}

// func UnmarshalBoundedCustomCrawls(rows *sql.Rows, limit int) ([]*CustomCrawl, error) {
//  defer rows.Close()
//  subuncrawlables := make([]*CustomCrawl, limit)
//  i := 0
//  for rows.Next() {
//    u := &CustomCrawl{}
//    if err := u.UnmarshalSQL(rows); err != nil {
//      return nil, err
//    }
//    subuncrawlables[i] = u
//    i++
//  }

//  return subuncrawlables[:i], nil
// }
