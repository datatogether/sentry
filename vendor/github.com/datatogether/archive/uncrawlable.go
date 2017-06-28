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

// Uncrawlables are urls that contain content that cannot be extracted
// with traditional web crawling / scraping methods. This model classifies
// the nature of the uncrawlable, setting the stage for writing custom scripts
// to extract the underlying content.
type Uncrawlable struct {
	// version 4 uuid
	Id string `json:"id"`
	// url from urls table, must be unique
	Url string `json:"url"`
	// Created timestamp rounded to seconds in UTC
	Created time.Time `json:"created"`
	// Updated timestamp rounded to seconds in UTC
	Updated time.Time `json:"updated"`
	// sha256 multihash of the public key that created this uncrawlable
	Creator string `json:"creator"`
	// name of person making submission
	Name string `json:"name"`
	// email address of person making submission
	Email string `json:"email"`
	// name of data rescue event where uncrawlable was added
	EventName string `json:"eventName"`
	// agency name
	Agency string `json:"agency"`
	// EDGI agency Id
	AgencyId string `json:"agencyId"`
	// EDGI subagency Id
	SubagencyId string `json:"subagencyId"`
	// EDGI organization Id
	OrgId string `json:"orgId"`
	// EDGI Suborganization Id
	SuborgId string `json:"orgId"`
	// EDGI subprimer Id
	SubprimerId string `json:"subprimerId"`
	// flag for ftp content
	Ftp bool `json:"ftp"`
	// flag for 'database'
	// TODO - refine this?
	Database bool `json:"database"`
	// flag for visualization / interactive content
	// obfuscating data
	Interactive bool `json:"interactive"`
	// flag for a page that links to many files
	ManyFiles bool `json:"manyFiles"`
	// uncrawlable comments
	Comments string `json:"comments"`
}

func (u Uncrawlable) DatastoreType() string {
	return "Uncrawlable"
}

func (u Uncrawlable) GetId() string {
	return u.Id
}

func (u Uncrawlable) Key() datastore.Key {
	return datastore.NewKey(fmt.Sprintf("%s:%s", u.DatastoreType(), u.GetId()))
}

// Read uncrawlable from db
func (u *Uncrawlable) Read(store datastore.Datastore) error {
	// return store.Delete(u.Key())
	if u.Id != "" {
		ui, err := store.Get(u.Key())
		if err != nil {
			return err
		}

		got, ok := ui.(*Uncrawlable)
		if !ok {
			return ErrInvalidResponse
		}
		*u = *got
		return nil
	} else {
		// TODO - figure out a way to query stores by url...
		if sqlstore, ok := store.(*sql_datastore.Datastore); ok {
			if u.Url != "" {
				row := sqlstore.DB.QueryRow(qUncrawlableByUrl, u.Url)
				return u.UnmarshalSQL(row)
			}
		}
	}

	return ErrNotFound
}

// Save a uncrawlable
func (u *Uncrawlable) Save(store datastore.Datastore) (err error) {
	var exists bool

	if u.Id != "" {
		exists, err = store.Has(u.Key())
		if err != nil {
			return err
		}
	}

	if !exists {
		u.Id = uuid.New()
		u.Created = time.Now().Round(time.Second)
		u.Updated = u.Created
	} else {
		u.Updated = time.Now().Round(time.Second)
	}

	return store.Put(u.Key(), u)
}

// Delete a uncrawlable, should only do for erronious additions
func (u *Uncrawlable) Delete(store datastore.Datastore) error {
	return store.Delete(u.Key())
}

func (u *Uncrawlable) NewSQLModel(id string) sql_datastore.Model {
	return &Uncrawlable{
		Id:  id,
		Url: u.Url,
	}
}

func (u *Uncrawlable) SQLQuery(cmd sql_datastore.Cmd) string {
	switch cmd {
	case sql_datastore.CmdCreateTable:
		return qUncrawlableCreateTable
	case sql_datastore.CmdExistsOne:
		if u.Id == "" {
			return qUncrawlableExistsByUrl
		} else {
			return qUncrawlableExists
		}
	case sql_datastore.CmdSelectOne:
		if u.Id == "" {
			return qUncrawlableByUrl
		} else {
			return qUncrawlableById
		}
	case sql_datastore.CmdInsertOne:
		return qUncrawlableInsert
	case sql_datastore.CmdUpdateOne:
		return qUncrawlableUpdate
	case sql_datastore.CmdDeleteOne:
		return qUncrawlableDelete
	case sql_datastore.CmdList:
		return qUncrawlablesList
	default:
		return ""
	}
}

// SQLParams formats a uncrawlable struct for inserting / updating into postgres
func (u *Uncrawlable) SQLParams(cmd sql_datastore.Cmd) []interface{} {
	switch cmd {
	case sql_datastore.CmdList:
		return []interface{}{}
	case sql_datastore.CmdSelectOne, sql_datastore.CmdExistsOne, sql_datastore.CmdDeleteOne:
		return []interface{}{u.Id}
	default:
		return []interface{}{
			u.Id,
			u.Url,
			u.Created.In(time.UTC),
			u.Updated.In(time.UTC),
			u.Creator,
			u.Name,
			u.Email,
			u.EventName,
			u.Agency,
			u.AgencyId,
			u.SubagencyId,
			u.OrgId,
			u.SuborgId,
			u.SubprimerId,
			u.Ftp,
			u.Database,
			u.Interactive,
			u.ManyFiles,
			u.Comments,
		}
	}
}

// UnmarshalSQL reads an sql response into the uncrawlable receiver
// it expects the request to have used uncrawlableCols() for selection
func (u *Uncrawlable) UnmarshalSQL(row sqlutil.Scannable) (err error) {
	var (
		created, updated                         time.Time
		id, creator, url, name, email, eventName string
		agency, agencyId, subagencyId            string
		orgId, suborgId, subprimerId, comments   string
		ftp, database, interactive, manyFiles    bool
	)

	if err := row.Scan(
		&id, &url, &created, &updated,
		&creator, &name, &email, &eventName, &agency,
		&agencyId, &subagencyId, &orgId, &suborgId, &subprimerId,
		&ftp, &database, &interactive, &manyFiles,
		&comments); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}

	*u = Uncrawlable{
		Id:          id,
		Created:     created.In(time.UTC),
		Updated:     updated.In(time.UTC),
		Creator:     creator,
		Url:         url,
		Name:        name,
		Email:       email,
		EventName:   eventName,
		Agency:      agency,
		AgencyId:    agencyId,
		SubagencyId: subagencyId,
		OrgId:       orgId,
		SuborgId:    suborgId,
		SubprimerId: subprimerId,
		Ftp:         ftp,
		Database:    database,
		Interactive: interactive,
		ManyFiles:   manyFiles,
		Comments:    comments,
	}

	return nil
}
