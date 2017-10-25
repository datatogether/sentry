package core

import (
	"database/sql"
	"github.com/datatogether/sqlutil"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
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
func ListPrimers(store datastore.Datastore, limit, offset int) ([]*Primer, error) {
	q := query.Query{
		Prefix: Primer{}.DatastoreType(),
		Limit:  limit,
		Offset: offset,
	}

	res, err := store.Query(q)
	if err != nil {
		return nil, err
	}

	primers := make([]*Primer, limit)
	i := 0
	for r := range res.Next() {
		if r.Error != nil {
			return nil, err
		}

		c, ok := r.Value.(*Primer)
		if !ok {
			return nil, ErrInvalidResponse
		}

		primers[i] = c
		i++
	}

	return primers[:i], nil
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
