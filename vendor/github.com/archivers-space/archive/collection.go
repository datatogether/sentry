package archive

import (
	"database/sql"
	"encoding/json"
	"github.com/archivers-space/sqlutil"
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
	// url this collection originates from
	Url string `json:"url,omitempty"`
	// csv column headers, first value must always be "hash"
	Schema []string `json:"schema,omitempty"`
	// actuall collection contents
	Contents [][]string `json:"contents,omitempty"`
}

// Read collection from db
func (c *Collection) Read(db sqlutil.Queryable) error {
	if c.Id != "" {
		row := db.QueryRow(qCollectionById, c.Id)
		return c.UnmarshalSQL(row)
	}
	return ErrNotFound
}

// Save a collection
func (c *Collection) Save(db sqlutil.Execable) error {
	prev := &Collection{Id: c.Id}
	if err := prev.Read(db); err != nil {
		if err == ErrNotFound {
			c.Id = uuid.New()
			c.Created = time.Now().Round(time.Second)
			c.Updated = c.Created
			_, err := db.Exec(qCollectionInsert, c.SQLArgs()...)
			return err
		} else {
			return err
		}
	} else {
		c.Updated = time.Now().Round(time.Second)
		_, err := db.Exec(qCollectionUpdate, c.SQLArgs()...)
		return err
	}
	return nil
}

// Delete a collection, should only do for erronious additions
func (c *Collection) Delete(db sqlutil.Execable) error {
	_, err := db.Exec(qCollectionDelete, c.Id)
	return err
}

// UnmarshalSQL reads an sql response into the collection receiver
// it expects the request to have used collectionCols() for selection
func (c *Collection) UnmarshalSQL(row sqlutil.Scannable) (err error) {
	var (
		id, creator, title, url   string
		created, updated          time.Time
		schemaBytes, contentBytes []byte
	)

	if err := row.Scan(&id, &created, &updated, &creator, &title, &url, &schemaBytes, &contentBytes); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}

	var schema []string
	if schemaBytes != nil {
		schema = []string{}
		err = json.Unmarshal(schemaBytes, &schema)
		if err != nil {
			return err
		}
	}

	var contents [][]string
	if contentBytes != nil {
		contents = [][]string{}
		err = json.Unmarshal(contentBytes, &contents)
		if err != nil {
			return err
		}
	}

	*c = Collection{
		Id:       id,
		Created:  created.In(time.UTC),
		Updated:  updated.In(time.UTC),
		Creator:  creator,
		Title:    title,
		Url:      url,
		Schema:   schema,
		Contents: contents,
	}

	return nil
}

// SQLArgs formats a collection struct for inserting / updating into postgres
func (c *Collection) SQLArgs() []interface{} {
	schemaBytes, err := json.Marshal(c.Schema)
	if err != nil {
		panic(err)
	}
	contentBytes, err := json.Marshal(c.Contents)
	if err != nil {
		panic(err)
	}

	return []interface{}{
		c.Id,
		c.Created.In(time.UTC),
		c.Updated.In(time.UTC),
		c.Creator,
		c.Title,
		c.Url,
		schemaBytes,
		contentBytes,
	}
}
