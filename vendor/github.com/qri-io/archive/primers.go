package archive

// CrawlingPrimers
// func CrawlingPrimers(db sqlQueryable, limit, offset int) (primers []*Primer, err error) {
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

// ListPrimers
func ListPrimers(db sqlQueryable, limit, offset int) (primers []*Primer, err error) {
	rows, err := db.Query(qPrimersList, limit, offset)
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
