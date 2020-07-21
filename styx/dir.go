// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package styx

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aqwari.net/net/styx"
)

// All public fields should be initialized before using the directory.
type Dir struct {
	Name    string
	Files   []File
	Perm    os.FileMode
	ModTime time.Time
}

func (d *Dir) Stat() (os.FileInfo, error) {
	return finfo{
		name:    filepath.Dir(d.Name),
		mode:    d.Perm | os.ModeDir,
		modTime: d.ModTime,
		isDir:   true,
		o:       d,
		s:       d,
	}, nil
}

func (d *Dir) Truncate(size int64) error { return errors.New("dir: truncate not supported") }
func (d *Dir) Size() int64               { return 0 }

func (d *Dir) OpenDir() (styx.Directory, error) {
	return &dirReader{Dir: d}, nil
}

func (d *Dir) OpenFile() (io.ReadWriteCloser, error) {
	return nil, errors.New("dir: open file: not a regular file")
}
func (d *Dir) Lookup(name string) (File, error) {
	if name == d.Name {
		return d, nil
	}
	filename := strings.TrimPrefix(name, d.Name)

	for _, v := range d.Files {
		info, err := v.Stat()
		if err != nil {
			continue
		}
		if info.Name() == filename {
			return v, nil
		}
	}
	return nil, os.ErrNotExist
}

type dirReader struct {
	*Dir
	offset int
}

func (d *dirReader) Readdir(n int) ([]os.FileInfo, error) {
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
