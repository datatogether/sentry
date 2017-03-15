package main

import (
	"database/sql"
	"fmt"
	"github.com/pborman/uuid"
	"time"
)

// Primer is tracking information about a base URL
type Primer struct {
	Id          string    `json:"id"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
}

// CrawlUrls returns the list of listed urls for crawling associated with this primer
func (p *Primer) CrawlUrls(db sqlQueryable) ([]*CrawlUrl, error) {
	rows, err := db.Query(fmt.Sprintf("select %s from crawl_urls where primer_id = $1", crawlUrlCols()), p.Id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	urls := make([]*CrawlUrl, 0)
	for rows.Next() {
		c := &CrawlUrl{}
		if err := c.UnmarshalSQL(rows); err != nil {
			return nil, err
		}
		urls = append(urls, c)
	}

	return urls, nil
}

func (p *Primer) Read(db sqlQueryable) error {
	if p.Id != "" {
		row := db.QueryRow(fmt.Sprintf("select %s from primers where id = $1", primerCols()), p.Id)
		return p.UnmarshalSQL(row)
	}
	return ErrNotFound
}

func (p *Primer) Save(db sqlQueryExecable) error {
	prev := &Primer{Id: p.Id}
	if err := prev.Read(db); err != nil {
		if err == ErrNotFound {
			p.Id = uuid.New()
			p.Created = time.Now().Round(time.Second)
			p.Updated = p.Created
			_, err := db.Exec(fmt.Sprintf("insert into primers (%s) values ($1, $2, $3, $4, $5)", primerCols()), p.SQLArgs()...)
			return err
		} else {
			return err
		}
	} else {
		p.Updated = time.Now().Round(time.Second)
		_, err := db.Exec("update primers set created=$2, updated = $3, title = $4, description = $5 where id = $1", p.SQLArgs()...)
		return err
	}
	return nil
}

func (p *Primer) Delete(db sqlQueryExecable) error {
	_, err := db.Exec("delete from primers where id = $1", p.Id)
	return err
}

func (p *Primer) UnmarshalSQL(row sqlScannable) error {
	var (
		id, title, description string
		created, updated       time.Time
	)

	if err := row.Scan(&id, &created, &updated, &title, &description); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}

	*p = Primer{
		Id:          id,
		Created:     created.In(time.UTC),
		Updated:     updated.In(time.UTC),
		Title:       title,
		Description: description,
	}

	return nil
}

func primerCols() string {
	return "id, created, updated, title, description"
}

func (p *Primer) SQLArgs() []interface{} {
	return []interface{}{
		p.Id,
		p.Created.In(time.UTC),
		p.Updated.In(time.UTC),
		p.Title,
		p.Description,
	}
}
