package main

import (
	"database/sql"
	"fmt"
	"net/url"

	"time"
)

type Link struct {
	Created time.Time
	Updated time.Time
	Src     *Url
	Dst     *Url
}

func (l *Link) Read(db sqlQueryable) error {
	var row *sql.Row
	if l.Src != nil && l.Dst != nil {
		row = db.QueryRow(fmt.Sprintf("select %s from links where src = $1 and dst= $2", linkCols()), l.Src.Url.String(), l.Dst.Url.String())
	} else {
		return ErrNotFound
	}
	return l.UnmarshalSQL(row)
}

func (l *Link) Insert(db sqlQueryExecable) error {
	l.Created = time.Now()
	l.Updated = l.Created
	_, err := db.Exec(fmt.Sprintf("insert into links (%s) values ($1, $2, $3, $4)", linkCols()), l.SQLArgs()...)
	return err
}

func (l *Link) Update(db sqlQueryExecable) error {
	l.Updated = time.Now()
	_, err := db.Exec(fmt.Sprintf("update link set created = $1, updated = $2 where src = $3 and dst = $4", linkCols()), l.SQLArgs()...)
	return err
}

func linkCols() string {
	return "created, updated, src, dst"
}

func (l *Link) SQLArgs() []interface{} {
	return []interface{}{
		l.Created.Unix(),
		l.Updated.Unix(),
		l.Src.Url.String(),
		l.Dst.Url.String(),
	}
}

func (l *Link) UnmarshalSQL(row sqlScannable) error {
	var (
		created, updated int64
		src, dst         string
	)

	if err := row.Scan(&created, &updated, &src, &dst); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}

	srcUrl, err := url.Parse(src)
	if err != nil {
		return err
	}

	dstUrl, err := url.Parse(dst)
	if err != nil {
		return err
	}

	*l = Link{
		Created: time.Unix(created, 0),
		Updated: time.Unix(updated, 0),
		Src:     &Url{Url: srcUrl},
		Dst:     &Url{Url: dstUrl},
	}

	return nil
}
