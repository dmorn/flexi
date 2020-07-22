package styx

import (
	"path/filepath"
	"testing"
)

func TestLookup(t *testing.T) {
	tt := []struct {
		path string
		dir  *Dir
	}{
		{
			path: "/",
			dir: &Dir{
				Ls: func() []File {
					return []File{}
				},
			},
		},
		{
			path: "/ctl",
			dir: &Dir{
				Ls: func() []File {
					ctl := NewInputBuffer("ctl", nil)
					return []File{ctl}
				},
			},
		},
	}

	for i, v := range tt {
		fs := &fs{Root: v.dir}
		file, err := fs.Lookup(v.path)
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
