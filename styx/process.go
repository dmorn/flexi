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
	file, err := p.lookup(name)
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

func (p *Process) startProcessor(ib *InputBuffer) error {
	if p.served {
		return errors.New("process: served already")
	}

	if ib.Size() == 0 {
		// ctl was closed but no data has been
		// written to it.
		return nil
	}
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, ib)
	if err != nil {
		return err
	}

	p.served = true
	go p.r.Run(&flexi.Stdio{
		In:   buf,
		Err:  p.openErr,
		Retv: p.openRetv,
	})

	return nil
}

func ServeProcess(ln net.Listener, r flexi.Processor) error {
	p := &Process{r: r, ln: ln}

	files := []File{
		NewInputBuffer("ctl", p.startProcessor),
		NewOutputBuffer("retv"),
		NewOutputBuffer("err"),
	}
	p.rootfs = &Dir{
		Name:    "/",
		Files:   files,
		Perm:    0555,
		ModTime: time.Now(),
	}

	return p.Serve()
}

type Process struct {
	served bool
	rootfs *Dir
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

func (p *Process) Close() error { return p.ln.Close() }

func (p *Process) lookup(path string) (File, error) {
	return p.rootfs.Lookup(path)
}

func (p *Process) open(path string) (interface{}, error) {
	file, err := p.lookup(path)
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

func (p *Process) stat(path string) (os.FileInfo, error) {
	file, err := p.lookup(path)
	if err != nil {
		return nil, err
	}
	return file.Stat()
}
func (p *Process) serveRequest(t styx.Request) {
	switch msg := t.(type) {
	case styx.Topen:
		msg.Ropen(p.open(msg.Path()))
	case styx.Twalk:
		msg.Rwalk(p.stat(msg.Path()))
	case styx.Tstat:
		msg.Rstat(p.stat(msg.Path()))
	case styx.Ttruncate:
		file, err := p.lookup(msg.Path())
		if err != nil {
			msg.Rtruncate(err)
			return
		}
		msg.Rtruncate(file.Truncate(msg.Size))
	default:
	}
}

func (p *Process) Serve9P(s *styx.Session) {
	for s.Next() {
		p.serveRequest(s.Request())
	}
}
