package main

import (
	"testing"
)

func TestSubprimerStorage(t *testing.T) {
	defer resetTestData(appDB, "crawl_urls")

	c := &Subprimer{Url: "youtube.com", PrimerId: "5b1031f4-38a8-40b3-be91-c324bf686a87", Crawl: true}
	if err := c.Save(appDB); err != nil {
		t.Error(err.Error())
		return
	}

	c.Crawl = false
	if err := c.Save(appDB); err != nil {
		t.Error(err.Error())
		return
	}

	c2 := &Subprimer{Url: "youtube.com"}
	if err := c2.Read(appDB); err != nil {
		t.Error(err.Error())
		return
	}

	if c2.Crawl != c.Crawl {
		t.Errorf("crawl doesn't match: %t != %t", c2.Crawl, c.Crawl)
	}

	if !c2.Created.Equal(c.Created) {
		t.Errorf("created doesn't match: %s != %s", c2.Created.String(), c.Created.String())
	}

	if !c2.Updated.Equal(c.Updated) {
		t.Errorf("updated doesn't match: %s != %s", c2.Updated.String(), c.Updated.String())
	}

	if err := c.Delete(appDB); err != nil {
		t.Error(err.Error())
		return
	}
}
