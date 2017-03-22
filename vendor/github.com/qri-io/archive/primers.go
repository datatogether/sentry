package archive

import (
	"fmt"
)

// CrawlingPrimers
func CrawlingPrimers(db sqlQueryable, limit, offset int) (primers []*Primer, err error) {
	rows, err := db.Query(fmt.Sprintf("select %s from primers where crawl = true limit $1 offset $2", primerCols()), limit, offset)
	if err != nil {
		return primers, err
	}
	defer rows.Close()

	for rows.Next() {
		d := &Primer{}
		if err := d.UnmarshalSQL(rows); err != nil {
			return primers, err
		}

		primers = append(primers, d)
	}

	return
}

// ListPrimers
func ListPrimers(db sqlQueryable, limit, offset int) (primers []*Primer, err error) {
	rows, err := db.Query(fmt.Sprintf("select %s from primers", primerCols()))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		d := &Primer{}
		if err := d.UnmarshalSQL(rows); err != nil {
			return primers, err
		}

		primers = append(primers, d)
	}

	return
}
