package archive

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/archivers-space/sqlutil"
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

func (s *Source) CalcStats(db sqlutil.Execable) error {
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
	return s.Save(db)
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
func (c *Source) AsUrl(db sqlutil.Execable) (*Url, error) {
	// TODO - this assumes http protocol, make moar robust
	addr, err := url.Parse(fmt.Sprintf("http://%s", c.Url))
	if err != nil {
		return nil, err
	}

	u := &Url{Url: addr.String()}
	if err := u.Read(db); err != nil {
		if err == ErrNotFound {
			if err := u.Insert(db); err != nil {
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

func (s *Source) Read(db sqlutil.Queryable) error {
	if s.Id != "" {
		row := db.QueryRow(qSourceById, s.Id)
		return s.UnmarshalSQL(row)
	} else if s.Url != "" {
		row := db.QueryRow(qSourceByUrl, s.Url)
		return s.UnmarshalSQL(row)
	}
	return ErrNotFound
}

func (c *Source) Save(db sqlutil.Execable) error {
	prev := &Source{Url: c.Url}
	if err := prev.Read(db); err != nil {
		if err == ErrNotFound {
			c.Id = uuid.New()
			c.Created = time.Now().Round(time.Second)
			c.Updated = c.Created
			_, err := db.Exec(qSourceInsert, c.SQLArgs()...)
			return err
		} else {
			return err
		}
	} else {
		c.Updated = time.Now().Round(time.Second)
		_, err := db.Exec(qSourceUpdate, c.SQLArgs()...)
		return err
	}

	return nil
}

func (c *Source) Delete(db sqlutil.Execable) error {
	_, err := db.Exec(qSourceDelete, c.Url)
	return err
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

func (c *Source) SQLArgs() []interface{} {
	date := c.LastAlertSent
	if date != nil {
		utc := date.In(time.UTC)
		date = &utc
	}

	metaBytes, err := json.Marshal(c.Meta)
	if err != nil {
		panic(err)
	}

	statBytes, err := json.Marshal(c.Stats)
	if err != nil {
		panic(err)
	}

	return []interface{}{
		c.Id,
		c.Created.In(time.UTC),
		c.Updated.In(time.UTC),
		c.Title,
		c.Description,
		c.Url,
		c.Primer.Id,
		c.Crawl,
		c.StaleDuration / 1000000,
		date,
		metaBytes,
		statBytes,
	}
}
