package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type CrawlUrl struct {
	Url           string                 `json:"url"`
	Created       time.Time              `json:"created"`
	Updated       time.Time              `json:"updated"`
	PrimerId      string                 `json:"primerId"`
	Crawl         bool                   `json:"crawl"`
	StaleDuration time.Duration          `json:"staleDuration"`
	LastAlertSent *time.Time             `json:"lastAlertSent"`
	Meta          map[string]interface{} `json:"meta"`
}

// AsUrl retrieves for the crawlUrl. If one doesn't exist & the url is saved,
// a new url is created
func (c *CrawlUrl) AsUrl(db sqlQueryExecable) (*Url, error) {
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

func (c *CrawlUrl) Read(db sqlQueryable) error {
	if c.Url != "" {
		row := db.QueryRow(fmt.Sprintf("select %s from crawl_urls where url = $1", crawlUrlCols()), c.Url)
		return c.UnmarshalSQL(row)
	}
	return ErrNotFound
}

func (c *CrawlUrl) Save(db sqlQueryExecable) error {
	prev := &CrawlUrl{Url: c.Url}
	if err := prev.Read(db); err != nil {
		if err == ErrNotFound {
			c.Created = time.Now().Round(time.Second)
			c.Updated = c.Created
			_, err := db.Exec(fmt.Sprintf("insert into crawl_urls (%s) values ($1, $2, $3, $4, $5, $6, $7, $8)", crawlUrlCols()), c.SQLArgs()...)
			return err
		} else {
			return err
		}
	} else {
		c.Updated = time.Now().Round(time.Second)
		_, err := db.Exec("update crawl_urls set created = $2, updated = $3, primer_id = $4, crawl = $5, stale_duration = $6, last_alert_sent = $7, meta = $8 where url = $1", c.SQLArgs()...)
		return err
	}

	return nil
}

func (c *CrawlUrl) Delete(db sqlQueryExecable) error {
	_, err := db.Exec("delete from crawl_urls where url = $1", c.Url)
	return err
}

func (c *CrawlUrl) UnmarshalSQL(row sqlScannable) error {
	var (
		url, pId         string
		created, updated time.Time
		lastAlert        *time.Time
		stale            int64
		crawl            bool
		metaBytes        []byte
	)

	if err := row.Scan(&url, &created, &updated, &pId, &crawl, &stale, &lastAlert, &metaBytes); err != nil {
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

	*c = CrawlUrl{
		Url:           url,
		Created:       created.In(time.UTC),
		Updated:       updated.In(time.UTC),
		PrimerId:      pId,
		Crawl:         crawl,
		StaleDuration: time.Duration(stale * 1000000),
		LastAlertSent: lastAlert,
		Meta:          meta,
	}

	return nil
}

func crawlUrlCols() string {
	return "url, created, updated, primer_id, crawl, stale_duration, last_alert_sent, meta"
}

func (c *CrawlUrl) SQLArgs() []interface{} {
	date := c.LastAlertSent
	if date != nil {
		utc := date.In(time.UTC)
		date = &utc
	}

	metaBytes, err := json.Marshal(c.Meta)
	if err != nil {
		panic(err)
	}

	return []interface{}{
		c.Url,
		c.Created.In(time.UTC),
		c.Updated.In(time.UTC),
		c.PrimerId,
		c.Crawl,
		c.StaleDuration / 1000000,
		date,
		metaBytes,
	}
}
