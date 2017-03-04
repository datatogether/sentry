package main

import (
	"testing"
)

func TestFilename(t *testing.T) {
	expect := "341af0ae3fa5d603fc3d9a772cce67bdb42dbe2b0aa2bd81a4a1546d799d80c3"
	f := &File{
		Data: []byte("and the cow jumped over the moon"),
	}

	fn, err := f.Filename()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	if fn != expect {
		t.Errorf("mismatch. expected: %s, got: %s", expect, fn)
	}
}
