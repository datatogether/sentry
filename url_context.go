package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type UrlContext struct {
	Url           *url.URL
	Created       time.Time
	Updated       time.Time
	Hash          string
	ContributorId string
	Context       map[string]interface{}
}

func (c *UrlContext) valid() error {
	if c.Url.String() == "" {
		return fmt.Errorf("Url is required")
	}
	if c.ContributorId == "" {
		return fmt.Errorf("ContributorId is required")
	}

	return nil
}

func (c *UrlContext) Read(db sqlQueryable) error {
	if err := c.valid(); err != nil {
		return err
	}
	return c.UnmarshalSQL(db.QueryRow(fmt.Sprintf("select %s from context where url = $1 and contributor_id = $2", c.cols()), c.Url.String(), c.ContributorId))
}

func (c *UrlContext) Save(db sqlQueryExecable) (err error) {
	if err := c.valid(); err != nil {
		return err
	}

	prev := &UrlContext{Url: c.Url, ContributorId: c.ContributorId}
	err = prev.Read(db)
	if err == ErrNotFound {
		// insert
		c.Created = time.Now()
		c.Updated = c.Created
		_, err = db.Exec("insert into context values ($1, $2, $3, $4, $5, $6)", c.SQLArgs()...)
	} else if err == nil {
		c.Updated = time.Now()
		_, err = db.Exec("update context set created = $3, updated = $4, hash = $5, meta = $6 where url = $1 and contributor_id = $2", c.SQLArgs()...)
	} else {
		return err
	}

	return err
}

func (c *UrlContext) Delete(db sqlQueryExecable) error {
	if err := c.valid(); err != nil {
		return err
	}

	_, err := db.Exec("delete from context where url = $1 and contributor_id = $2", c.Url.String(), c.ContributorId)
	return err
}

func (c UrlContext) cols() string {
	return "url, contributor_id, created, updated, hash, meta"
}

func (c *UrlContext) UnmarshalSQL(row sqlScannable) error {
	var (
		rawurl, hash, contributorId string
		created, updated            int64
		contextBytes                []byte
	)

	if err := row.Scan(&rawurl, &contributorId, &created, &updated, &hash, &contextBytes); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		logger.Println(err)
		return err
	}

	parsedUrl, err := url.Parse(rawurl)
	if err != nil {
		return err
	}

	var context map[string]interface{}
	if contextBytes != nil {
		context = map[string]interface{}{}
		err = json.Unmarshal(contextBytes, &context)
		if err != nil {
			return err
		}
	}

	*c = UrlContext{
		Url:           parsedUrl,
		ContributorId: contributorId,
		Created:       time.Unix(created, 0),
		Updated:       time.Unix(updated, 0),
		Hash:          hash,
		Context:       context,
	}

	return nil
}

func (c *UrlContext) SQLArgs() []interface{} {
	contextBytes, err := json.Marshal(c.Context)
	if err != nil {
		panic(err)
	}

	return []interface{}{
		c.Url.String(),
		c.ContributorId,
		c.Created.Unix(),
		c.Updated.Unix(),
		c.Hash,
		contextBytes,
	}
}
