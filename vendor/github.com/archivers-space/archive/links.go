package archive

// ReadDstLinks returns all links that specify a given url as src
func ReadDstLinks(db sqlQueryable, src *Url) ([]*Link, error) {
	res, err := db.Query(qUrlDstLinks, src.Url)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	links := make([]*Link, 0)
	for res.Next() {
		dst := &Url{}
		if err := dst.UnmarshalSQL(res); err != nil {
			return nil, err
		}
		l := &Link{
			Src: src,
			Dst: dst,
		}
		links = append(links, l)
	}

	return links, nil
}

// ReadSrcLinks returns all links that specify a given url as dst
func ReadSrcLinks(db sqlQueryable, dst *Url) ([]*Link, error) {
	res, err := db.Query(qUrlSrcLinks, dst.Url)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	links := make([]*Link, 0)
	for res.Next() {
		src := &Url{}
		if err := src.UnmarshalSQL(res); err != nil {
			return nil, err
		}
		l := &Link{
			Src: src,
			Dst: dst,
		}
		links = append(links, l)
	}

	return links, nil
}
