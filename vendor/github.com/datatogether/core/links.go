package core

import (
	"github.com/datatogether/sqlutil"
)

// ReadDstLinks returns all links that specify a given url as src
func ReadDstLinks(db sqlutil.Queryable, src *Url) ([]*Link, error) {
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
func ReadSrcLinks(db sqlutil.Queryable, dst *Url) ([]*Link, error) {
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

// ReadDstContentLinks returns a list of links that specify a gien url as src that are content urls
func ReadDstContentLinks(db sqlutil.Queryable, src *Url) ([]*Link, error) {
	res, err := db.Query(qUrlDstContentLinks, src.Url)
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
