package memfs

import (
	"path/filepath"
	"testing"

	"github.com/jecoz/flexi/file"
	"github.com/jecoz/flexi/fs"
)

func TestOpen(t *testing.T) {
	tt := []struct {
		path string
		dir  *file.Dir
	}{
		{
			path: "/",
			dir:  file.NewDirFiles(""),
		},
		{
			path: "/retv",
			dir:  file.NewDirFiles("", file.NewMulti("retv")),
		},
	}

	for i, v := range tt {
		fs := New(v.dir)
		file, err := fs.Open(v.path)
		if err != nil {
			t.Fatalf("%d: %v", i, err)
		}
		info, err := file.Stat()
		if err != nil {
			t.Fatalf("%d: %v", i, err)
		}
		_, filename := filepath.Split(v.path)
		if info.Name() != filename {
			t.Fatalf("%d: unexpected filename returned: wanted [%v], found [%v]", i, filename, info.Name())
		}
	}

}
