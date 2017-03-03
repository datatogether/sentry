package main

import (
	"testing"
)

func TestLinkStorage(t *testing.T) {
	defer resetTestData(appDB, "links")

	l := &Link{Src: &Url{Url: "http://www.epa.gov"}, Dst: &Url{Url: "http://www.epa.gov"}}
	if err := l.Insert(appDB); err != nil {
		t.Error(err.Error())
		return
	}

	if err := l.Update(appDB); err != nil {
		t.Error(err.Error())
		return
	}

	l2 := &Link{
		Src: &Url{Url: "http://www.epa.gov"},
		Dst: &Url{Url: "http://www.epa.gov"},
	}
	if err := l2.Read(appDB); err != nil {
		t.Error(err.Error())
		return
	}

	if !l2.Created.Equal(l.Created) {
		t.Errorf("created doesn't match: %s != %s", l2.Created.String(), l.Created.String())
	}

	if !l2.Updated.Equal(l.Updated) {
		t.Errorf("updated doesn't match: %s != %s", l2.Updated.String(), l.Updated.String())
	}

	if err := l.Delete(appDB); err != nil {
		t.Error(err.Error())
		return
	}

}
