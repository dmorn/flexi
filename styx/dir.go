// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package styx

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// All public fields should be initialized before using the directory.
type Dir struct {
	Name    string
	Files   []File
	Perm    os.FileMode
	ModTime time.Time
}

func (d *Dir) Stat() (os.FileInfo, error) {
	return &DirInfo{Dir: d}, nil
}

func (d *Dir) Open() (interface{}, error) { return &DirReader{Dir: d}, nil }
func (d *Dir) Truncate(size int64) error  { return fmt.Errorf("not supported") }

func (d *Dir) Lookup(name string) (File, error) {
	for _, v := range d.Files {
		info, err := v.Stat()
		if err != nil {
			continue
		}
		if info.Name() == name {
			return v, nil
		}
	}
	return nil, os.ErrNotExist
}

type DirInfo struct {
	*Dir
}

func (d DirInfo) Name() string       { return filepath.Dir(d.Dir.Name) }
func (d DirInfo) Size() int64        { return 0 }
func (d DirInfo) Mode() os.FileMode  { return d.Dir.Perm | os.ModeDir }
func (d DirInfo) ModTime() time.Time { return d.Dir.ModTime }
func (d DirInfo) IsDir() bool        { return true }
func (d DirInfo) Sys() interface{} {
	reader, _ := d.Dir.Open()
	return reader
}

type DirReader struct {
	*Dir
	offset int
}

func (d *DirReader) Readdir(n int) ([]os.FileInfo, error) {
	if d.Dir == nil {
		return nil, os.ErrInvalid
	}
	if d.offset >= len(d.Dir.Files) {
		return nil, io.EOF
	}
	files := d.Dir.Files[d.offset:]
	count := len(files)
	take := n
	if take <= 0 || take > count {
		take = count
	}

	files = files[:take]
	fis := make([]os.FileInfo, len(files))
	for i, v := range files {
		info, err := v.Stat()
		if err != nil {
			return nil, err
		}
		fis[i] = info
	}
	d.offset += take
	return fis, nil
}
