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

func TestShouldEnqueue(t *testing.T) {
	defer resetTestData(appDB, "urls")

	cases := []struct {
		url       string
		get, head bool
	}{
		// TODO - this test isn't working the func properly. Should enhance with DB interaction
		{"https://youtube.com", true, true},
		{"http://www.epa.gov", true, true},
		{"http://epa.gov/new", true, true},
	}

	for _, c := range cases {
		u := &Url{Url: "http://youtube.com"}
		head := u.ShouldEnqueueGet()
		if head != c.head {
			t.Errorf("shouldEnqueueHead: %s error. expected %t, got %t", c.url, c.head, head)
		}

		get := u.ShouldEnqueueGet()
		if get != c.get {
			t.Errorf("shouldEnqueueGet: %s expected %t, got %t", c.url, c.get, get)
		}
	}
}

func TestUrlGet(t *testing.T) {

}
