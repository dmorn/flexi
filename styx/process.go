package styx

import (
	"io"
	"log"
	"net"
	"os"
	"fmt"

	"aqwari.net/net/styx"
	"github.com/jecoz/flexi"
)

type Process struct {
	ctl    *MemFile
	status *MemFile
	retv   *MemFile
	err    *MemFile
	fs     *Fs
	ln     net.Listener
	port   string
}

func NewProcess() *Process {
	ctl := NewMemFile("/ctl", 0222)
	status := NewMemFile("/status", 0444)
	retv := NewMemFile("/retv", 0444)
	err := NewMemFile("/err", 0444)
	dir := NewDir("/", 0555, ctl, status, retv, err)

	fs := NewFs()
	for _, v := range []*MemFile{ctl, status, retv, err} {
		if err := fs.Add(v.path, v); err != nil {
			panic(fmt.Sprintf("new process: %v: %v", v.path, err))
		}
	}
	if err := fs.Add(dir.path, dir); err != nil {
		panic(err)
	}

	return &Process{
		ctl:    ctl,
		status: status,
		retv:   retv,
		err:    err,
		fs:     fs,
	}
}

func (p *Process) Ctl() io.Reader    { return p.ctl }
func (p *Process) Status() io.Writer { return p.status }
func (p *Process) Retv() io.Writer   { return p.retv }
func (p *Process) Err() io.Writer    { return p.err }

func (p *Process) Serve9P(s *styx.Session) {
	for s.Next() {
		switch msg := s.Request().(type) {
		case styx.Topen:
			msg.Ropen(p.fs.Open(msg.Path()))
		case styx.Twalk:
			msg.Rwalk(p.fs.Stat(msg.Path()))
		case styx.Tstat:
			msg.Rstat(p.fs.Stat(msg.Path()))
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
