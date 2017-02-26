package main

import (
	"testing"
)

func TestUrlContextStorage(t *testing.T) {

	c := &UrlContext{
		Url:           "http://www.epa.gov",
		ContributorId: "bal",
		Metadata: map[string]interface{}{
			"test": "context",
		},
	}

	if err := c.Save(appDB); err != nil {
		t.Error(err.Error())
		return
	}

	c.Metadata = map[string]interface{}{
		"test": "updated context",
	}

	if err := c.Save(appDB); err != nil {
		t.Error(err.Error())
		return
	}

	c2 := &UrlContext{Url: "http://www.epa.gov", ContributorId: "bal"}
	if err := c2.Read(appDB); err != nil {
		t.Error(err.Error())
		return
	}

	if c2.Metadata["test"].(string) != "updated context" {
		t.Errorf("context didn't save")
		return
	}

	if err := c.Delete(appDB); err != nil {
		t.Error(err.Error())
		return
	}
}
