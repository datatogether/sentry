package core

import (
	"github.com/datatogether/sql_datastore"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
)

// ItemCount gets the number of items in the collection
func (c *Collection) ItemCount(store datastore.Datastore) (count int, err error) {
	if sqls, ok := store.(*sql_datastore.Datastore); ok {
		row := sqls.DB.QueryRow(qCollectionLength, c.Id)
		err = row.Scan(&count)
		return
	}

	// TODO - untested code :(
	res, err := store.Query(query.Query{
		Prefix:   c.Key().String(),
		KeysOnly: true,
	})
	if err != nil {
		return 0, err
	}

	for r := range res.Next() {
		if r.Error != nil {
			return 0, err
		}
		if _, ok := r.Value.(*CollectionItem); ok {
			count++
		}
	}

	return
}

// SaveItems saves a slice of items to the collection.
// It's up to you to ensure that the "index" param doesn't get all messed up.
// TODO - validate / automate the Index param?
func (c *Collection) SaveItems(store datastore.Datastore, items []*CollectionItem) error {
	for _, item := range items {
		item.collectionId = c.Id
		if err := item.Save(store); err != nil {
			return err
		}
	}
	return nil
}

// DeleteItems removes a given list of items from the collection
func (c *Collection) DeleteItems(store datastore.Datastore, items []*CollectionItem) error {
	for _, item := range items {
		item.collectionId = c.Id
		if err := item.Delete(store); err != nil {
			return err
		}
	}
	return nil
}

// ReadItems reads a bounded set of items from the collection
// the orderby param currently only supports SQL-style input of a single proprty, eg: "index" or "index DESC"
func (c *Collection) ReadItems(store datastore.Datastore, orderby string, limit, offset int) (items []*CollectionItem, err error) {
	items = make([]*CollectionItem, limit)

	res, err := store.Query(query.Query{
		Limit:  limit,
		Offset: offset,
		// Keeping in mind that CollectionItem keys take the form /Collection:[id]/CollectionItem:[id]
		// and Collections have the key /Collection:[id], the Collection key is the prefix for looking up keys
		Prefix: c.Key().String(),
		Filters: []query.Filter{
			// Pass in a Filter Type to specify that results must be of type CollectionItem
			// In abstract terms this combined with the Prefix query param amounts to querying:
			// /Collection:[id]/CollectionItem:*
			sql_datastore.FilterKeyTypeEq(CollectionItem{}.DatastoreType()),
		},
		Orders: []query.Order{
			query.OrderByValue{
				TypedOrder: sql_datastore.OrderBy(orderby),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	i := 0
	for r := range res.Next() {
		if r.Error != nil {
			return nil, r.Error
		}

		c, ok := r.Value.(*CollectionItem)
		if !ok {
			return nil, ErrInvalidResponse
		}

		items[i] = c
		i++
	}

	// fmt.Println(items)
	return items[:i], nil
}
