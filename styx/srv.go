package styx

import (
	"log"
	"net"
	"os"

	"aqwari.net/net/styx"
	"github.com/jecoz/flexi"
)

type Srv struct {
	spw flexi.Spawner
	ln  net.Listener
}

func NewSrv(s flexi.Spawner) *Srv {
	return &Srv{spw: s}
}

func (s *Srv) serveReq(t styx.Request) error {
	return nil
}

func (s *Srv) Serve9P(sess *styx.Session) {
	for sess.Next() {
		t := sess.Request()
		if err := s.serveReq(t); err != nil {
			t.Rerror(err.Error())
		}
	}
}

func (s *Srv) Serve(port string) error {
	addr := net.JoinHostPort("", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.ln = ln

	srv := &styx.Server{
		Addr:     addr,
		Handler:  styx.Stack(logrequests, s),
		ErrorLog: log.New(os.Stderr, "", 0),
		TraceLog: log.New(os.Stderr, "", 0),
	}
	return srv.Serve(ln)
}
