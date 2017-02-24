package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type Context struct {
	Url           *url.Url
	Created       time.Time
	Updated       time.Time
	Hash          string
	ContributorId string
	Context       map[string]interface{}
}

func (c *Context) Read(db sqlQueryable) error {
	return c.UnmarshalSQL(db.QueryRow(fmt.Sprintf("select %s from context where url = $1 and contributor_id = $2", c.cols()), c.Url.String(), c.ContributorId))
}

func (c *Context) Save(db sqlQueryExecable) (err error) {
	prev := &Context{Url: c.Url, ContributorId: c.ContributorId}
	err = prev.Read(db)
	if err == ErrNotFound {
		// insert
		c.Created = time.Now()
		c.Updated = c.Created
		_, err = db.Exec("insert into context values ($1, $2, $3, $4, $5, $6)", c.SQLArgs()...)
	} else {
		c.Updated = time.Now()
		_, err = db.Exec("update context set created = $2, updated = $3, hash = $4, contributor_id = $5, context = $6 where url = $1", c.SQLArgs()...)
	}

	return err
}

func (c *Context) Delete(db sqlQueryExecable) error {
	_, err := db.Exec("delete from context where url = $1 and contributor_id = $2", c.Url.String(), c.ContributorId)
	return err
}

func (c Context) cols() string {
	return "url, created, updated, hash, contributor_id, context"
}

func (c *Context) UnmarshalSQL(row sqlScannable) error {
	var (
		rawurl, hash, contributorId string
		created, updated            int64
		contextBytes                []byte
	)

	if err := row.Scan(&rawurl, &created, &updated, &hash, &contributorId, &contextBytes); err != nil {
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

	*c = Context{
		Url:           parsedUrl,
		Created:       time.Unix(created, 0),
		Updated:       time.Unix(updated, 0),
		Hash:          hash,
		ContributorId: contributorId,
		Context:       context,
	}

	return nil
}

func (c *Context) SQLArgs() []interface{} {
	contextBytes, err := json.Marshal()
	if err != nil {
		panic(err)
	}

	return []interface{}{
		c.Url.String(),
		c.Created.Unix(),
		c.Updated.Unix(),
		c.Hash,
		c.ContributorId,
		contextBytes,
	}
}
