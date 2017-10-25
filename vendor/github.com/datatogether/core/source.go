package core

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/datatogether/sql_datastore"
	"github.com/datatogether/sqlutil"
	"github.com/ipfs/go-datastore"
	"github.com/pborman/uuid"
	"net/url"
	"strings"
	"time"
)

// Source is a concreate handle for archiving. Crawlers use
// source's url as a base of a link tree. Sources are connected
// to a parent Primer to provide context & organization.
type Source struct {
	// version 4 uuid
	Id string `json:"id"`
	// Created timestamp rounded to seconds in UTC
	Created time.Time `json:"created"`
	// Updated timestamp rounded to seconds in UTC
	Updated time.Time `json:"updated"`
	// human-readable title for this source
	Title string `json:"title"`
	// description of the source, ideally one paragraph
	Description string `json:"description"`
	// absolute url to serve as the root of the
	Url string `json:"url"`
	// primer this source is connected to
	Primer *Primer `json:"primer"`
	// weather or not this url should be crawled be a web crawler
	Crawl bool `json:"crawl"`
	// amount of time before a link within this tree is considered in need
	// of re-checking for changes. currently not in use, but planned.
	StaleDuration time.Duration `json:"staleDuration"`
	// yeah this'll probably get depricated. Part of a half-baked alerts feature idea.
	LastAlertSent *time.Time `json:"lastAlertSent"`
	// Metadata associated with this source that should be added to all
	// child urls, currently not in use, but planned
	Meta map[string]interface{} `json:"meta"`
	// Stats about this source
	Stats *SourceStats `json:"stats"`
}

type SourceStats struct {
	UrlCount             int `json:"urlCount"`
	ArchivedUrlCount     int `json:"archivedUrlCount"`
	ContentUrlCount      int `json:"contentUrlCount"`
	ContentMetadataCount int `json:"contentMetadataCount"`
}

func (s Source) DatastoreType() string {
	return "Source"
}

func (s Source) GetId() string {
	return s.Id
}

func (s Source) Key() datastore.Key {
	return datastore.NewKey(fmt.Sprintf("%s:%s", s.DatastoreType(), s.GetId()))
}

func (s *Source) CalcStats(db *sql.DB) error {
	urlCount, err := s.urlCount(db)
	if err != nil {
		return err
	}

	contentUrlCount, err := s.contentUrlCount(db)
	if err != nil {
		return err
	}

	metadataCount, err := s.contentWithMetadataCount(db)
	if err != nil {
		return err
	}

	s.Stats = &SourceStats{
		UrlCount:             urlCount,
		ContentUrlCount:      contentUrlCount,
		ContentMetadataCount: metadataCount,
	}

	// TODO - stop saving here & instead hook this up to some sort of cron task
	store := sql_datastore.NewDatastore(db)
	if err := store.Register(&Source{}); err != nil {
		return err
	}
	return s.Save(store)
}

func (s *Source) urlCount(db sqlutil.Queryable) (count int, err error) {
	err = db.QueryRow(qSourceUrlCount, "%"+s.Url+"%").Scan(&count)
	return
}

func (s *Source) contentUrlCount(db sqlutil.Queryable) (count int, err error) {
	err = db.QueryRow(qSourceContentUrlCount, "%"+s.Url+"%").Scan(&count)
	return
}

func (s *Source) contentWithMetadataCount(db sqlutil.Queryable) (count int, err error) {
	err = db.QueryRow(qSourceContentWithMetadataCount, "%"+s.Url+"%").Scan(&count)
	return
}

// MatchesUrl checks to see if the url pattern of Source is contained
// within the passed-in url string
// TODO - make this more sophisticated, checking against the beginning of the
// url to avoid things like accidental matches, or urls in query params matching
// within rawurl
func (s *Source) MatchesUrl(rawurl string) bool {
	return strings.Contains(rawurl, s.Url)
}

// AsUrl retrieves the url that corresponds for the crawlUrl. If one doesn't exist & the url is saved,
// a new url is created
func (c *Source) AsUrl(db *sql.DB) (*Url, error) {
	// TODO - this assumes http protocol, make moar robust
	addr, err := url.Parse(fmt.Sprintf("http://%s", c.Url))
	if err != nil {
		return nil, err
	}

	store := sql_datastore.NewDatastore(db)
	if err := store.Register(&Url{}); err != nil {
		return nil, err
	}

	u := &Url{Url: addr.String()}
	if err := u.Read(store); err != nil {
		if err == ErrNotFound {
			if err := u.Save(store); err != nil {
				return u, err
			}
		} else {
			return nil, err
		}
	}

	return u, nil
}

// TODO - this currently doesn't check the status of metadata, gonna need to do that
// UndescribedContent returns a list of content-urls from this subprimer that need work.
func (s *Source) UndescribedContent(db sqlutil.Queryable, limit, offset int) ([]*Url, error) {
	rows, err := db.Query(qSourceUndescribedContentUrls, "%"+s.Url+"%", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls := make([]*Url, limit)
	i := 0
	for rows.Next() {
		u := &Url{}
		if err := u.UnmarshalSQL(rows); err != nil {
			return nil, err
		}
		urls[i] = u
		i++
	}

	return urls[:i], nil
}

// TODO - this currently doesn't check the status of metadata, gonna need to do that
// DescribedContent returns a list of content-urls from this subprimer that need work.
func (s *Source) DescribedContent(db sqlutil.Queryable, limit, offset int) ([]*Url, error) {
	rows, err := db.Query(qSourceDescribedContentUrls, "%"+s.Url+"%", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls := make([]*Url, limit)
	i := 0
	for rows.Next() {
		u := &Url{}
		if err := u.UnmarshalSQL(rows); err != nil {
			return nil, err
		}
		urls[i] = u
		i++
	}

	return urls[:i], nil
}

// func (s *Source) Stats() {
// }

func (s *Source) Read(store datastore.Datastore) error {
	if s.Id != "" {
		ci, err := store.Get(s.Key())
		if err != nil {
			return err
		}

		got, ok := ci.(*Source)
		if !ok {
			return ErrInvalidResponse
		}
		*s = *got
		return nil
	} else if s.Url != "" {
		// TODO - figure out a way to query stores by url...
		if sqlstore, ok := store.(*sql_datastore.Datastore); ok {
			row := sqlstore.DB.QueryRow(qSourceByUrl, s.Url)
			return s.UnmarshalSQL(row)
		}
	}
	return ErrNotFound
}

func (s *Source) Save(store datastore.Datastore) (err error) {
	var exists bool

	if s.Id != "" {
		exists, err = store.Has(s.Key())
		if err != nil {
			return err
		}
	}

	if !exists {
		s.Id = uuid.New()
		s.Created = time.Now().Round(time.Second)
		s.Updated = s.Created
	} else {
		s.Updated = time.Now().Round(time.Second)
	}

	return store.Put(s.Key(), s)
}

func (s *Source) Delete(store datastore.Datastore) error {
	return store.Delete(s.Key())
}

func (s *Source) NewSQLModel(key datastore.Key) sql_datastore.Model {
	return &Source{
		Id:  key.Name(),
		Url: s.Url,
	}
}

func (s *Source) SQLQuery(cmd sql_datastore.Cmd) string {
	switch cmd {
	case sql_datastore.CmdCreateTable:
		return qSourceCreateTable
	case sql_datastore.CmdExistsOne:
		if s.Id != "" {
			return qSourceExists
		} else {
			return qSourceExistsByUrl
		}
	case sql_datastore.CmdSelectOne:
		if s.Id != "" {
			return qSourceById
		} else {
			return qSourceByUrl
		}
	case sql_datastore.CmdInsertOne:
		return qSourceInsert
	case sql_datastore.CmdUpdateOne:
		return qSourceUpdate
	case sql_datastore.CmdDeleteOne:
		return qSourceDelete
	case sql_datastore.CmdList:
		return qSourcesList
	default:
		return ""
	}
}

func (s *Source) SQLParams(cmd sql_datastore.Cmd) []interface{} {
	switch cmd {
	case sql_datastore.CmdList:
		return []interface{}{}
	case sql_datastore.CmdSelectOne, sql_datastore.CmdExistsOne:
		if s.Id != "" {
			return []interface{}{s.Id}
		} else {
			return []interface{}{s.Url}
		}
	case sql_datastore.CmdDeleteOne:
		return []interface{}{s.Id}
	default:
		date := s.LastAlertSent
		if date != nil {
			utc := date.In(time.UTC)
			date = &utc
		}

		metaBytes, err := json.Marshal(s.Meta)
		if err != nil {
			panic(err)
		}

		statBytes, err := json.Marshal(s.Stats)
		if err != nil {
			panic(err)
		}

		if s.Primer == nil {
			s.Primer = &Primer{}
		}

		return []interface{}{
			s.Id,
			s.Created.In(time.UTC),
			s.Updated.In(time.UTC),
			s.Title,
			s.Description,
			s.Url,
			s.Primer.Id,
			s.Crawl,
			s.StaleDuration / 1000000,
			date,
			metaBytes,
			statBytes,
		}
	}
}

func (c *Source) UnmarshalSQL(row sqlutil.Scannable) error {
	var (
		id, url, pId, title, description string
		created, updated                 time.Time
		lastAlert                        *time.Time
		stale                            int64
		crawl                            bool
		metaBytes, statsBytes            []byte
	)

	if err := row.Scan(&id, &created, &updated, &title, &description, &url, &pId, &crawl, &stale, &lastAlert, &metaBytes, &statsBytes); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}

	if lastAlert != nil {
		utc := lastAlert.In(time.UTC)
		lastAlert = &utc
	}

	var meta map[string]interface{}
	if metaBytes != nil {
		if err := json.Unmarshal(metaBytes, &meta); err != nil {
			return err
		}
	}

	stats := &SourceStats{}
	if statsBytes != nil {
		if err := json.Unmarshal(statsBytes, stats); err != nil {
			return err
		}
	}

	*c = Source{
		Id:            id,
		Created:       created.In(time.UTC),
		Updated:       updated.In(time.UTC),
		Title:         title,
		Description:   description,
		Url:           url,
		Primer:        &Primer{Id: pId},
		Crawl:         crawl,
		StaleDuration: time.Duration(stale * 1000000),
		LastAlertSent: lastAlert,
		Meta:          meta,
		Stats:         stats,
	}

	return nil
}
