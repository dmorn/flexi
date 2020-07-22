// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package styx

import (
	"errors"
	"io"
	"os"
	"time"

	"aqwari.net/net/styx"
)

type Opener interface {
	OpenDir() (styx.Directory, error)
	// Should be a styxfile.Interface implementation. The library takes
	// care of the conversion though.
	OpenFile() (io.ReadWriteCloser, error)
}

type File interface {
	Opener
	Stat() (os.FileInfo, error)
	Truncate(int64) error
}

type sizer interface {
	Size() int64
}

type finfo struct {
	o       Opener
	s       sizer
	name    string
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

var _ os.FileInfo = finfo{}

func (i finfo) Name() string       { return i.name }
func (i finfo) Size() int64        { return i.s.Size() }
func (i finfo) Mode() os.FileMode  { return i.mode }
func (i finfo) ModTime() time.Time { return i.modTime }
func (i finfo) IsDir() bool        { return i.isDir }
func (i finfo) Sys() (v interface{}) {
	var err error
	if i.IsDir() {
		v, err = i.o.OpenDir()
	} else {
		v, err = i.o.OpenFile()
	}
	if err != nil {
		panic(err)
	}
	return
}

type regularFile struct {
	path string
	info os.FileInfo
}

var _ File = &regularFile{}

func (f *regularFile) OpenDir() (styx.Directory, error) {
	return nil, errors.New("regular file: open dir: not a directory")
}

func (f *regularFile) OpenFile() (io.ReadWriteCloser, error) {
	return os.Open(f.path)
}

func (f *regularFile) Stat() (os.FileInfo, error) { return f.info, nil }
func (f *regularFile) Truncate(n int64) error {
	file, err := os.Open(f.path)
	if err != nil {
		return err
	}
	return file.Truncate(n)
}
