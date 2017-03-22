package archive

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type Collection struct {
	Id       string     `json:"id"`
	Created  time.Time  `json:"created"`
	Updated  time.Time  `json:"updated"`
	Creator  string     `json:"creator"`
	Title    string     `json:"title"`
	Schema   []string   `json:"schema"`
	Contents [][]string `json:"contents"`
}

// Read collection from db
func (c *Collection) Read(db sqlQueryable) error {
	if c.Id != "" {
		row := db.QueryRow(fmt.Sprintf("select %s from collections where id = $1", collectionCols()), c.Id)
		return c.UnmarshalSQL(row)
	}
	return ErrNotFound
}

// Save a collection
func (c *Collection) Save(db sqlQueryExecable) error {
	prev := &Collection{Id: c.Id}
	if err := prev.Read(db); err != nil {
		if err == ErrNotFound {
			c.Id = NewUuid()
			c.Created = time.Now().Round(time.Second)
			c.Updated = c.Created
			_, err := db.Exec(fmt.Sprintf("insert into collections (%s) values ($1, $2, $3, $4, $5, $6, $7)", collectionCols()), c.SQLArgs()...)
			return err
		} else {
			return err
		}
	} else {
		c.Updated = time.Now().Round(time.Second)
		_, err := db.Exec("update collections set created=$2, updated=$3, creator=$4, title=$5, schema=$6, contents=$7 where id = $1", c.SQLArgs()...)
		return err
	}
	return nil
}

// Delete a collection, should only do for erronious additions
func (c *Collection) Delete(db sqlQueryExecable) error {
	_, err := db.Exec("delete from collections where id = $1", c.Id)
	return err
}

// standard-form columns for selection from postgres
func collectionCols() string {
	return "id, created, updated, creator, title, schema, contents"
}

// UnmarshalSQL reads an sql response into the collection receiver
// it expects the request to have used collectionCols() for selection
func (c *Collection) UnmarshalSQL(row sqlScannable) (err error) {
	var (
		id, creator, title        string
		created, updated          time.Time
		schemaBytes, contentBytes []byte
	)

	if err := row.Scan(&id, &created, &updated, &creator, &title, &schemaBytes, &contentBytes); err != nil {
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
		c.Title,
		c.Creator,
		schemaBytes,
		contentBytes,
	}
}
