package archive

import (
	"database/sql"
	"github.com/archivers-space/sqlutil"
)

func ListUncrawlables(db sqlutil.Queryable, limit, offset int) ([]*Uncrawlable, error) {
	rows, err := db.Query(qUncrawlablesList, limit, offset)
	if err != nil {
		return nil, err
	}
	return UnmarshalBoundedUncrawlables(rows, limit)
}

func UnmarshalBoundedUncrawlables(rows *sql.Rows, limit int) ([]*Uncrawlable, error) {
	defer rows.Close()
	subprimers := make([]*Uncrawlable, limit)
	i := 0
	for rows.Next() {
		u := &Uncrawlable{}
		if err := u.UnmarshalSQL(rows); err != nil {
			return nil, err
		}
		subprimers[i] = u
		i++
	}

	return subprimers[:i], nil
}
