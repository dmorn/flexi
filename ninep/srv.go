package ninep

import (
	"net"

	"aqwari.net/net/styx"
	"github.com/jecoz/flexi"
)

type Srv struct {
	Port string
	ln   net.Listener
}

func (s *Srv) Run(r flexi.Processor) error {
	p := new(Process)
	addr := net.JoinHostPort("", s.Port)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.ln = ln

	go r(p)

	srv := &styx.Server{
		Addr:    addr,
		Handler: p,
	}
	return srv.Serve(ln)
}

func (s *Srv) Close() error {
	return s.ln.Close()
}
