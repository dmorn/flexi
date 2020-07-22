// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package styx

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"time"

	"aqwari.net/net/styx"
	"github.com/jecoz/flexi"
)

func (p *Process) panicOpen(name string) io.WriteCloser {
	file, err := p.fs.Lookup(name)
	if err != nil {
		panic("open buffer: " + err.Error())
	}
	wc, ok := file.(io.WriteCloser)
	if !ok {
		panic("open buffer: not a write/closer")
	}
	return wc
}

func (p *Process) openErr() io.WriteCloser  { return p.panicOpen("err") }
func (p *Process) openRetv() io.WriteCloser { return p.panicOpen("retv") }

func (p *Process) startProcessor(rc io.ReadCloser) error {
	defer rc.Close()
	if p.served {
		return errors.New("process: served already")
	}

	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, rc)
	if err != nil {
		return err
	}

	p.served = true
	go func() {
		stdio := &flexi.Stdio{
			In:   buf,
			Err:  p.openErr(),
			Retv: p.openRetv(),
		}
		p.r.Run(stdio)
		stdio.Err.Close()
		stdio.Retv.Close()
	}()

	return nil
}

// ServeProcess creates a styx process serving 9p over ln. To gracefully shutdown
// the server, close the listener.
func ServeProcess(ln net.Listener, r flexi.Processor) error {
	p := &Process{r: r, ln: ln}

	files := []File{
		NewInputBuffer("ctl", p.startProcessor),
		NewOutputBuffer("retv"),
		NewOutputBuffer("err"),
	}
	root := &Dir{
		Name: "",
		Ls: func() []File {
			return files
		},
		Perm:    0555,
		ModTime: time.Now(),
	}
	p.fs = &fs{Root: root}

	return p.Serve()
}

type Process struct {
	served bool
	rootfs *Dir
	fs     *fs
	r      flexi.Processor
	ln     net.Listener
}

func (p *Process) Serve() error {
	echo := styx.HandlerFunc(func(s *styx.Session) {
		for s.Next() {
			log.Printf("%q %T %s", s.User, s.Request(), s.Request().Path())
		}
	})
	srv := &styx.Server{
		Handler:  styx.Stack(echo, p),
		ErrorLog: log.New(os.Stderr, "", log.LstdFlags),
	}
	return srv.Serve(p.ln)
}

func (p *Process) Serve9P(s *styx.Session) {
	for s.Next() {
		p.fs.serveRequest(s.Request())
	}
}
