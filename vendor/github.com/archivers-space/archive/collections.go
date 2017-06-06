package archive

import (
	"github.com/archivers-space/sqlutil"
)

func ListCollections(db sqlutil.Queryable, limit, skip int) ([]*Collection, error) {
	rows, err := db.Query(qCollections, limit, skip)
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
