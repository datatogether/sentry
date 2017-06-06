package archive

import (
	"database/sql"
	"github.com/archivers-space/sqlutil"
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
	Title string
	// Human-readable description
	Description string
	// Main url link to the DataRepository
	Url string
}

// Read dataRepo from db
func (d *DataRepo) Read(db sqlutil.Queryable) error {
	if d.Id != "" {
		row := db.QueryRow(qDataRepoById, d.Id)
		return d.UnmarshalSQL(row)
	}
	return ErrNotFound
}

// Save a dataRepo
func (d *DataRepo) Save(db sqlutil.Execable) error {
	prev := &DataRepo{Id: d.Id}
	if err := prev.Read(db); err != nil {
		if err == ErrNotFound {
			d.Id = uuid.New()
			d.Created = time.Now().Round(time.Second)
			d.Updated = d.Created
			_, err := db.Exec(qDataRepoInsert, d.SQLArgs()...)
			return err
		} else {
			return err
		}
	} else {
		d.Updated = time.Now().Round(time.Second)
		_, err := db.Exec(qDataRepoUpdate, d.SQLArgs()...)
		return err
	}
	return nil
}

// Delete a dataRepo, should only do for erronious additions
func (d *DataRepo) Delete(db sqlutil.Execable) error {
	_, err := db.Exec(qDataRepoDelete, d.Id)
	return err
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

// SQLArgs formats a dataRepo struct for inserting / updating into postgres
func (d *DataRepo) SQLArgs() []interface{} {
	return []interface{}{
		d.Id,
		d.Created.In(time.UTC),
		d.Updated.In(time.UTC),
		d.Title,
		d.Description,
		d.Url,
	}
}
