package archive

func Search(db sqlQueryable, q string, limit, offset int) ([]*Url, error) {
	if limit == 0 || limit > 50 {
		limit = 50
	}

	rows, err := db.Query(qUrlsSearch, "%"+q+"%", limit, offset)
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
