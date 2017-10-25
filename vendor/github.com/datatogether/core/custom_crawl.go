package core

import (
	"database/sql"
	"fmt"
	"github.com/datatogether/sql_datastore"
	"github.com/datatogether/sqlutil"
	"github.com/ipfs/go-datastore"
	"github.com/pborman/uuid"
	"time"
)

// CustomCrawls are urls that contain content that cannot be extracted
// with traditional web crawling / scraping methods. This model classifies
// the nature of the custom crawl, setting the stage for writing custom scripts
// to extract the underlying content.
type CustomCrawl struct {
	// version 4 uuid
	Id string `json:"id"`
	// Created timestamp rounded to seconds in UTC
	Created time.Time `json:"created"`
	// Updated timestamp rounded to seconds in UTC
	Updated time.Time `json:"updated"`
	// Json Web token that created this request
	Jwt string `json:"jwt"`
	// MorphRunId
	MorphRunId string `json:"morphRunId"`
	// timestamp this run was completed
	DateCompleted time.Time
	// repository for code that ran the crawl
	GithubRepo string `json:"githubRepo"`
	// OriginalUrl
	OriginalUrl string `json:"originalUrl"`
	// SqliteChecksum
	SqliteChecksum string `json:"sqliteChecksum"`
}

func (CustomCrawl) DatastoreType() string {
	return "CustomCrawl"
}

func (c CustomCrawl) GetId() string {
	return c.Id
}

func (u CustomCrawl) Key() datastore.Key {
	return datastore.NewKey(fmt.Sprintf("%s:%s", u.DatastoreType(), u.GetId()))
}

// Read custom crawl from db
func (c *CustomCrawl) Read(store datastore.Datastore) error {

	if c.Id != "" {
		ci, err := store.Get(c.Key())
		if err != nil {
			return err
		}

		got, ok := ci.(*CustomCrawl)
		if !ok {
			return ErrInvalidResponse
		}
		*c = *got
		return nil
	}

	return ErrNotFound
}

// Save a custom crawl
func (c *CustomCrawl) Save(store datastore.Datastore) (err error) {
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

// Delete a custom crawl, should only do for erronious additions
func (c *CustomCrawl) Delete(store datastore.Datastore) error {
	return store.Delete(c.Key())
}

func (c *CustomCrawl) NewSQLModel(key datastore.Key) sql_datastore.Model {
	return &CustomCrawl{
		Id: key.Name(),
	}
}

func (c *CustomCrawl) SQLQuery(cmd sql_datastore.Cmd) string {
	switch cmd {
	case sql_datastore.CmdCreateTable:
		return qCustomCrawlCreateTable
	case sql_datastore.CmdExistsOne:
		return qCustomCrawlExists
	case sql_datastore.CmdSelectOne:
		return qCustomCrawlById
	case sql_datastore.CmdInsertOne:
		return qCustomCrawlInsert
	case sql_datastore.CmdUpdateOne:
		return qCustomCrawlUpdate
	case sql_datastore.CmdDeleteOne:
		return qCustomCrawlDelete
	case sql_datastore.CmdList:
		return qCustomCrawlsList
	default:
		return ""
	}
}

// SQLParams formats a custom crawl struct for inserting / updating into postgres
func (c *CustomCrawl) SQLParams(cmd sql_datastore.Cmd) []interface{} {
	switch cmd {
	case sql_datastore.CmdList:
		return []interface{}{}
	case sql_datastore.CmdSelectOne, sql_datastore.CmdExistsOne, sql_datastore.CmdDeleteOne:
		return []interface{}{c.Id}
	default:
		return []interface{}{
			c.Id,
			c.Created.In(time.UTC),
			c.Updated.In(time.UTC),
			c.Jwt,
			c.MorphRunId,
			c.DateCompleted,
			c.GithubRepo,
			c.OriginalUrl,
			c.SqliteChecksum,
		}
	}
}

// UnmarshalSQL reads an sql response into the custom crawl receiver
// it expects the request to have used custom crawlCols() for selection
func (c *CustomCrawl) UnmarshalSQL(row sqlutil.Scannable) (err error) {
	var (
		created, updated, dateCompleted                              time.Time
		id, jwt, morphRunId, githubRepo, originalUrl, sqliteChecksum string
	)

	if err := row.Scan(
		&id, &created, &updated,
		&jwt, &morphRunId, &dateCompleted,
		&githubRepo, &originalUrl,
		&sqliteChecksum); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}

	*c = CustomCrawl{
		Id:             id,
		Created:        created.In(time.UTC),
		Updated:        updated.In(time.UTC),
		Jwt:            jwt,
		MorphRunId:     morphRunId,
		DateCompleted:  dateCompleted,
		GithubRepo:     githubRepo,
		OriginalUrl:    originalUrl,
		SqliteChecksum: sqliteChecksum,
	}

	return nil
}
