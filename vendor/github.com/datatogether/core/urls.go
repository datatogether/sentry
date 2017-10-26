package core

import (
	"database/sql"
	"fmt"
	"github.com/datatogether/sqlutil"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
)

func ListUrls(store datastore.Datastore, limit, offset int) ([]*Url, error) {
	q := query.Query{
		Prefix: Url{}.DatastoreType(),
		Limit:  limit,
		Offset: offset,
	}

	res, err := store.Query(q)
	if err != nil {
		return nil, err
	}

	urls := make([]*Url, limit)
	i := 0
	for r := range res.Next() {
		if r.Error != nil {
			return nil, err
		}

		c, ok := r.Value.(*Url)
		if !ok {
			return nil, ErrInvalidResponse
		}

		urls[i] = c
		i++
	}

	return urls[:i], nil
	// rows, err := db.Query(qUrlsList, limit, skip)
	// if err != nil {
	// 	return nil, err
	// }
	// defer rows.Close()
	// return UnmarshalBoundedUrls(rows, limit)
}

func ContentUrls(db sqlutil.Queryable, limit, skip int) ([]*Url, error) {
	rows, err := db.Query(qContentUrlsList, limit, skip)
	if err != nil {
		return nil, err
	}
	return UnmarshalBoundedUrls(rows, limit)
}

func ContentUrlsCount(db sqlutil.Queryable) (count int, err error) {
	err = db.QueryRow(qContentUrlsCount).Scan(&count)
	return
}

func FetchedUrls(db sqlutil.Queryable, limit, offset int) ([]*Url, error) {
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

func UnfetchedUrls(db sqlutil.Queryable, limit, offset int) ([]*Url, error) {
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

func UrlsForHash(db sqlutil.Queryable, hash string) ([]*Url, error) {
	rows, err := db.Query(qUrlsForHash, hash)
	if err != nil {
		return nil, err
	}
	return UnmarshalUrls(rows)
}

func ValidArchivingUrl(db sqlutil.Queryable, url string) error {
	var exists bool
	err := db.QueryRow("select exists(select 1 from suburls where $1 ilike concat('%', url ,'%'))", url).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("Oops! Only urls contained in suburls can be archived. cannot archive %s", url)
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
