package tarbuild

import "testing"

func Test_deepList(t *testing.T) {
	dir, err := NewDirFromOS("testdata")
	if err != nil {
		t.Fatal(err)
	}

	err = dir.ApplyIgnore(".tarignore")
	if err != nil {
		t.Fatal(err)
	}

	for _, n := range dir.DeepEntries {
		t.Log(n)
	}
}
