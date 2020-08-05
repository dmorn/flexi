// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package file

import (
	"io"
	"os"
)

type Regular struct {
	path string
	info os.FileInfo
}

func (r *Regular) Open() (io.ReadWriteCloser, error) {
	return os.OpenFile(r.path, os.O_RDWR|os.O_CREATE, r.info.Mode())
}
func (r *Regular) Stat() (os.FileInfo, error) {
	i, err := os.Stat(r.path)
	if err != nil {
		return nil, err
	}
	r.info = i
	return i, nil
}
func (r *Regular) Close() error { return nil }

func NewRegular(path string, i os.FileInfo) *Regular {
	return &Regular{path: path, info: i}
}
