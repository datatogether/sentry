package archive

import (
	"database/sql"
	"fmt"

	"time"
)

// A link represents an <a> tag in an html document src who's href
// attribute points to the url that resolves to dst.
// both src & dst must be stored as urls
type Link struct {
	// created timestamp rounded to seconds in UTC
	Created time.Time `json:"created"`
	// updated timestamp rounded to seconds in UTC
	Updated time.Time `json:"updated"`
	// origin url of the linking document
	Src *Url `json:"src"`
	// absolute url of the <a> href property
	Dst *Url `json:"dst"`
}

func (l *Link) Read(db sqlQueryable) error {
	var row *sql.Row
	if l.Src != nil && l.Dst != nil {
		row = db.QueryRow(fmt.Sprintf("select %s from links where src = $1 and dst= $2", linkCols()), l.Src.Url, l.Dst.Url)
	} else {
		return ErrNotFound
	}
	return l.UnmarshalSQL(row)
}

func (l *Link) Insert(db sqlQueryExecable) error {
	l.Created = time.Now().In(time.UTC).Round(time.Second)
	l.Updated = l.Created
	_, err := db.Exec(fmt.Sprintf("insert into links (%s) values ($1, $2, $3, $4)", linkCols()), l.SQLArgs()...)
	return err
}

func (l *Link) Update(db sqlQueryExecable) error {
	l.Updated = time.Now().Round(time.Second)
	_, err := db.Exec("update links set created = $1, updated = $2 where src = $3 and dst = $4", l.SQLArgs()...)
	return err
}

func (l *Link) Delete(db sqlQueryExecable) error {
	_, err := db.Exec("delete from links where src = $1 and dst = $2", l.Src.Url, l.Dst.Url)
	return err
}

func linkCols() string {
	return "created, updated, src, dst"
}

func (l *Link) SQLArgs() []interface{} {
	return []interface{}{
		l.Created.In(time.UTC),
		l.Updated.In(time.UTC),
		l.Src.Url,
		l.Dst.Url,
	}
}

func (l *Link) UnmarshalSQL(row sqlScannable) error {
	var (
		created, updated time.Time
		src, dst         string
	)

	if err := row.Scan(&created, &updated, &src, &dst); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}

	*l = Link{
		Created: created.In(time.UTC),
		Updated: updated.In(time.UTC),
		Src:     &Url{Url: src},
		Dst:     &Url{Url: dst},
	}

	return nil
}
