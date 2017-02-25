package main

import (
	"net/url"
	"testing"
)

func TestUrlContextStorage(t *testing.T) {
	// urls, err := before(appDB)
	// if err != nil {
	// 	t.Errorf(err.Error())
	// 	return
	// }
	// defer after(appDB, urls)
	_u, _ := url.Parse("http://www.epa.gov")

	c := &UrlContext{
		Url:           _u,
		ContributorId: "bal",
		Context: map[string]interface{}{
			"test": "context",
		},
	}

	if err := c.Save(appDB); err != nil {
		t.Error(err.Error())
		return
	}

	c.Context = map[string]interface{}{
		"test": "updated context",
	}

	if err := c.Save(appDB); err != nil {
		t.Error(err.Error())
		return
	}

	c2 := &UrlContext{Url: _u, ContributorId: "bal"}
	if err := c2.Read(appDB); err != nil {
		t.Error(err.Error())
		return
	}

	if c2.Context["test"].(string) != "updated context" {
		t.Errorf("context didn't save")
		return
	}

	if err := c.Delete(appDB); err != nil {
		t.Error(err.Error())
		return
	}
}
