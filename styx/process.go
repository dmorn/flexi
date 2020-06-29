package styx

import (
	"io"
	"log"
	"net"
	"os"

	"aqwari.net/net/styx"
	"github.com/jecoz/flexi"
)

type Process struct {
	ctl    *File
	status *File
	retv   *File
	err    *File
	ln     net.Listener
	port   string
}

func NewProcess() *Process {
	ctl := NewFile("ctl", 0222)
	status := NewFile("status", 0444)
	retv := NewFile("retv", 0444)
	err := NewFile("err", 0444)

	return &Process{
		ctl:    ctl,
		status: status,
		retv:   retv,
		err:    err,
	}
}

func (p *Process) Dir() *Dir {
	return NewDir(0555, p.ctl, p.status, p.retv, p.err)
}

func (p *Process) Ctl() io.Reader {
	return p.ctl
}

func (p *Process) Status() io.Writer {
	return p.status
}

func (p *Process) Retv() io.Writer {
	return p.retv
}

func (p *Process) Err() io.Writer {
	return p.err
}

func (p *Process) serveReq(t styx.Request) error {
	info, err := p.Dir().Stat(t.Path())
	if err != nil {
		return err
	}

	switch msg := t.(type) {
	case styx.Topen:
		msg.Ropen(info.Sys(), nil)
	case styx.Twalk:
		msg.Rwalk(info, nil)
	case styx.Tstat:
		msg.Rstat(info, nil)
	default:
	}
	return nil
}

func (p *Process) Serve9P(s *styx.Session) {
	for s.Next() {
		t := s.Request()
		if err := p.serveReq(t); err != nil {
			t.Rerror(err.Error())
		}
	}
}

var logrequests styx.HandlerFunc = func(s *styx.Session) {
	for s.Next() {
		log.Printf("%q %T %s", s.User, s.Request(), s.Request().Path())
	}
}

func (p *Process) Serve(port string, r flexi.Processor) error {
	addr := net.JoinHostPort("", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	p.ln = ln

	go r(p)

	srv := &styx.Server{
		Addr:     addr,
		Handler:  styx.Stack(logrequests, p),
		ErrorLog: log.New(os.Stderr, "", 0),
		TraceLog: log.New(os.Stderr, "", 0),
	}
	return srv.Serve(ln)
}

func (p *Process) Close() error {
	return p.ln.Close()
}
