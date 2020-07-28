package file

import (
	"io"
	"os"
)

type Regular struct {
	path string
}

func (r *Regular) Open() (io.ReadWriteCloser, error) { return os.Open(r.path) }
func (r *Regular) Stat() (os.FileInfo, error)        { return os.Stat(r.path) }

func NewRegular(path string) *Regular { return &Regular{path} }
