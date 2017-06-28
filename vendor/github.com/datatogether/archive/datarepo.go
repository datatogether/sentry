package archive

import (
	"database/sql"
	"fmt"
	"github.com/datatogether/sql_datastore"
	"github.com/datatogether/sqlutil"
	"github.com/ipfs/go-datastore"
	"github.com/pborman/uuid"
	"time"
)

// DataRepo is a place that holds data in a structured format
type DataRepo struct {
	// version 4 uuid
	Id string
	// Created timestamp rounded to seconds in UTC
	Created time.Time `json:"created"`
	// Updated timestamp rounded to seconds in UTC
	Updated time.Time `json:"updated"`
	// Title of this data repository
	Title string `json:"title"`
	// Human-readable description
	Description string `json:"description"`
	// Main url link to the DataRepository
	Url string `json:"url"`
}

func (d *DataRepo) DatastoreType() string {
	return "DataRepo"
}

func (d *DataRepo) GetId() string {
	return d.Id
}

func (d *DataRepo) Key() datastore.Key {
	return datastore.NewKey(fmt.Sprintf("%s:%s", d.DatastoreType(), d.GetId()))
}

// Read dataRepo from db
func (d *DataRepo) Read(store datastore.Datastore) error {
	di, err := store.Get(d.Key())
	if err != nil {
		return err
	}

	got, ok := di.(*DataRepo)
	if !ok {
		return ErrInvalidResponse
	}
	*d = *got
	return nil
}

// Save a dataRepo
func (d *DataRepo) Save(store datastore.Datastore) (err error) {
	var exists bool

	if d.Id != "" {
		exists, err = store.Has(d.Key())
		if err != nil {
			return err
		}
	}

	if !exists {
		d.Id = uuid.New()
		d.Created = time.Now().Round(time.Second)
		d.Updated = d.Created
	} else {
		d.Updated = time.Now().Round(time.Second)
	}

	return store.Put(d.Key(), d)
}

// Delete a dataRepo, should only do for erronious additions
func (d *DataRepo) Delete(store datastore.Datastore) error {
	return store.Delete(d.Key())
}

func (d *DataRepo) NewSQLModel(id string) sql_datastore.Model {
	return &DataRepo{
		Id: id,
	}
}

func (d DataRepo) SQLQuery(cmd sql_datastore.Cmd) string {
	switch cmd {
	case sql_datastore.CmdCreateTable:
		return qDataRepoCreateTable
	case sql_datastore.CmdExistsOne:
		return qDataRepoExists
	case sql_datastore.CmdSelectOne:
		return qDataRepoById
	case sql_datastore.CmdInsertOne:
		return qDataRepoInsert
	case sql_datastore.CmdUpdateOne:
		return qDataRepoUpdate
	case sql_datastore.CmdDeleteOne:
		return qDataRepoDelete
	case sql_datastore.CmdList:
		return qDataRepos
	default:
		return ""
	}
}

func (d DataRepo) SQLParams(cmd sql_datastore.Cmd) []interface{} {
	switch cmd {
	case sql_datastore.CmdSelectOne, sql_datastore.CmdExistsOne, sql_datastore.CmdDeleteOne:
		return []interface{}{d.Id}
	default:
		return []interface{}{
			d.Id,
			d.Created.In(time.UTC),
			d.Updated.In(time.UTC),
			d.Title,
			d.Description,
			d.Url,
		}
	}
}

// UnmarshalSQL reads an sql response into the dataRepo receiver
// it expects the request to have used dataRepoCols() for selection
func (d *DataRepo) UnmarshalSQL(row sqlutil.Scannable) (err error) {
	var (
		id, title, description, url string
		created, updated            time.Time
	)

	if err := row.Scan(&id, &created, &updated, &title, &description, &url); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}

	*d = DataRepo{
		Id:          id,
		Created:     created.In(time.UTC),
		Updated:     updated.In(time.UTC),
		Title:       title,
		Description: description,
		Url:         url,
	}

	return nil
}
