// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package process

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"aqwari.net/net/styx"
	"github.com/jecoz/flexi/fs"
	"github.com/jecoz/flexi/fs/styxfs"
	"github.com/jecoz/flexi/fs/synthfs"
)

type Stdio struct {
	// In contains the input bytes. For 9p processes,
	// its the data written to the in file, which triggers
	// the execution of the process.
	In io.Reader
	// Write here execution errors.
	Err io.WriteCloser
	// Write here final output.
	Retv io.WriteCloser
	// Write here status updates.
	State io.WriteCloser
}

// Runner describes an entity that is capable of executing
// a task reading and writing from Stdio.
type Runner interface {
	// Stdio Err and Retv buffers should not be used
	// after Run returns.
	Run(*Stdio)
}

// RunnerFunc is an helper Runner implementation that
// allows to make Runner even ordinary functions.
type RunnerFunc func(*Stdio)

func (f RunnerFunc) Run(i *Stdio) { f(i) }

type Srv struct {
	*styxfs.FS
}

func (s *Srv) Serve(ln net.Listener) error {
	srv := &styx.Server{
		// Remember that it is possible to stack handlers
		// using the styx.Stack helper.
		Handler:  s,
		TraceLog: log.New(os.Stderr, "", log.Ltime),
	}
	return srv.Serve(ln)
}

func NewSrv(fsys fs.RWFS) *Srv { return &Srv{styxfs.New(fsys)} }

func insertBuffer(fsys *synthfs.FS, n string, m os.FileMode, in ...string) *synthfs.Buffer {
	b := &synthfs.Buffer{
		Name: n,
		Mode: m,
	}
	if err := fsys.InsertOpener(b, n, in...); err != nil {
		panic(err)
	}
	return b
}

func panicOpenWriteCloser(of func() (fs.File, error)) io.WriteCloser {
	f, err := of()
	if err != nil {
		panic(err)
	}
	bf, ok := f.(*synthfs.BufferFile)
	if !ok {
		panic(fmt.Errorf("of() did not return a synthfs.BufferFile"))
	}
	return bf
}

func Serve(ln net.Listener, r Runner) error {
	fsys := new(synthfs.FS)
	state := insertBuffer(fsys, "state", 0440)
	retv := insertBuffer(fsys, "retv", 0440)
	errf := insertBuffer(fsys, "err", 0440)

	in := newHackBuffer("in", 0220, func(b *synthfs.Buffer) bool {
		// TODO: if buffer contains some content, create the Stdin
		// struct and start the runner. Remember that this function
		// prevents the hackBufferFile to Close.
		// TODO: decide how to react to errors.
		bf, err := b.Open()
		if err != nil {
			panic(err)
		}
		defer bf.Close()

		var bb bytes.Buffer
		n, err := io.Copy(&bb, bf)
		if err != nil {
			panic(err)
		}
		if n == 0 {
			return false
		}
		go func() {
			ewc := panicOpenWriteCloser(errf.Open)
			swc := panicOpenWriteCloser(state.Open)
			rwc := panicOpenWriteCloser(retv.Open)
			r.Run(&Stdio{
				In:    &bb,
				Err:   ewc,
				State: swc,
				Retv:  rwc,
			})
			ewc.Close()
			swc.Close()
			rwc.Close()
		}()
		return true
	})
	if err := fsys.InsertOpener(in, "in"); err != nil {
		panic(err)
	}

	log.Printf("*** listening on %v", ln.Addr())
	return NewSrv(fsys).Serve(ln)
}

type bufferCallback func(*synthfs.Buffer) bool

type hackBufferFile struct {
	*synthfs.BufferFile

	b       *synthfs.Buffer
	plumbed bool
	onClose bufferCallback
}

func (f *hackBufferFile) Close() error {
	// When a BufferFile is closed, it syncs its updated contents
	// with the Buffer itself. First we close the BufferFile and,
	// if no error occurs, we call the onClose callback if present.
	if err := f.BufferFile.Close(); err != nil {
		return err
	}
	if f.onClose != nil && !f.plumbed {
		// N.B. onClose might block.
		f.plumbed = f.onClose(f.b)
	}
	return nil
}

type hackBuffer struct {
	*synthfs.Buffer
	onClose bufferCallback
}

func (b *hackBuffer) Open() (fs.File, error) {
	f, err := b.Buffer.Open()
	if err != nil {
		return nil, err
	}
	bf, ok := f.(*synthfs.BufferFile)
	if !ok {
		return nil, fmt.Errorf("Open() did not return a synthfs.BufferFile")
	}

	return &hackBufferFile{
		BufferFile: bf,
		b:          b.Buffer,
		onClose:    b.onClose,
	}, nil
}

func newHackBuffer(n string, m os.FileMode, onClose bufferCallback) *hackBuffer {
	return &hackBuffer{
		Buffer: &synthfs.Buffer{
			Name: n,
			Mode: m,
		},
		onClose: onClose,
	}
}
