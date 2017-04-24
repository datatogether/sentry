package archive

import (
	"database/sql"
	"fmt"
)

func ContentUrls(db sqlQueryable, limit, skip int) ([]*Url, error) {
	rows, err := db.Query(qContentUrlsList, limit, skip)
	if err != nil {
		return nil, err
	}
	return UnmarshalBoundedUrls(rows, limit)
}

func ContentUrlsCount(db sqlQueryable) (count int, err error) {
	err = db.QueryRow(qContentUrlsCount).Scan(&count)
	return
}

func ListUrls(db sqlQueryable, limit, skip int) ([]*Url, error) {
	rows, err := db.Query(qUrlsList, limit, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return UnmarshalBoundedUrls(rows, limit)
}

func FetchedUrls(db sqlQueryable, limit, offset int) ([]*Url, error) {
	if limit == 0 {
		limit = 100
	}
	rows, err := db.Query(qUrlsFetched, limit, offset)
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

func UnfetchedUrls(db sqlQueryable, limit, offset int) ([]*Url, error) {
	if limit == 0 {
		limit = 50
	}
	rows, err := db.Query(qUrlsUnfetched, limit, offset)
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
	rows, err := db.Query(qUrlsForHash, hash)
	if err != nil {
		return nil, err
	}
	return UnmarshalUrls(rows)
}

func ValidArchivingUrl(db sqlQueryable, url string) error {
	var exists bool
	err := db.QueryRow("select exists(select 1 from subprimers where $1 ilike concat('%', url ,'%'))", url).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("Oops! Only urls contained in subprimers can be archived. cannot archive %s", url)
	}
	return nil
}

func UnmarshalBoundedUrls(rows *sql.Rows, limit int) ([]*Url, error) {
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
