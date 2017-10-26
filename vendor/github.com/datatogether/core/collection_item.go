package core

import (
	"database/sql"
	"fmt"
	"github.com/datatogether/sql_datastore"
	"github.com/datatogether/sqlutil"
	"github.com/ipfs/go-datastore"
)

// CollectionItem is an item in a collection. They are urls
// with added collection-specific information.
// This has the effect of storing all of the "main properties"
// of a collection item in the common list of urls
type CollectionItem struct {
	// Collection Items are Url's at heart
	Url
	// need a reference to the collection Id to be set to distinguish
	// this item's membership in this particular list
	collectionId string
	// this item's index in the collection
	Index int `json:"index"`
	// unique description of this item
	Description string `json:"description"`
}

// DatastoreType is to satisfy sql_datastore.Model interface
func (c CollectionItem) DatastoreType() string {
	return "CollectionItem"
}

// GetId returns the Id of the collectionItem, which is the id
// of the underlying Url
func (c CollectionItem) GetId() string {
	return c.Url.Id
}

// Key is somewhat special as CollectionItems always have a Collection
// as their parent. This relationship is represented in directory-form:
// /Collection:[collection-id]/CollectionItem:[item-id]
func (c CollectionItem) Key() datastore.Key {
	return datastore.NewKey(fmt.Sprintf("%s:%s/%s:%s", Collection{}.DatastoreType(), c.collectionId, c.DatastoreType(), c.GetId()))
}

// Read collection from db
func (c *CollectionItem) Read(store datastore.Datastore) error {
	if c.Url.Id == "" && c.Url.Url != "" {
		if sqls, ok := store.(*sql_datastore.Datastore); ok {
			row := sqls.DB.QueryRow(qUrlByUrlString, c.Url.Url)
			prev := &Url{}
			if err := prev.UnmarshalSQL(row); err == nil {
				c.Id = prev.Id
				// exists = true
			}
		}
	}

	ci, err := store.Get(c.Key())
	if err != nil {
		return err
	}

	got, ok := ci.(*CollectionItem)
	if !ok {
		return ErrInvalidResponse
	}
	*c = *got
	return nil
}

// Save a collection item to a store
func (c *CollectionItem) Save(store datastore.Datastore) (err error) {
	u := &c.Url
	if err := u.Save(store); err != nil {
		return err
	}

	c.Url = *u
	return store.Put(c.Key(), c)
}

// Delete a collection item
func (c *CollectionItem) Delete(store datastore.Datastore) error {
	return store.Delete(c.Key())
}

func (c *CollectionItem) NewSQLModel(key datastore.Key) sql_datastore.Model {
	l := key.List()
	if len(l) == 1 {
		return &CollectionItem{
			collectionId: datastore.NamespaceValue(l[0]),
		}
	} else if len(l) == 2 {
		return &CollectionItem{
			collectionId: datastore.NamespaceValue(l[0]),
			Url:          Url{Id: datastore.NamespaceValue(l[1])},
		}
	}
	return &CollectionItem{}
}

// SQLQuery is to satisfy the sql_datastore.Model interface, it
// returns the concrete query for a given type of SQL command
func (c CollectionItem) SQLQuery(cmd sql_datastore.Cmd) string {
	switch cmd {
	case sql_datastore.CmdCreateTable:
		return qCollectionItemCreateTable
	case sql_datastore.CmdExistsOne:
		return qCollectionItemExists
	case sql_datastore.CmdSelectOne:
		return qCollectionItemById
	case sql_datastore.CmdInsertOne:
		return qCollectionItemInsert
	case sql_datastore.CmdUpdateOne:
		return qCollectionItemUpdate
	case sql_datastore.CmdDeleteOne:
		return qCollectionItemDelete
	case sql_datastore.CmdList:
		return qCollectionItems
	default:
		return ""
	}
}

// SQLQuery is to satisfy the sql_datastore.Model interface, it
// returns this CollectionItem's parameters for a given type of SQL command
func (c *CollectionItem) SQLParams(cmd sql_datastore.Cmd) []interface{} {
	switch cmd {
	case sql_datastore.CmdSelectOne, sql_datastore.CmdExistsOne, sql_datastore.CmdDeleteOne:
		return []interface{}{c.collectionId, c.Url.Id}
	case sql_datastore.CmdList:
		return []interface{}{c.collectionId}
	default:
		return []interface{}{
			c.collectionId,
			c.Url.Id,
			c.Index,
			c.Description,
		}
	}
}

// UnmarshalSQL reads an sql response into the collection receiver
// it expects the request to have used collectionCols() for selection
func (c *CollectionItem) UnmarshalSQL(row sqlutil.Scannable) (err error) {
	var (
		collectionId, urlId, url, hash, title, description string
		index                                              int
	)

	if err := row.Scan(&collectionId, &urlId, &hash, &url, &title, &index, &description); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}

	*c = CollectionItem{
		collectionId: collectionId,
		Url:          Url{Id: urlId, Hash: hash, Url: url, Title: title},
		Index:        index,
		Description:  description,
	}

	return nil
}
