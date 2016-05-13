package tarbuild

import (
	"io/ioutil"
	"testing"
)

func TestBuild(t *testing.T) {
	err := Build(ioutil.Discard, "testdata", "testdata/Tarfile")
	if err != nil {
		t.Fatal(err)
	}
}
