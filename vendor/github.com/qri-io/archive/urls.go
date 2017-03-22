package archive

import "fmt"

func CrawlingUrls(db sqlQueryable) ([]*Subprimer, error) {
	rows, err := db.Query(fmt.Sprintf("select %s from subprimers where crawl = true", subprimerCols()))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls := make([]*Subprimer, 0)
	for rows.Next() {
		c := &Subprimer{}
		if err := c.UnmarshalSQL(rows); err != nil {
			return nil, err
		}
		urls = append(urls, c)
	}
	return urls, nil
}

func ListUrls(db sqlQueryable, limit, skip int) ([]*Url, error) {
	rows, err := db.Query(fmt.Sprintf("select %s from urls limit $1 offset $2", urlCols()), limit, skip)
	if err != nil {
		return nil, err
	}
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

func UnfetchedUrls(db sqlQueryable, limit int) ([]*Url, error) {
	if limit == 0 {
		limit = 50
	}
	rows, err := db.Query(fmt.Sprintf("select %s from urls where last_get is null limit $1", urlCols()), limit)
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
