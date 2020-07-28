// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package flexi

import (
	"bytes"
	"io"
	"net"

	"github.com/jecoz/flexi/file"
	"github.com/jecoz/flexi/file/memfs"
	"github.com/jecoz/flexi/fs"
	"github.com/jecoz/flexi/styx"
)

type Stdio struct {
	// In contains the input bytes. For 9p processes,
	// its the data written to the ctl file, which triggers
	// the execution of the process.
	In io.Reader
	// Write here execution errors.
	Err io.WriteCloser
	// Write here final output.
	Retv io.WriteCloser
}

// Processor describes an entity that is capable of executing
// a task reading and writing from Stdio.
type Processor interface {
	// Stdio Err and Retv buffers should not be used
	// after Run returns.
	Run(*Stdio)
}

// ProcessorFunc is an helper Processor implementation that
// allows to make Processor even ordinary functions.
type ProcessorFunc func(*Stdio)

func (f ProcessorFunc) Run(i *Stdio) { f(i) }

type Process struct {
	FS     fs.FS
	Ln     net.Listener
	Runner Processor
}

func (p *Process) Serve() error {
	return styx.Serve(p.Ln, p.FS)
}

func ServeProcess(ln net.Listener, r Processor) error {
	err := file.NewMulti("err")
	retv := file.NewMulti("retv")
	ctl := file.NewPlumber("ctl", func(p *file.Plumber) bool {
		if p.Size() == 0 {
			return false
		}
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, p); err != nil {
			return false
		}

		go func() {
			stdio := &Stdio{
				In:   buf,
				Err:  err,
				Retv: retv,
			}
			r.Run(stdio)
			err.Close()
			retv.Close()
		}()
		return true
	})

	root := file.NewDirFiles("", ctl, err, retv)
	p := Process{
		FS:     memfs.New(root),
		Ln:     ln,
		Runner: r,
	}
	return p.Serve()
}
