package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/pborman/uuid"
	"net/url"
	"time"
)

type Subprimer struct {
	Id            string                 `json:"id"`
	Url           string                 `json:"url"`
	Created       time.Time              `json:"created"`
	Updated       time.Time              `json:"updated"`
	PrimerId      string                 `json:"primerId"`
	Crawl         bool                   `json:"crawl"`
	StaleDuration time.Duration          `json:"staleDuration"`
	LastAlertSent *time.Time             `json:"lastAlertSent"`
	Meta          map[string]interface{} `json:"meta"`
}

// AsUrl retrieves the url that corresponds for the crawlUrl. If one doesn't exist & the url is saved,
// a new url is created
func (c *Subprimer) AsUrl(db sqlQueryExecable) (*Url, error) {
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

func (c *Subprimer) Read(db sqlQueryable) error {
	if c.Id != "" {
		row := db.QueryRow(fmt.Sprintf("select %s from subprimers where id = $1", subprimerCols()), c.Id)
		return c.UnmarshalSQL(row)
	} else if c.Url != "" {
		row := db.QueryRow(fmt.Sprintf("select %s from subprimers where url = $1", subprimerCols()), c.Url)
		return c.UnmarshalSQL(row)
	}
	return ErrNotFound
}

func (c *Subprimer) Save(db sqlQueryExecable) error {
	prev := &Subprimer{Url: c.Url}
	if err := prev.Read(db); err != nil {
		if err == ErrNotFound {
			c.Id = uuid.New()
			c.Created = time.Now().Round(time.Second)
			c.Updated = c.Created
			_, err := db.Exec(fmt.Sprintf("insert into subprimers (%s) values ($1, $2, $3, $4, $5, $6, $7, $8, $9)", subprimerCols()), c.SQLArgs()...)
			return err
		} else {
			return err
		}
	} else {
		c.Updated = time.Now().Round(time.Second)
		_, err := db.Exec("update subprimers set url = $2, created = $3, updated = $4, primer_id = $5, crawl = $6, stale_duration = $7, last_alert_sent = $8, meta = $9 where id = $1", c.SQLArgs()...)
		return err
	}

	return nil
}

func (c *Subprimer) Delete(db sqlQueryExecable) error {
	_, err := db.Exec("delete from subprimers where url = $1", c.Url)
	return err
}

func (c *Subprimer) UnmarshalSQL(row sqlScannable) error {
	var (
		id, url, pId     string
		created, updated time.Time
		lastAlert        *time.Time
		stale            int64
		crawl            bool
		metaBytes        []byte
	)

	if err := row.Scan(&id, &url, &created, &updated, &pId, &crawl, &stale, &lastAlert, &metaBytes); err != nil {
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

	*c = Subprimer{
		Id:            id,
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

func subprimerCols() string {
	return "id, url, created, updated, primer_id, crawl, stale_duration, last_alert_sent, meta"
}

func (c *Subprimer) SQLArgs() []interface{} {
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
		c.Id,
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
