package main

import (
	"fmt"
)

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
