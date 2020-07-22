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

// DiskLs returns an Ls function that inspects path on disk. Basically
// it works just like Unix's ls command, but returns a list of File.
// It can be used to create Dir instances that act on the disk.
func DiskLs(path string) func() []File {
	return func() []File {
		dir, err := os.Open(path)
		if err != nil {
			return []File{}
		}
		defer dir.Close()

		// Even though Readdir might return an error, it will
		// return the FileInfos found till that point. That's
		// enough for our use-case.
		infos, _ := dir.Readdir(-1)
		files := make([]File, len(infos))
		for i, v := range infos {
			path := filepath.Join(path, v.Name())
			if v.IsDir() {
				files[i] = &Dir{
					Name:    v.Name(),
					Perm:    v.Mode(),
					ModTime: v.ModTime(),
					Ls:      DiskLs(path),
				}
			} else {
				files[i] = &regularFile{
					path: path,
					info: v,
				}
			}
		}
		return files
	}
}

// All public fields should be initialized before using the directory.
type Dir struct {
	Name    string
	Ls      func() []File
	Perm    os.FileMode
	ModTime time.Time
}

func (d *Dir) Stat() (os.FileInfo, error) {
	return finfo{
		name:    d.Name,
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

	for _, v := range d.Ls() {
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
	all := d.Dir.Ls()
	if d.offset >= len(all) {
		return nil, io.EOF
	}
	files := all[d.offset:]
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
