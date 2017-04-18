package archive

import (
	"database/sql"
	"time"
)

// Uncrawlables are urls that contain content that cannot be extracted
// with traditional web crawling / scraping methods. This model classifies
// the nature of the uncrawlable, setting the stage for writing custom scripts
// to extract the underlying content.
type Uncrawlable struct {
	// version 4 uuid
	Id string `json:"id"`
	// url from urls table, must be unique
	Url string `json:"url"`
	// Created timestamp rounded to seconds in UTC
	Created time.Time `json:"created"`
	// Updated timestamp rounded to seconds in UTC
	Updated time.Time `json:"updated"`
	// sha256 multihash of the public key that created this uncrawlable
	Creator string `json:"creator"`
	// name of person making submission
	Name string `json:"name"`
	// email address of person making submission
	Email string `json:"email"`
	// name of data rescue event where uncrawlable was added
	EventName string `json:"eventName"`
	// agency name
	Agency string `json:"agency"`
	// EDGI agency Id
	AgencyId string `json:"agencyId"`
	// EDGI subagency Id
	SubagencyId string `json:"subagencyId"`
	// EDGI organization Id
	OrgId string `json:"orgId"`
	// EDGI Suborganization Id
	SuborgId string `json:"orgId"`
	// EDGI subprimer Id
	SubprimerId string `json:"subprimerId"`
	// flag for ftp content
	Ftp bool `json:"ftp"`
	// flag for 'database'
	// TODO - refine this?
	Database bool `json:"database"`
	// flag for visualization / interactive content
	// obfuscating data
	Interactive bool `json:"interactive"`
	// flag for a page that links to many files
	ManyFiles bool `json:"manyFiles"`
	// uncrawlable comments
	Comments string `json:"comments"`
}

// Read uncrawlable from db
func (c *Uncrawlable) Read(db sqlQueryable) error {
	if c.Url != "" {
		row := db.QueryRow(qUncrawlableByUrl, c.Url)
		return c.UnmarshalSQL(row)
	}
	return ErrNotFound
}

// Save a uncrawlable
func (c *Uncrawlable) Save(db sqlQueryExecable) error {
	prev := &Uncrawlable{Url: c.Url}
	if err := prev.Read(db); err != nil {
		if err == ErrNotFound {
			c.Id = NewUuid()
			c.Created = time.Now().Round(time.Second)
			c.Updated = c.Created
			_, err := db.Exec(qUncrawlableInsert, c.SQLArgs()...)
			return err
		} else {
			return err
		}
	} else {
		c.Updated = time.Now().Round(time.Second)
		_, err := db.Exec(qUncrawlableUpdate, c.SQLArgs()...)
		return err
	}
	return nil
}

// Delete a uncrawlable, should only do for erronious additions
func (c *Uncrawlable) Delete(db sqlQueryExecable) error {
	_, err := db.Exec(qUncrawlableDelete, c.Url)
	return err
}

// UnmarshalSQL reads an sql response into the uncrawlable receiver
// it expects the request to have used uncrawlableCols() for selection
func (u *Uncrawlable) UnmarshalSQL(row sqlScannable) (err error) {
	var (
		created, updated                         time.Time
		id, creator, url, name, email, eventName string
		agency, agencyId, subagencyId            string
		orgId, suborgId, subprimerId, comments   string
		ftp, database, interactive, manyFiles    bool
	)

	if err := row.Scan(
		&id, &url, &created, &updated,
		&creator, &name, &email, &eventName, &agency,
		&agencyId, &subagencyId, &orgId, &suborgId, &subprimerId,
		&ftp, &database, &interactive, &manyFiles,
		&comments); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}

	*u = Uncrawlable{
		Id:          id,
		Created:     created.In(time.UTC),
		Updated:     updated.In(time.UTC),
		Creator:     creator,
		Url:         url,
		Name:        name,
		Email:       email,
		EventName:   eventName,
		Agency:      agency,
		AgencyId:    agencyId,
		SubagencyId: subagencyId,
		OrgId:       orgId,
		SuborgId:    suborgId,
		SubprimerId: subprimerId,
		Ftp:         ftp,
		Database:    database,
		Interactive: interactive,
		ManyFiles:   manyFiles,
		Comments:    comments,
	}

	return nil
}

// SQLArgs formats a uncrawlable struct for inserting / updating into postgres
func (u *Uncrawlable) SQLArgs() []interface{} {
	return []interface{}{
		u.Id,
		u.Url,
		u.Created.In(time.UTC),
		u.Updated.In(time.UTC),
		u.Creator,
		u.Name,
		u.Email,
		u.EventName,
		u.Agency,
		u.AgencyId,
		u.SubagencyId,
		u.OrgId,
		u.SuborgId,
		u.SubprimerId,
		u.Ftp,
		u.Database,
		u.Interactive,
		u.ManyFiles,
		u.Comments,
	}
}
