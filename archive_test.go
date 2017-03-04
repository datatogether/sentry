package main

import (
	"testing"
)

func TestArchive(t *testing.T) {
	defer resetTestData(appDB, "urls", "links", "snapshots")
	var (
		links []*Link
		res   *Url
		err   error
	)
	close := make(chan bool)

	done := func(err error) {
		if err != nil {
			t.Error(err.Error())
			close <- true
			return
		}

		for _, l := range links {
			dst := l.Dst
			f, err := dst.File()
			if err != nil {
				t.Error(err.Error())
				close <- true
				return
			}

			if err := f.GetS3(); err != nil {
				t.Error(err.Error())
				close <- true
				return
			}

			if err := f.Delete(); err != nil {
				t.Error(err.Error())
				close <- true
			}
		}

		f, err := res.File()
		if err != nil {
			t.Error(err.Error())
			close <- true
			return
		}

		if err := f.GetS3(); err != nil {
			t.Error(err.Error())
			close <- true
			return
		}

		if err := f.Delete(); err != nil {
			t.Error(err.Error())
			close <- true
			return
		}

		close <- true
	}

	res, links, err = ArchiveUrl(appDB, "http://apple.com", done)
	if err != nil {
		t.Error(err.Error())
		return
	}
	<-close
}
