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
	ctl, status, retv, err *VolFile
	root                   *Dir
	ln                     net.Listener
	port                   string
}

func NewProcess() *Process {
	ctl := &VolFile{Name: "ctl", Perm: 0222}
	status := &VolFile{Name: "status", Perm: 0444}
	retv := &VolFile{Name: "retv", Perm: 0444}
	err := &VolFile{Name: "err", Perm: 0444}
	root := &Dir{
		Perm:  0555,
		Name:  "/",
		Files: []File{ctl, status, retv, err},
	}

	return &Process{
		ctl:    ctl,
		status: status,
		retv:   retv,
		err:    err,
		root:   root,
	}
}

func (p *Process) Ctl() io.Reader    { return p.ctl }
func (p *Process) Status() io.Writer { return p.status }
func (p *Process) Retv() io.Writer   { return p.retv }
func (p *Process) Err() io.Writer    { return p.err }

func (p *Process) lookup(path string) (File, error) {
	switch path {
	case "/":
		return p.root, nil
	case "/ctl":
		return p.ctl, nil
	case "/status":
		return p.status, nil
	case "/retv":
		return p.retv, nil
	case "/err":
		return p.err, nil
	default:
		return nil, os.ErrNotExist
	}
}

func (p *Process) open(path string) (interface{}, error) {
	file, err := p.lookup(path)
	if err != nil {
		return nil, err
	}
	return file.Sys(), nil
}

func (p *Process) stat(path string) (os.FileInfo, error) {
	file, err := p.lookup(path)
	if err != nil {
		return nil, err
	}
	return file.Stat()
}

func (p *Process) Serve9P(s *styx.Session) {
	for s.Next() {
		switch msg := s.Request().(type) {
		case styx.Topen:
			msg.Ropen(p.open(msg.Path()))
		case styx.Twalk:
			msg.Rwalk(p.stat(msg.Path()))
		case styx.Tstat:
			msg.Rstat(p.stat(msg.Path()))
		default:
		}
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

	echo := styx.HandlerFunc(func(s *styx.Session) {
		for s.Next() {
			log.Printf("%q %T %s", s.User, s.Request(), s.Request().Path())
		}
	})

	srv := &styx.Server{
		Addr:     addr,
		Handler:  styx.Stack(echo, p),
		ErrorLog: log.New(os.Stderr, "", log.LstdFlags),
	}
	return srv.Serve(ln)
}

func (p *Process) Close() error {
	return p.ln.Close()
}
