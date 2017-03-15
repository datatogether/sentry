package main

import (
	"testing"
)

func TestPrimerStorage(t *testing.T) {
	defer resetTestData(appDB, "primers")

	p := &Primer{Title: "Test Primer", Description: "test primer description!"}
	if err := p.Save(appDB); err != nil {
		t.Error(err.Error())
		return
	}

	p.Description = "new description"
	if err := p.Save(appDB); err != nil {
		t.Error(err.Error())
		return
	}

	p2 := &Primer{Id: p.Id}
	if err := p2.Read(appDB); err != nil {
		t.Error(err.Error())
		return
	}

	if !p2.Created.Equal(p.Created) {
		t.Errorf("created doesn't match: %s != %s", p2.Created.String(), p.Created.String())
	}

	if !p2.Updated.Equal(p.Updated) {
		t.Errorf("updated doesn't match: %s != %s", p2.Updated.String(), p.Updated.String())
	}

	if err := p.Delete(appDB); err != nil {
		t.Error(err.Error())
		return
	}
}
