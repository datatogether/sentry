package main

import (
	"database/sql"
	"fmt"
	"net/url"
	"time"
)

// Domain is tracking information about a base URL
type Domain struct {
	Host          string
	Created       time.Time
	Updated       time.Time
	Crawl         bool
	StaleDuration time.Duration
	LastAlertSent time.Time
}

// Url retrieves for the domain. If one doesn't exist & the url is saved,
// a new url is created
func (d *Domain) Url(db sqlQueryExecable) (*Url, error) {
	addr, err := url.Parse(fmt.Sprintf("http://%s", d.Host))
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

func (d *Domain) Read(db sqlQueryable) error {
	if d.Host != "" {
		row := db.QueryRow(fmt.Sprintf("select %s from domains where host = $1", domainCols()), d.Host)
		return d.UnmarshalSQL(row)
	}
	return ErrNotFound
}

func (d *Domain) Insert(db sqlQueryExecable) error {
	d.Created = time.Now()
	d.Updated = d.Created
	_, err := db.Exec(fmt.Sprintf("insert into domains (%s) values ($1, $2, $3, $4, $5, $6, $7)", domainCols()), d.SQLArgs()...)
	return err
}

func (d *Domain) Update(db sqlQueryExecable) error {
	d.Updated = time.Now()
	_, err := db.Exec("update domains set created=$2 updated = $3, host = $4, stale_duration = $5, crawl = $6, last_alert_sent = $7 where url = $1", d.SQLArgs()...)
	return err
}

func (d *Domain) Delete(db sqlQueryExecable) error {
	_, err := db.Exec("delete from domains where host = $1", d.Host)
	return err
}

func (d *Domain) UnmarshalSQL(row sqlScannable) error {
	var (
		host                               string
		created, updated, stale, lastAlert int64
		crawl                              bool
	)

	if err := row.Scan(&host, &created, &updated, &stale, &crawl, &lastAlert); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}

	*d = Domain{
		Host:          host,
		Created:       time.Unix(created, 0),
		Updated:       time.Unix(updated, 0),
		StaleDuration: time.Duration(stale * 1000000),
		Crawl:         crawl,
		LastAlertSent: time.Unix(lastAlert, 0),
	}

	return nil
}

func domainCols() string {
	return "host, created, updated, stale_duration, crawl, last_alert_sent"
}

func (d *Domain) SQLArgs() []interface{} {
	return []interface{}{
		d.Host,
		d.Created.Unix(),
		d.Updated.Unix(),
		d.StaleDuration / 1000000,
		d.Crawl,
		d.LastAlertSent.Unix(),
	}
}
