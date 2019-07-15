package db

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestFileDB(t *testing.T) {
	path, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)
	runGenericTests(t, &FileDB{Dir: path})
}
