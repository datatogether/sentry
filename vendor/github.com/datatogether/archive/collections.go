package archive

import (
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
