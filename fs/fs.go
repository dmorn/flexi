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
	// Close is used to ask a File to release
	// associated resources. In case of a
	// Remote, it might mean unmounting and
	// killing the remote process.
	Close() error
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
	// Create adds newfile at path within FS.
	Create(path string, newfile File) error
	// Remove removes a previously added file
	// from FS. First the file should be opened,
	// file.Close() should be called and eventually
	// the file should not be present in FS anymore.
	Remove(path string) error
}
