// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package file

import (
	"os"
	"time"
)

type Info struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (i Info) Name() string       { return i.name }
func (i Info) Size() int64        { return i.size }
func (i Info) Mode() os.FileMode  { return i.mode }
func (i Info) ModTime() time.Time { return i.modTime }
func (i Info) IsDir() bool        { return i.isDir }
func (i Info) Sys() interface{}   { return nil }
