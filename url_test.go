package main

import (
	"testing"
)

func TestUrlStorage(t *testing.T) {
	defer resetTestData(appDB, "urls", "links")

	u := &Url{Url: "http://youtube.com"}
	if err := u.Insert(appDB); err != nil {
		t.Error(err.Error())
		return
	}

	u.ContentLength = 10
	if err := u.Update(appDB); err != nil {
		t.Error(err.Error())
		return
	}

	u2 := &Url{Url: "http://youtube.com"}
	if err := u2.Read(appDB); err != nil {
		t.Error(err.Error())
		return
	}

	if !u2.Created.Equal(u.Created) {
		t.Errorf("created doesn't match: %s != %s", u2.Created.String(), u.Created.String())
	}

	if !u2.Updated.Equal(u.Updated) {
		t.Errorf("updated doesn't match: %s != %s", u2.Updated.String(), u.Updated.String())
	}

	if err := u.Delete(appDB); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestUrlGet(t *testing.T) {

}
