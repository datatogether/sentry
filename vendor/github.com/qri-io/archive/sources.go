package archive

import (
	"database/sql"
)

func ListSources(db sqlQueryable, limit, offset int) ([]*Source, error) {
	rows, err := db.Query(qSourcesList, limit, offset)
	if err != nil {
		return nil, err
	}
	return UnmarshalBoundedSources(rows, limit)
}

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
