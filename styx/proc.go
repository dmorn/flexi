// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package styx

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"aqwari.net/net/styx"
	"github.com/jecoz/flexi"
)

func (p *Process) openPanic(name string) io.WriteCloser {
	file, err := p.rootfs.Lookup(name)
	if err != nil {
		panic("open buffer: " + err.Error())
	}
	wc, ok := file.(io.WriteCloser)
	if !ok {
		panic("open buffer: not a write/closer")
	}
	return wc
}

func (p *Process) openErr() io.WriteCloser  { return p.openPanic("err") }
func (p *Process) openRetv() io.WriteCloser { return p.openPanic("retv") }

func (p *Process) startProcessor(pf *Pluffer) error {
	if p.served {
		return fmt.Errorf("process already executed")
	}

	pflen := pf.Len()
	if pflen == 0 {
		// ctl pluffer was closed but no data has been
		// written to it.
		return nil
	}
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, pf)
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
		&Pluffer{
			Name:    "ctl",
			Perm:    0222,
			ModTime: time.Now(),
			OnClose: p.startProcessor,
		},
		&DemuxBuffer{
			Name:    "retv",
			Perm:    0444,
			ModTime: time.Now(),
		},
		&DemuxBuffer{
			Name:    "err",
			Perm:    0444,
			ModTime: time.Now(),
		},
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

type File interface {
	Stat() (os.FileInfo, error)
	Open() (interface{}, error)
	Truncate(int64) error
}

func (p *Process) lookup(path string) (File, error) {
	switch path {
	case "/":
		return p.rootfs, nil
	case "/ctl", "/status", "/retv", "/err":
		_, file := filepath.Split(path)
		return p.rootfs.Lookup(file)
	default:
		return nil, os.ErrNotExist
	}
}

func (p *Process) open(path string) (interface{}, error) {
	file, err := p.lookup(path)
	if err != nil {
		return nil, err
	}
	return file.Open()
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