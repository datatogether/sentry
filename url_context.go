package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

// UrlContext associates contextualizing metadata with a url
// each "context" is associated with a userId to track their contribution
type UrlContext struct {
	Url           *url.URL
	Created       time.Time
	Updated       time.Time
	Hash          string
	ContributorId string
	Metadata      map[string]interface{}
}

// Read a context from the DB based on url & contributor ID
func (c *UrlContext) Read(db sqlQueryable) error {
	if err := c.valid(); err != nil {
		return err
	}
	return c.UnmarshalSQL(db.QueryRow(fmt.Sprintf("select %s from context where url = $1 and contributor_id = $2", urlContextCols()), c.Url.String(), c.ContributorId))
}

// Save inserts if no document exists, updates otherwise
func (c *UrlContext) Save(db sqlQueryExecable) (err error) {
	if err := c.valid(); err != nil {
		return err
	}

	prev := &UrlContext{Url: c.Url, ContributorId: c.ContributorId}
	err = prev.Read(db)
	if err == ErrNotFound {
		if err := c.ReadCurrentHash(db); err != nil {
			return err
		}

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

// Delete a context
func (c *UrlContext) Delete(db sqlQueryExecable) error {
	if err := c.valid(); err != nil {
		return err
	}

	_, err := db.Exec("delete from context where url = $1 and contributor_id = $2", c.Url.String(), c.ContributorId)
	return err
}

// valid checks for general valid-stateness, returns nil when valid
func (c *UrlContext) valid() error {
	if c.Url.String() == "" {
		return fmt.Errorf("Url is required")
	}
	if c.ContributorId == "" {
		return fmt.Errorf("ContributorId is required")
	}

	return nil
}

func (c *UrlContext) ReadCurrentHash(db sqlQueryable) error {
	var hash string

	err := db.QueryRow("select hash from urls where url = $1", c.Url.String()).Scan(&hash)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}

	c.Hash = hash

	return nil
}

// TODO - Storing urls as actual urls is a pain
func (c *UrlContext) UnmarshalJSON(data []byte) error {
	d := struct {
		Url           string
		Created       time.Time
		Updated       time.Time
		Hash          string
		ContributorId string
		Metadata      map[string]interface{}
	}{}
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}
	u, err := url.Parse(d.Url)
	if err != nil {
		return err
	}

	*c = UrlContext{
		Url:           u,
		Created:       d.Created,
		Updated:       d.Updated,
		Hash:          d.Hash,
		ContributorId: d.ContributorId,
		Metadata:      d.Metadata,
	}
	return nil
}

// UnmarshalSQL scans into the context
func (c *UrlContext) UnmarshalSQL(row sqlScannable) error {
	var (
		rawurl, hash, contributorId string
		created, updated            int64
		metadataBytes               []byte
	)

	if err := row.Scan(&rawurl, &contributorId, &created, &updated, &hash, &metadataBytes); err != nil {
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

	var metadata map[string]interface{}
	if metadataBytes != nil {
		metadata = map[string]interface{}{}
		err = json.Unmarshal(metadataBytes, &metadata)
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
		Metadata:      metadata,
	}

	return nil
}

// urlContextCols gives the expected order & selection of columns from the db
func urlContextCols() string {
	return "url, contributor_id, created, updated, hash, meta"
}

// SQLArgs gives context values as sql arguments
func (c *UrlContext) SQLArgs() []interface{} {
	metadataBytes, err := json.Marshal(c.Metadata)
	if err != nil {
		panic(err)
	}

	return []interface{}{
		c.Url.String(),
		c.ContributorId,
		c.Created.Unix(),
		c.Updated.Unix(),
		c.Hash,
		metadataBytes,
	}
}
