package archive

import (
	"database/sql"
	"fmt"
)

func CrawlingUrls(db sqlQueryable) ([]*Subprimer, error) {
	rows, err := db.Query(fmt.Sprintf("select %s from subprimers where crawl = true", subprimerCols()))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls := make([]*Subprimer, 0)
	for rows.Next() {
		c := &Subprimer{}
		if err := c.UnmarshalSQL(rows); err != nil {
			return nil, err
		}
		urls = append(urls, c)
	}
	return urls, nil
}

func ListUrls(db sqlQueryable, limit, skip int) ([]*Url, error) {
	rows, err := db.Query(fmt.Sprintf("select %s from urls order by created desc limit $1 offset $2", urlCols()), limit, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls := make([]*Url, limit)
	i := 0
	for rows.Next() {
		u := &Url{}
		if err := u.UnmarshalSQL(rows); err != nil {
			return nil, err
		}
		urls[i] = u
		i++
	}

	return urls[:i], nil
}

func FetchedUrls(db sqlQueryable, limit, offset int) ([]*Url, error) {
	if limit == 0 {
		limit = 100
	}
	rows, err := db.Query(fmt.Sprintf("select %s from urls where last_get is not null order by created desc limit $1 offset $2", urlCols()), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls := []*Url{}
	for rows.Next() {
		u := &Url{}
		if err := u.UnmarshalSQL(rows); err != nil {
			return nil, err
		}
		urls = append(urls, u)
	}
	return urls, nil
}

func UnfetchedUrls(db sqlQueryable, limit int) ([]*Url, error) {
	if limit == 0 {
		limit = 50
	}
	rows, err := db.Query(fmt.Sprintf("select %s from urls where last_get is null order by created desc limit $1", urlCols()), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls := []*Url{}
	for rows.Next() {
		u := &Url{}
		if err := u.UnmarshalSQL(rows); err != nil {
			return nil, err
		}
		urls = append(urls, u)
	}
	return urls, nil
}

func UrlsForHash(db sqlQueryable, hash string) ([]*Url, error) {
	rows, err := db.Query(fmt.Sprintf("select %s from urls where hash = $1", urlCols()), hash)
	if err != nil {
		return nil, err
	}
	return UnmarshalUrls(rows)
}

// UnmarshalUrls takes an sql cursor & returns a slice of url pointers
// expects columns to math urlCols()
func UnmarshalUrls(rows *sql.Rows) ([]*Url, error) {
	defer rows.Close()

	urls := []*Url{}
	for rows.Next() {
		u := &Url{}
		if err := u.UnmarshalSQL(rows); err != nil {
			return nil, err
		}
		urls = append(urls, u)
	}
	return urls, nil
}
