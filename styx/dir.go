package styx

import (
	"io"
	"os"
	"path/filepath"
	"time"
)

type Dir struct {
	name   string
	perm   os.FileMode
	t      time.Time
	offset int
	files  []os.FileInfo
}

func NewDir(perm os.FileMode, fis ...os.FileInfo) *Dir {
	return &Dir{
		name:  "/",
		perm:  perm,
		t:     time.Now(),
		files: fis,
	}
}

func (d *Dir) Stat(p string) (os.FileInfo, error) {
	if p == "" || p == "/" {
		return d, nil
	}
	dir, file := filepath.Split(p)
	if dir != "/" {
		return nil, os.ErrNotExist
	}
	for _, v := range d.files {
		if v.Name() == file {
			return v, nil
		}
	}

	return nil, os.ErrNotExist
}

func (d *Dir) Name() string {
	return d.name
}

func (d *Dir) Size() int64 {
	return 0
}

func (d *Dir) Mode() os.FileMode {
	return d.perm | os.ModeDir
}

func (d *Dir) ModTime() time.Time {
	return d.t
}

func (d *Dir) IsDir() bool {
	return true
}

func (d *Dir) Sys() interface{} {
	return d
}

// https://github.com/golang/go/blob/fc821667dd73987d1e579a813b50e403f8ff3c22/src/os/dir.go#L22
func (d *Dir) Readdir(n int) (files []os.FileInfo, err error) {
	if d == nil {
		err = os.ErrInvalid
		return
	}
	if d.offset >= len(d.files) {
		err = io.EOF
		return
	}
	old := d.files[d.offset:]
	len := len(old)
	take := n
	if take <= 0 || take > len {
		take = len
		err = io.EOF
	}
	d.offset += take
	files = old[:take]
	return
}
