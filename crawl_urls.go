package main

import (
	"fmt"
)

func CrawlingUrls(db sqlQueryable) ([]*CrawlUrl, error) {
	rows, err := db.Query(fmt.Sprintf("select %s from crawl_urls where crawl = true", crawlUrlCols()))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls := make([]*CrawlUrl, 0)
	for rows.Next() {
		c := &CrawlUrl{}
		if err := c.UnmarshalSQL(rows); err != nil {
			return nil, err
		}
		urls = append(urls, c)
	}
	return urls, nil
}
