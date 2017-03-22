package archive

import (
	"fmt"
)

func ListCollections(db sqlQueryable, limit, skip int) ([]*Collection, error) {
	rows, err := db.Query(fmt.Sprintf("SELECT %s FROM collections ORDER BY created DESC LIMIT $1 OFFSET $2", collectionCols()), limit, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	collections := make([]*Collection, limit)
	i := 0
	for rows.Next() {
		u := &Collection{}
		if err := u.UnmarshalSQL(rows); err != nil {
			return nil, err
		}
		collections[i] = u
		i++
	}

	return collections[:i], nil
}
