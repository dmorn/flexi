// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package fs

import (
	"io"
	"os"
)

// File describes the minimal list of functions
// required to interact with a file in flexi's
// context.
// Some files might be Directory instances too.
type File interface {
	// Open File for i/o operations.
	Open() (io.ReadWriteCloser, error)
	// Stat returns file information. Use
	// it to determine wether the File is
	// actually a Directory or not.
	Stat() (os.FileInfo, error)
}

type Directory interface {
	Readdir(n int) ([]os.FileInfo, error)
}

// RoFS is a read-only file-system.
type RoFS interface {
	Open(path string) (File, error)
}

type FS interface {
	RoFS
	Create(path string, newfile File) error
}
