package main

import "fmt"

func ListUrls(db sqlQueryable, limit, skip int) ([]*Url, error) {
	rows, err := appDB.Query(fmt.Sprintf("select %s from urls limit $1 offset $2", urlCols()), limit, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls := make([]*Url, limit)
	i := 0
	for rows.Next() {
		u := &Url{}
		if err := u.UnmarshalSQL(rows); err != nil {
			logger.Println(err)
			return nil, err
		}
		urls[i] = u
		i++
	}

	return urls[:i], nil
}

func UnfetchedDomainUrls(db sqlQueryable, host string) ([]*Url, error) {
	rows, err := appDB.Query(fmt.Sprintf("select %s from urls where host = $1 and last_get = 0 limit 50", urlCols()), host)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls := []*Url{}
	for rows.Next() {
		u := &Url{}
		if err := u.UnmarshalSQL(rows); err != nil {
			logger.Println(err)
			return nil, err
		}
		urls = append(urls, u)
	}
	return urls, nil
}
