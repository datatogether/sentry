package core

import (
	"database/sql"
	"github.com/datatogether/sqlutil"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
)

// ListSources lists all sources from most to least recent, paginated
func ListSources(store datastore.Datastore, limit, offset int) ([]*Source, error) {
	q := query.Query{
		Prefix: Source{}.DatastoreType(),
		Limit:  limit,
		Offset: offset,
	}

	res, err := store.Query(q)
	if err != nil {
		return nil, err
	}

	sources := make([]*Source, limit)
	i := 0
	for r := range res.Next() {
		if r.Error != nil {
			return nil, err
		}

		c, ok := r.Value.(*Source)
		if !ok {
			return nil, ErrInvalidResponse
		}

		sources[i] = c
		i++
	}

	return sources[:i], nil
	// return UnmarshalBoundedSources(rows, limit)
}

// CountSources grabs the total number of sources
func CountSources(db sqlutil.Queryable) (count int, err error) {
	err = db.QueryRow(qSourcesCount).Scan(&count)
	return
}

// CrawlingSources lists sources with crawling = true, paginated
func CrawlingSources(db sqlutil.Queryable, limit, offset int) ([]*Source, error) {
	rows, err := db.Query(qSourcesCrawling, limit, offset)
	if err != nil {
		return nil, err
	}
	return UnmarshalBoundedSources(rows, limit)
}

// UnmarshalBoundedSources turns a standard sql.Rows of Source results into a *Source slice
func UnmarshalBoundedSources(rows *sql.Rows, limit int) ([]*Source, error) {
	defer rows.Close()
	subsources := make([]*Source, limit)
	i := 0
	for rows.Next() {
		u := &Source{}
		if err := u.UnmarshalSQL(rows); err != nil {
			return nil, err
		}
		subsources[i] = u
		i++
	}

	return subsources[:i], nil
}
