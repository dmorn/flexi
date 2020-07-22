// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package styx

import (
	"os"
	"strings"

	"aqwari.net/net/styx"
)

type fs struct {
	Root *Dir
}

func lookup(path []string, files []File) File {
	if len(files) == 0 {
		// We do not have anything more to check!
		return nil
	}
	head := files[0]
	tail := files[1:]
	target := ""
	if len(path) > 0 {
		target = strings.TrimSpace(path[0])
	}

	info, err := head.Stat()
	if err != nil {
		return nil
	}
	if info.Name() != target {
		return lookup(path, tail)
	}
	if len(path) <= 1 {
		return head
	}

	// Reaching this point means the file is relevant,
	// but we're not yet at the end of the path (on the
	// right way though).
	// What we have in our files list does not matter:
	// the file we're looking for is under this one, hence
	// we should search in the files contained in this
	// directory.
	dir, ok := head.(*Dir)
	if !ok {
		// We cannot do anything. The file is supposed to
		// be under this directory, but this is not a
		// directory.
		return nil
	}
	return lookup(path[1:], dir.Ls())
}

func (fs *fs) Lookup(path string) (File, error) {
	fields := strings.Split(path, "/")
	if path == "/" {
		fields = []string{}
	}

	file := lookup(fields, []File{fs.Root})
	if file == nil {
		return nil, os.ErrNotExist
	}
	return file, nil
}

func (fs *fs) Open(path string) (interface{}, error) {
	file, err := fs.Lookup(path)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return file.OpenDir()
	} else {
		return file.OpenFile()
	}
}

func (fs *fs) Stat(path string) (os.FileInfo, error) {
	file, err := fs.Lookup(path)
	if err != nil {
		return nil, err
	}
	return file.Stat()
}

func (fs *fs) serveRequest(t styx.Request) {
	switch msg := t.(type) {
	case styx.Topen:
		msg.Ropen(fs.Open(msg.Path()))
	case styx.Twalk:
		msg.Rwalk(fs.Stat(msg.Path()))
	case styx.Tstat:
		msg.Rstat(fs.Stat(msg.Path()))
	case styx.Ttruncate:
		file, err := fs.Lookup(msg.Path())
		if err != nil {
			msg.Rtruncate(err)
			return
		}
		msg.Rtruncate(file.Truncate(msg.Size))
	default:
	}
}
