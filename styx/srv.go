package styx

import (
	"log"
	"net"
	"os"

	"aqwari.net/net/styx"
	"github.com/jecoz/flexi"
)

var logrequests styx.HandlerFunc = func(s *styx.Session) {
	for s.Next() {
		log.Printf("%q %T %s", s.User, s.Request(), s.Request().Path())
	}
}

type Srv struct {
	Port string
	ln   net.Listener
}

func (s *Srv) Run(r flexi.Processor) error {
	p, err := NewProcess()
	if err != nil {
		return err
	}

	addr := net.JoinHostPort("", s.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.ln = ln

	go r(p)

	srv := &styx.Server{
		Addr:     addr,
		Handler:  styx.Stack(logrequests, p),
		ErrorLog: log.New(os.Stderr, "", 0),
		TraceLog: log.New(os.Stderr, "", 0),
	}
	return srv.Serve(ln)
}

func (s *Srv) Close() error {
	return s.ln.Close()
}
