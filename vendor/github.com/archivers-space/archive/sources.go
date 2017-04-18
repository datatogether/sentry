package archive

import (
	"database/sql"
)

// ListSources lists all sources from most to least recent, paginated
func ListSources(db sqlQueryable, limit, offset int) ([]*Source, error) {
	rows, err := db.Query(qSourcesList, limit, offset)
	if err != nil {
		return nil, err
	}
	return UnmarshalBoundedSources(rows, limit)
}

// CrawlingSources lists sources with crawling = true, paginated
func CrawlingSources(db sqlQueryable, limit, offset int) ([]*Source, error) {
	rows, err := db.Query(qSourcesCrawling, limit, offset)
	if err != nil {
		return nil, err
	}
	return UnmarshalBoundedSources(rows, limit)
}

// UnmarshalBoundedSources turns a standard sql.Rows of Source results into a *Source slice
func UnmarshalBoundedSources(rows *sql.Rows, limit int) ([]*Source, error) {
	defer rows.Close()
	subprimers := make([]*Source, limit)
	i := 0
	for rows.Next() {
		u := &Source{}
		if err := u.UnmarshalSQL(rows); err != nil {
			return nil, err
		}
		subprimers[i] = u
		i++
	}

	return subprimers[:i], nil
}
