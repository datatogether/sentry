package core

import (
	"database/sql"
	"fmt"
	"github.com/datatogether/sql_datastore"
	"github.com/datatogether/sqlutil"
	datastore "github.com/ipfs/go-datastore"
	"github.com/pborman/uuid"
	"time"
)

// Collections are generic groupings of content
// collections can be thought of as a csv file listing content hashes
// as the first column, and whatever other information is necessary in
// subsequent columns
type Collection struct {
	// version 4 uuid
	Id string `json:"id"`
	// Created timestamp rounded to seconds in UTC
	Created time.Time `json:"created"`
	// Updated timestamp rounded to seconds in UTC
	Updated time.Time `json:"updated"`
	// sha256 multihash of the public key that created this collection
	Creator string `json:"creator"`
	// human-readable title of the collection
	Title string `json:"title"`
	// description of the collection
	Description string `json:"description"`
	// url this collection originates from
	Url string `json:"url,omitempty"`
}

func (c Collection) DatastoreType() string {
	return "Collection"
}

func (c Collection) GetId() string {
	return c.Id
}

func (c Collection) Key() datastore.Key {
	return datastore.NewKey(fmt.Sprintf("%s:%s", c.DatastoreType(), c.GetId()))
}

// Read collection from db
func (c *Collection) Read(store datastore.Datastore) error {
	if c.Id == "" && c.Url != "" {
		// TODO - figure out a way to query stores by url...
		if sqlstore, ok := store.(*sql_datastore.Datastore); ok {
			row := sqlstore.DB.QueryRow(qCollectionByUrl, c.Url)
			return c.UnmarshalSQL(row)
		}
	}

	ci, err := store.Get(c.Key())
	if err != nil {
		return err
	}

	got, ok := ci.(*Collection)
	if !ok {
		return ErrInvalidResponse
	}
	*c = *got
	return nil
}

// Save a collection
func (c *Collection) Save(store datastore.Datastore) (err error) {
	var exists bool

	if c.Id != "" {
		exists, err = store.Has(c.Key())
		if err != nil {
			return err
		}
	}

	if !exists {
		c.Id = uuid.New()
		c.Created = time.Now().Round(time.Second)
		c.Updated = c.Created
	} else {
		c.Updated = time.Now().Round(time.Second)
	}

	return store.Put(c.Key(), c)
}

// Delete a collection, should only do for erronious additions
func (c *Collection) Delete(store datastore.Datastore) error {
	return store.Delete(c.Key())
}

func (c *Collection) NewSQLModel(key datastore.Key) sql_datastore.Model {
	return &Collection{
		Id: key.Name(),
	}
}

func (c Collection) SQLQuery(cmd sql_datastore.Cmd) string {
	switch cmd {
	case sql_datastore.CmdCreateTable:
		return qCollectionCreateTable
	case sql_datastore.CmdExistsOne:
		return qCollectionExists
	case sql_datastore.CmdSelectOne:
		return qCollectionById
	case sql_datastore.CmdInsertOne:
		return qCollectionInsert
	case sql_datastore.CmdUpdateOne:
		return qCollectionUpdate
	case sql_datastore.CmdDeleteOne:
		return qCollectionDelete
	case sql_datastore.CmdList:
		return qCollections
	default:
		return ""
	}
}

func (c *Collection) SQLParams(cmd sql_datastore.Cmd) []interface{} {
	switch cmd {
	case sql_datastore.CmdSelectOne, sql_datastore.CmdExistsOne, sql_datastore.CmdDeleteOne:
		return []interface{}{c.Id}
	case sql_datastore.CmdList:
		return nil
	default:
		return []interface{}{
			c.Id,
			c.Created.In(time.UTC),
			c.Updated.In(time.UTC),
			c.Creator,
			c.Title,
			c.Description,
			c.Url,
		}
	}
}

// UnmarshalSQL reads an sql response into the collection receiver
// it expects the request to have used collectionCols() for selection
func (c *Collection) UnmarshalSQL(row sqlutil.Scannable) (err error) {
	var (
		id, creator, title, description, url string
		created, updated                     time.Time
	)

	if err := row.Scan(&id, &created, &updated, &creator, &title, &description, &url); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}

	*c = Collection{
		Id:          id,
		Created:     created.In(time.UTC),
		Updated:     updated.In(time.UTC),
		Creator:     creator,
		Title:       title,
		Description: description,
		Url:         url,
	}

	return nil
}
