package archive

import (
	"fmt"
)

func Search(db sqlQueryable, q string, limit, offset int) ([]*Url, error) {
	if limit == 0 || limit > 50 {
		limit = 50
	}

	rows, err := db.Query(fmt.Sprintf("select %s from urls where url ilike $1 limit $2 offset $3", urlCols()), "%"+q+"%", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]*Url, 0)
	for rows.Next() {
		u := &Url{}

		if err := u.UnmarshalSQL(rows); err != nil {
			return nil, err
		}
		results = append(results, u)
	}

	return results, nil
}
