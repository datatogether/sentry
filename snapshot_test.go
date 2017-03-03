package main

import (
	"testing"
	"time"
)

func TestSnapshotStorge(t *testing.T) {
	defer resetTestData(appDB, "snapshots")

	now := time.Now()
	u := &Url{
		Url:          "http://www.epa.gov",
		LastGet:      &now,
		Status:       200,
		DownloadTook: 20,
		Headers:      []string{"test", "header"},
		Hash:         "thisshouldbeahash",
	}

	if err := WriteSnapshot(appDB, u); err != nil {
		t.Error(err.Error())
		return
	}
}
