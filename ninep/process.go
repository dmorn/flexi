package ninep

import (
	"bytes"
	"io"
	"log"

	"aqwari.net/net/styx"
)

type Process struct {
	ctl    bytes.Buffer
	status bytes.Buffer
	retv   bytes.Buffer
	err    bytes.Buffer
}

func (p *Process) Ctl() io.Reader {
	return &p.ctl
}

func (p *Process) Status() io.Writer {
	return &p.status
}

func (p *Process) Retv() io.Writer {
	return &p.retv
}

func (p *Process) Err() io.Writer {
	return &p.err
}

func (p *Process) Serve9P(s *styx.Session) {
	for s.Next() {
		req := s.Request()
		switch msg := req.(type) {
		case styx.Twalk:
			log.Printf("path: %v\n", msg.Path())
		default:
		}
	}
}
