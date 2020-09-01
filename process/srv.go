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
	"github.com/jecoz/flexi/styx/styxfs"
	"github.com/jecoz/flexi/synth"
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

func insertBuffer(fsys *synth.FS, n string, m os.FileMode, in ...string) *synth.Buffer {
	b := &synth.Buffer{
		Name: n,
		Mode: m,
	}
	if err := fsys.InsertOpener(b, in...); err != nil {
		panic(err)
	}
	return b
}

func panicOpenWriteCloser(of func() (fs.File, error)) io.WriteCloser {
	f, err := of()
	if err != nil {
		panic(err)
	}
	bf, ok := f.(*synth.BufferFile)
	if !ok {
		panic(fmt.Errorf("of() did not return a synth.BufferFile"))
	}
	return bf
}

func Serve(ln net.Listener, r Runner) error {
	fsys := new(synth.FS)
	state := insertBuffer(fsys, "state", 0440)
	retv := insertBuffer(fsys, "retv", 0440)
	errf := insertBuffer(fsys, "err", 0440)

	inb := &synth.Buffer{Name: "in", Mode: 0220}
	plumbed := false
	in := synth.HackClose(inb, func() error {
		if plumbed {
			return nil
		}
		bf, err := inb.Open()
		if err != nil {
			return err
		}
		defer bf.Close()

		var bb bytes.Buffer
		n, err := io.Copy(&bb, bf)
		if err != nil {
			return err
		}
		if n == 0 {
			return nil
		}
		plumbed = true
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
		return nil
	})
	if err := fsys.InsertOpener(in); err != nil {
		panic(err)
	}

	log.Printf("*** listening on %v", ln.Addr())
	return NewSrv(fsys).Serve(ln)
}
