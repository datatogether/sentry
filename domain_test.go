package main

import (
	"testing"
)

func TestDomainStorage(t *testing.T) {
	defer resetTestData(appDB, "domains")

	d := &Domain{Host: "youtube.com", Crawl: true}
	if err := d.Insert(appDB); err != nil {
		t.Error(err.Error())
		return
	}

	d.Crawl = false
	if err := d.Update(appDB); err != nil {
		t.Error(err.Error())
		return
	}

	d2 := &Domain{Host: "youtube.com"}
	if err := d2.Read(appDB); err != nil {
		t.Error(err.Error())
		return
	}

	if !d2.Created.Equal(d.Created) {
		t.Errorf("created doesn't match: %s != %s", d2.Created.String(), d.Created.String())
	}

	if !d2.Updated.Equal(d.Updated) {
		t.Errorf("updated doesn't match: %s != %s", d2.Updated.String(), d.Updated.String())
	}

	if err := d.Delete(appDB); err != nil {
		t.Error(err.Error())
		return
	}
}
