package main

import (
	"fmt"
)

// CrawlingDomains
func CrawlingDomains(db sqlQueryable, limit, offset int) (domains []*Domain, err error) {
	rows, err := db.Query(fmt.Sprintf("select %s from domains where crawl = true limit $1 offset $2", domainCols()), limit, offset)
	if err != nil {
		return domains, err
	}
	defer rows.Close()

	for rows.Next() {
		d := &Domain{}
		if err := d.UnmarshalSQL(rows); err != nil {
			return domains, err
		}

		domains = append(domains, d)
	}

	return
}

// ListDomains
func ListDomains(db sqlQueryable, limit, offset int) (domains []*Domain, err error) {
	rows, err := db.Query(fmt.Sprintf("select %s from domains", domainCols()))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		d := &Domain{}
		if err := d.UnmarshalSQL(rows); err != nil {
			return domains, err
		}

		domains = append(domains, d)
	}

	return
}
