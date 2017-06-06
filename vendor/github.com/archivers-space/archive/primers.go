package archive

import (
	"database/sql"
	"github.com/archivers-space/sqlutil"
)

// CrawlingPrimers
// func CrawlingPrimers(db sqlutil.Queryable, limit, offset int) (primers []*Primer, err error) {
// 	rows, err := db.Query(qPrimersCrawling, limit, offset)
// 	if err != nil {
// 		return primers, err
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		d := &Primer{}
// 		if err := d.UnmarshalSQL(rows); err != nil {
// 			return primers, err
// 		}

// 		primers = append(primers, d)
// 	}

// 	return
// }

// CountPrimers returns the total number of primers
func CountPrimers(db sqlutil.Queryable) (count int64, err error) {
	err = db.QueryRow(qPrimersCount).Scan(&count)
	return
}

// ListPrimers
func ListPrimers(db sqlutil.Queryable, limit, offset int) (primers []*Primer, err error) {
	rows, err := db.Query(qPrimersList, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return UnmarshalBoundedPrimers(rows, limit)
}

// BasePrimers lists primers that have no parent
func BasePrimers(db sqlutil.Queryable, limit, offset int) (primers []*Primer, err error) {
	rows, err := db.Query(qBasePrimersList, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return UnmarshalBoundedPrimers(rows, limit)
}

// UnmarshalBoundedPrimers turns sql.Rows into primers, expecting len(rows) <= limit
func UnmarshalBoundedPrimers(rows *sql.Rows, limit int) (primers []*Primer, err error) {
	primers = make([]*Primer, limit)
	i := 0
	for rows.Next() {
		p := &Primer{}
		if err := p.UnmarshalSQL(rows); err != nil {
			return primers, err
		}

		primers[i] = p
		i++
	}
	return primers[:i], nil
}
