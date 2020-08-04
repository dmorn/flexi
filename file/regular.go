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
}

func (r *Regular) Open() (io.ReadWriteCloser, error) { return os.Create(r.path) }
func (r *Regular) Stat() (os.FileInfo, error)        { return os.Stat(r.path) }
func (r *Regular) Close() error                      { return nil }

func NewRegular(path string) *Regular { return &Regular{path} }
