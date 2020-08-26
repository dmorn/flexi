// Package fs provides a filesystem abstraction. Code has been
// carried from the (maybe) upcoming io/fs package that may land
// in the go standard library soon. Once it is accepted, this
// package may be replaced with the official one (or at least
// part of it).
package fs

import (
	"errors"
	"io"
	"os"
	"time"
)

var (
	ErrNotExist = os.ErrNotExist
	ErrExist = os.ErrExist
)

type File interface {
	io.Reader
	Stat() (os.FileInfo, error)
	Close() error
}

type FS interface {
	Open(path string) (File, error)
}

type RWFS interface {
	FS
	Create(path string, mode os.FileMode) (File, error)
	Remove(path string) error
}

type FileStat struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (i FileStat) Name() string       { return i.name }
func (i FileStat) Size() int64        { return i.size }
func (i FileStat) Mode() os.FileMode  { return i.mode }
func (i FileStat) ModTime() time.Time { return i.modTime }
func (i FileStat) IsDir() bool        { return i.isDir }
func (i FileStat) Sys() interface{}   { return nil }

type IsDirFile interface {
	File
	IsDir() bool
}

// FileIsDir is an helper function that returns true if it is certain that
// f is a directory. If false is returned, f may still be a directory. If
// callers need to ensure to get the correct answer, f.Stat() should be
// inspected instead (also callers could read and interpret the contents
// of the file themselves).
func FileIsDir(f File) bool {
	if idf, ok := f.(IsDirFile); ok {
		return idf.IsDir()
	}

	s, err := f.Stat()
	if err != nil {
		return false
	}
	return s.IsDir()
}

type DirFile struct {
	ModTime time.Time
	Files []os.FileInfo
	Name string
	off int
}

func (d *DirFile) Read(p []byte) (int, error) {
	return 0, errors.New("directory: Read is not supported, use Readdir instead")
}

func (d *DirFile) Stat() (os.FileInfo, error) {
	return FileStat{
		name: d.Name,
		size: 0,
		mode: os.ModeDir,
		modTime: d.ModTime,
		isDir: true,
	}, nil
}

func (d *DirFile) Close() error { return nil }

func (d *DirFile) Readdir(n int) (fi []os.FileInfo, err error) {
	if n < 0 || n > len(d.Files) {
		n = len(d.Files)
	}
	var i int
	for i = d.off; i < n; i++ {
		fi = append(fi, d.Files[i])
	}
	if i == n {
		err = io.EOF
	}
	d.off = i
	return
}
